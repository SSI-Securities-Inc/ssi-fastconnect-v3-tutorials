// Sample 16 — Streaming nâng cao
//   1. Trang thai ket noi (OnConnection): connected / reconnected / disconnected / failed
//   2. Handler theo loai (OnMarketFlag, OnIndexSummary, OnOrder, OnOrderMatch) thay vi switch tren OnData
//   3. Topic: nen 1m/5m, tick chi so, fan-out ro chi so VN30 (@*), ping dinh ky
//   4. ListSubscriptionAsync / UnsubscribeAllAsync (thoi gian theo mili-giay)
//   5. Quy tac topic: moi topic mot ma, khong dung danh sach dau phay
//
// Chi DOC (khong dat lenh). Chay khoang 60 giay roi tu ngat.

using SsiSdk;
using SsiSdk.Models;

public static class Sample16StreamingAdvanced
{
    private const int ListenMs = 60_000;

    public static async Task RunAsync()
    {
        SampleConfig.EnsureConfigured();
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        // --- Buoc 1: Quy tac topic (kiem tra tai client, chua gui gi) ---
        Console.WriteLine("--- Buoc 1: Quy tac topic ---");
        Console.WriteLine($"  Topic hop le : {Topic.Build("trade", "VNM", "1m")} | {Topic.Build("quote", "VN30", Topic.StreamIntervalAll)}");
        foreach (var (type, filter, interval) in new[]
        {
            ("trade", "VNM,FPT", (string?)null),   // danh sach dau phay: server khong tach
            ("trade", "VNM", "1w"),                // interval chi co tick/1m/5m
            ("quote", "VN30", "1m"),               // interval chi hop le cho trade
            ("trade", "*", Topic.StreamIntervalAll), // @* can mot ma cu the (chi so hoac san)
        })
        {
            try
            {
                Topic.Build(type, filter, interval);
                Console.WriteLine($"  (khong mong doi) {type}.{filter}@{interval} duoc chap nhan");
            }
            catch (ValidationException ex)
            {
                Console.WriteLine($"  Tu choi {type}.{filter}@{interval}: {ex.Message}");
            }
        }

        // --- Buoc 2: Callback ---
        using var stream = new StreamClient(auth);
        var counts = new System.Collections.Concurrent.ConcurrentDictionary<string, int>();
        void Count(string key, Action? first = null)
        {
            if (counts.AddOrUpdate(key, 1, (_, n) => n + 1) == 1) first?.Invoke();
        }

        // Trang thai ket noi: SDK tu reconnect, callback nay chi de theo doi
        stream.Streaming.OnConnection = e =>
            Console.WriteLine($"  [CONN] {e.State} (reconnect #{e.ReconnectCount})" +
                              (e.Error is null ? "" : $" loi: {e.Error.GetType().Name}"));

        stream.Streaming.OnMarketFlag = m => Count("marketFlag", () =>
            Console.WriteLine($"  [FLAG] {m.Board} {m.Flag} {m.TradingTime}"));
        stream.Streaming.OnIndexSummary = m => Count("indexSummary", () =>
            Console.WriteLine($"  [INDEX-SUMMARY] {m.Index} ({m.Data.Count} truong)"));
        stream.Streaming.OnOrder = o => Console.WriteLine(
            $"  [ORDER] {o.Symbol} {o.Side} | Status: {o.Status} | Khop: {o.FilledQuantity}/{o.Quantity}");
        stream.Streaming.OnOrderMatch = m => Console.WriteLine(
            $"  [MATCH] {m.Symbol} {m.Side} {m.MatchQty} @ {m.MatchPrice:N0}");

        // OnData van nhan MOI thong diep data (da parse) — dung cho cac loai khong co handler rieng
        stream.Streaming.OnData = msg =>
        {
            switch (msg)
            {
                case IntervalMessage bar:
                    Count("ohlcv-" + bar.IntervalTime, () => Console.WriteLine(
                        $"  [OHLCV] {bar.Symbol} {bar.IntervalTime} O:{bar.Open:N0} H:{bar.High:N0} L:{bar.Low:N0} C:{bar.Close:N0} V:{bar.Volume:N0}"));
                    break;
                case IndexTickMessage idx:
                    Count("indexTick", () => Console.WriteLine($"  [INDEX] {idx.Symbol} {idx.Value} ({idx.ChangePercent}%)"));
                    break;
                case TradeMessage t:
                    Count("trade", () => Console.WriteLine($"  [TRADE] {t.Symbol} {t.Price:N0} x{t.Quantity}"));
                    break;
                case QuoteMessage q:
                    Count("quote");
                    break;
                case ForeignRoomMessage:
                    Count("room");
                    break;
                default:
                    Count("other");
                    break;
            }
        };

        Console.WriteLine("\n--- Buoc 2: Ket noi ---");
        await stream.ConnectAsync();

        // --- Buoc 3: Subscribe ---
        Console.WriteLine("\n--- Buoc 3: Subscribe ---");
        // Nen 1 phut (interval hop le: tick, 1m, 5m)
        await stream.Streaming.SubscribeSymbolOhlcvAsync(["SSI"], "1m");
        // Tick cua chinh chi so
        await stream.Streaming.SubscribeIndexTradeAsync(["VNINDEX"]);
        // Du lieu tung ma trong ro VN30: gui trade/quote/room.VN30@* (nhieu thong diep, can than khi chay lau)
        await stream.Streaming.SubscribeIndexAsync(["VN30"]);
        // Co phien (ATO/LO/ATC...). OnMarketFlag chi chon noi nhan, can subscribe rieng
        await stream.Streaming.SubscribeMarketFlagAsync();
        // Lenh cua tai khoan: OnOrder / OnOrderMatch
        await stream.Streaming.SubscribeOrderStatusAsync(SampleConfig.AccountNo);

        // Ping dinh ky de giu ket noi (mili-giay); phan hoi nam o OnHeartbeat
        stream.Streaming.OnHeartbeat = h => Count("heartbeat");
        await stream.Streaming.PingAsync(intervalMs: 30_000);

        Console.CancelKeyPress += (_, e) =>
        {
            e.Cancel = true;
            stream.Disconnect();
        };

        // --- Buoc 4: Lang nghe ---
        Console.WriteLine($"\n--- Buoc 4: Lang nghe {ListenMs / 1000}s (Ctrl+C de dung) ---");
        try
        {
            await stream.WaitAsync(ListenMs);
        }
        catch (OperationCanceledException) { }

        // --- Buoc 5: Xem danh sach subscription server dang giu ---
        if (stream.IsConnected)
        {
            Console.WriteLine("\n--- Buoc 5: ListSubscriptionAsync ---");
            var masked = SampleConfig.MaskAccount(SampleConfig.AccountNo);
            var subs = await stream.Streaming.ListSubscriptionAsync(timeoutMs: 5000);
            foreach (var (channel, topics) in subs)
            {
                // Topic order.<tai khoan> co so tai khoan -> che truoc khi in
                var shown = topics.Select(t => t.Replace(SampleConfig.AccountNo, masked));
                Console.WriteLine($"  {channel}: {topics.Count} topic -> {string.Join(", ", shown.Take(8))}{(topics.Count > 8 ? ", ..." : "")}");
            }

            Console.WriteLine("\n--- Buoc 6: UnsubscribeAllAsync ---");
            // Tra ve nhung gi vua bi huy (lay tu danh sach server dang giu)
            var removed = await stream.Streaming.UnsubscribeAllAsync();
            Console.WriteLine($"  Da huy {removed.Values.Sum(v => v.Count)} topic");
            stream.Streaming.StopPingLoop();
        }

        Console.WriteLine($"\nTong ket: reconnect={stream.Streaming.ReconnectCount}, loi cuoi={(stream.Streaming.LastError?.GetType().Name ?? "khong")}");
        foreach (var (key, n) in counts.OrderBy(kv => kv.Key))
            Console.WriteLine($"  {key,-16}: {n}");

        Console.WriteLine("\n[Response] message_kinds");
        Console.WriteLine(counts.Count);
    }
}
