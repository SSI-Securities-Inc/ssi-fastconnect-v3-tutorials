// Sample 10 — WebSocket du lieu thi truong real-time
// Nhan tick data (gia khop, bang gia, room nuoc ngoai) tuc thoi.

using SsiSdk;
using SsiSdk.Models;

public static class Sample11WebsocketData
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

        // --- Buoc 4: Subscribe du lieu theo index ---
        Console.WriteLine("Subscribing du lieu index...");
        await stream.Streaming.SubscribeIndexAsync(["VNINDEX", "HNX-INDEX"]);

        // --- Buoc 5: Lang nghe lien tuc (5 phut) ---
        Console.WriteLine("\nDang lang nghe... (Ctrl+C de dung)\n");

        using var cts = new CancellationTokenSource();
        Console.CancelKeyPress += (_, e) =>
        {
            e.Cancel = true;
            Console.WriteLine("\nNgat ket noi...");
            cts.Cancel();
        };

        try
        {
            await stream.WaitAsync(TimeSpan.FromMinutes(5));
        }
        catch (OperationCanceledException) { }
    }
}
