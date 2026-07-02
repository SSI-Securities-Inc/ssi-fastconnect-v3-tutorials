// Sample 12 — MA Cross Signal Real-time (WebSocket bars)
// Ket hop WebSocket de nhan tick real-time, tu aggregate thanh nen,
// tinh MA5/MA10, dat lenh khi giao cat va theo doi lenh qua stream.

using SsiSdk;
using SsiSdk.Models;

static class Sample12MaCrossAutoTrade
{
    private const string Symbol = "SSI";
    private const int MaFast = 5;
    private const int MaSlow = 10;
    private const int Quantity = 100;
    private const int BarSeconds = 300; // Nen 5 phut

    private static readonly HashSet<string> TerminalStatuses = new()
    {
        OrderStatus.Filled,
        OrderStatus.Cancelled,
        OrderStatus.Rejected,
        OrderStatus.Expired,
        OrderStatus.PartialCancelled,
    };

    // --- Bar builder ---
    private sealed class Bar
    {
        public long Ts { get; set; }
        public double Open { get; set; }
        public double High { get; set; }
        public double Low { get; set; }
        public double Close { get; set; }
        public int Volume { get; set; }
    }

    private sealed class BarBuilder
    {
        private readonly int _interval;
        private readonly object _lock = new();
        private Bar? _current;
        private readonly List<Bar> _closed = new();

        public BarBuilder(int intervalSeconds) => _interval = intervalSeconds;

        public void Seed(IEnumerable<SsiSdk.Models.OhlcData> historicalBars)
        {
            lock (_lock)
            {
                foreach (var b in historicalBars)
                {
                    _closed.Add(new Bar
                    {
                        Open = b.OpenPrice, High = b.HighPrice,
                        Low = b.LowPrice, Close = b.ClosePrice, Volume = b.Volume,
                    });
                    if (_closed.Count > 200) _closed.RemoveAt(0);
                }
            }
            Console.WriteLine($"  Seeded {_closed.Count} historical bars");
        }

        public Bar? OnTrade(double price, int quantity)
        {
            var bucket = DateTimeOffset.UtcNow.ToUnixTimeSeconds() / _interval * _interval;
            lock (_lock)
            {
                if (_current is null || _current.Ts != bucket)
                {
                    var closed = _current;
                    _current = new Bar { Ts = bucket, Open = price, High = price, Low = price, Close = price, Volume = quantity };
                    if (closed is not null)
                    {
                        _closed.Add(closed);
                        if (_closed.Count > 200) _closed.RemoveAt(0);
                        return closed;
                    }
                    return null;
                }
                _current.High = Math.Max(_current.High, price);
                _current.Low = Math.Min(_current.Low, price);
                _current.Close = price;
                _current.Volume += quantity;
                return null;
            }
        }

        public List<Bar> Snapshot()
        {
            lock (_lock)
            {
                var result = new List<Bar>(_closed);
                if (_current is not null) result.Add(_current);
                return result;
            }
        }
    }

    // --- MA / Signal helpers ---

    private static double? CalculateMa(List<Bar> bars, int period)
    {
        if (bars.Count < period) return null;
        return bars.Skip(bars.Count - period).Sum(b => b.Close) / period;
    }

    private static string? DetectCross(List<Bar> bars, int fast, int slow)
    {
        if (bars.Count < slow + 1) return null;
        var mfNow = CalculateMa(bars, fast);
        var msNow = CalculateMa(bars, slow);
        var prev = bars.GetRange(0, bars.Count - 1);
        var mfPrev = CalculateMa(prev, fast);
        var msPrev = CalculateMa(prev, slow);
        if (mfNow is null || msNow is null || mfPrev is null || msPrev is null) return null;
        if (mfPrev <= msPrev && mfNow > msNow) return "BUY";
        if (mfPrev >= msPrev && mfNow < msNow) return "SELL";
        return null;
    }

    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        // ===== Buoc 1: Seed lich su 5m =====
        var builder = new BarBuilder(BarSeconds);
        var dataClient = new DataClient(auth);

        Console.WriteLine($"--- Load lich su OHLC 5m ({Symbol}) ---");
        var hist = await dataClient.MarketData.GetOhlc5MinuteHistoricalAsync(
            Symbol, "2026/01/01 00:00:00", "2026/05/12 23:59:59", 1, MaSlow + 5);
        builder.Seed(hist);

        // ===== State dung chung giua callbacks =====
        string? activeOrderId = null;
        string? lastSignal = null;
        var stateLock = new object();

        var trading = new TradingClient(auth);

        // ===== Buoc 2: Callbacks =====
        using var stream = new StreamClient(auth);

        stream.Streaming.SetOnData(msg =>
        {
            if (msg is not TradeMessage trade) return;

            var closedBar = builder.OnTrade(trade.Price, trade.Quantity);
            if (closedBar is null) return;

            var bars = builder.Snapshot();
            var mf = CalculateMa(bars, MaFast);
            var ms = CalculateMa(bars, MaSlow);
            if (mf is not null && ms is not null)
                Console.WriteLine($"  [BAR] close={closedBar.Close:N0} vol={closedBar.Volume:N0} | MA{MaFast}={mf:F2} | MA{MaSlow}={ms:F2}");
            else
                Console.WriteLine($"  [BAR] close={closedBar.Close:N0} | Chua du du lieu MA");

            var signal = DetectCross(bars, MaFast, MaSlow);
            if (signal is null) return;

            lock (stateLock)
            {
                if (activeOrderId is not null)
                {
                    Console.WriteLine($"  [SIGNAL {signal}] Dang co lenh cho, bo qua.");
                    return;
                }
                if (lastSignal == signal) return;
                lastSignal = signal;
            }

            Console.WriteLine($"\n  *** SIGNAL {signal} {Symbol} ***");
            var side = signal == "BUY" ? OrderSide.Buy : OrderSide.Sell;

            try
            {
                var maxBs = trading.Trading.GetMaxBuySellAtMarketPriceAsync(SampleConfig.AccountNo, Symbol).GetAwaiter().GetResult();
                var maxQty = signal == "BUY" ? maxBs.MaxBuyQuantity : maxBs.MaxSellQuantity;
                if (maxQty < Quantity)
                {
                    Console.WriteLine($"  [RISK] Khong du {Quantity} (co {maxQty}). Bo qua.");
                    return;
                }

                var result = trading.Trading.PlaceMarketOrderAsync(
                    SampleConfig.AccountNo, Symbol, side, Quantity).GetAwaiter().GetResult();
                var orderId = result.OrderId ?? "pending";
                Console.WriteLine($"  [ORDER] Dat lenh thanh cong: orderId={orderId}");
                lock (stateLock) { activeOrderId = orderId; }
            }
            catch (Exception ex)
            {
                Console.WriteLine($"  [ERROR] {ex.Message}");
                lock (stateLock) { lastSignal = null; }
            }
        });

        stream.Streaming.SetOnTrading(msg =>
        {
            if (msg is not OrderStatusMessage order) return;
            Console.WriteLine(
                $"  [ORDER UPDATE] {order.Symbol} {order.Side} | " +
                $"ID={order.OrderId} | Status={order.Status} | " +
                $"Khop={order.FilledQuantity}/{order.Quantity}");

            if (!TerminalStatuses.Contains(order.Status)) return;
            lock (stateLock) { activeOrderId = null; }

            if (order.Status == OrderStatus.Filled && order.FilledQuantity > 0)
            {
                var cost = order.FilledQuantity * order.Price;
                Console.WriteLine(
                    $"  [FILLED] {lastSignal} {order.Symbol}: " +
                    $"{order.FilledQuantity} CP @ {order.Price:N0} | " +
                    $"Tong: {cost:N0} VND");
            }
            else if (order.Status is OrderStatus.Cancelled or OrderStatus.Rejected)
            {
                Console.WriteLine($"  [CLOSED] Lenh ket thuc voi trang thai {order.Status}");
            }
        });

        stream.Streaming.SetOnHeartbeat(msg =>
        {
            Console.WriteLine($"  [HEARTBEAT] {msg.Status}");
        });

        // ===== Buoc 3: Ket noi WebSocket =====
        Console.WriteLine("\n--- Ket noi WebSocket ---");
        await stream.ConnectAsync();
        Console.WriteLine("Da ket noi!\n");

        await stream.Streaming.SubscribeSymbolAsync([Symbol]);
        await stream.Streaming.SubscribeOrderStatusAsync(SampleConfig.AccountNo);
        await stream.Streaming.PingAsync();

        Console.WriteLine($"Dang lang nghe nen {BarSeconds}s cho {Symbol}... (Ctrl+C de dung)\n");

        Console.CancelKeyPress += (_, e) =>
        {
            e.Cancel = true;
            Console.WriteLine("\nDung chien luoc.");
            stream.Disconnect();
        };

        await stream.WaitAsync();
    }
}
