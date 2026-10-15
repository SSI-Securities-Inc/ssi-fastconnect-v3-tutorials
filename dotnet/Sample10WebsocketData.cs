// Sample 10 — WebSocket dữ liệu thị trường real-time
// Nhận tick data (giá khớp, bảng giá, room nước ngoài) tức thời.

using SsiSdk;
using SsiSdk.Models;

public static class Sample10WebsocketData
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth);

        using var stream = new StreamClient(auth);

        // --- Buoc 1: Dang ky callback du lieu thi truong ---
        stream.Streaming.SetOnData(msg =>
        {
            switch (msg)
            {
                case TradeMessage trade:
                    Console.WriteLine(
                        $"  [TRADE] {trade.Symbol} | Gia: {trade.Price} " +
                        $"| KL: {trade.Quantity} | Side: {trade.Side}");
                    break;
                case QuoteMessage quote:
                    Console.WriteLine(
                        $"  [QUOTE] {quote.Symbol} | " +
                        $"Bid: [{string.Join(", ", quote.BidPrices.Take(3))}] | " +
                        $"Ask: [{string.Join(", ", quote.AskPrices.Take(3))}]");
                    break;
                case ForeignRoomMessage room:
                    Console.WriteLine(
                        $"  [ROOM]  {room.Symbol} | " +
                        $"Room con: {room.CurrentRoom}/{room.TotalRoom}");
                    break;
                case IndexTickMessage idx:
                    Console.WriteLine(
                        $"  [INDEX] {idx.Symbol} | Gia tri: {idx.Value} | " +
                        $"Thay doi: {idx.Change} ({idx.ChangePercent}%)");
                    break;
                default:
                    Console.WriteLine($"  [DATA]  {msg}");
                    break;
            }
        });

        stream.Streaming.SetOnHeartbeat(msg =>
        {
            Console.WriteLine($"  [HEARTBEAT] {msg.Status} - {msg.Message}");
        });

        // --- Buoc 2: Mo ket noi WebSocket ---
        Console.WriteLine("Dang ket noi WebSocket...");
        await stream.ConnectAsync();
        Console.WriteLine("Da ket noi!\n");

        // --- Buoc 3: Subscribe du lieu theo symbol ---
        Console.WriteLine("Subscribing du lieu symbol...");
        await stream.Streaming.SubscribeSymbolAsync(["SSI", "HPG", "VNM"]);

        // --- Buoc 4: Subscribe tick cua chinh chi so (trade.index.<ma>) ---
        // Luu y: SubscribeIndexAsync KHONG phai tick chi so. No gui trade/quote/room.<ma>@*
        // de lay du lieu tung ma trong ro chi so (VD VN30) -> xem Sample 16.
        Console.WriteLine("Subscribing tick chi so...");
        await stream.Streaming.SubscribeIndexTradeAsync(["VNINDEX", "VN30"]);

        // --- Buoc 5: Lang nghe lien tuc (5 phut) ---
        Console.WriteLine("\nDang lang nghe... (Ctrl+C de dung)\n");

        Console.CancelKeyPress += (_, e) =>
        {
            e.Cancel = true;
            Console.WriteLine("\nNgat ket noi...");
            stream.Disconnect();
        };

        try
        {
            // WaitAsync nhan so mili-giay (khong phai TimeSpan)
            await stream.WaitAsync(5 * 60 * 1000);
        }
        catch (OperationCanceledException) { }
    }
}
