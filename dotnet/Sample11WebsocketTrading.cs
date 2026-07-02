// Sample 11 — WebSocket trading real-time (trang thai lenh & danh muc)
// Nhan cap nhat tuc thoi ve lenh khop va danh muc tai khoan.

using SsiSdk;
using SsiSdk.Models;

static class Sample11WebsocketTrading
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        using var stream = new StreamClient(auth);

        // --- Buoc 1: Dang ky callback xu ly su kien trading ---
        stream.Streaming.SetOnTrading(msg =>
        {
            switch (msg)
            {
                case OrderStatusMessage order:
                    Console.WriteLine(
                        $"  [ORDER] {order.Symbol} {order.Side} | " +
                        $"OrderID: {order.OrderId} | Status: {order.Status} | " +
                        $"Khop: {order.FilledQuantity}/{order.Quantity}");
                    break;
                case PortfolioMessage portfolio:
                    Console.WriteLine(
                        $"  [PORTFOLIO] Account: {portfolio.AccountNo} | " +
                        $"Tong TS: {portfolio.TotalAsset} | Cash: {portfolio.CashBalance}");
                    break;
                default:
                    Console.WriteLine($"  [TRADING] {msg}");
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

        // --- Buoc 3: Subscribe trang thai lenh real-time ---
        Console.WriteLine("Subscribing trang thai lenh...");
        await stream.Streaming.SubscribeOrderStatusAsync(SampleConfig.AccountNo);

        // --- Buoc 4: Lang nghe lien tuc (5 phut) ---
        Console.WriteLine("\nDang lang nghe... (Ctrl+C de dung)\n");

        Console.CancelKeyPress += (_, e) =>
        {
            e.Cancel = true;
            Console.WriteLine("\nNgat ket noi...");
            stream.Disconnect();
        };

        try
        {
            await stream.WaitAsync(TimeSpan.FromMinutes(5));
        }
        catch (OperationCanceledException) { }
    }
}
