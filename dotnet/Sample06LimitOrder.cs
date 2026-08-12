// Sample 6 — Dat lenh Limit (LO)
// Sample 06 — Đặt lệnh Limit

using SsiSdk;
using SsiSdk.Models;

public static class Sample06LimitOrder
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var trading = new TradingClient(auth);

        // --- Buoc 1: Kiem tra suc mua truoc ---
        var maxBs = await trading.Trading.GetMaxBuySellAsync(SampleConfig.AccountNo, "SSI", 26000);
        Console.WriteLine($"Suc mua toi da SSI @ 26,000: {maxBs.MaxBuyQuantity} co phieu");

        if (maxBs.MaxBuyQuantity < 100)
        {
            Console.WriteLine("Khong du suc mua, dung lai.");
            return;
        }

        // --- Buoc 2: Dat lenh Limit mua ---
        Console.WriteLine("\n--- Dat lenh LIMIT mua SSI ---");
        var buyResult = await trading.Trading.PlaceLimitOrderAsync(
            SampleConfig.AccountNo, "SSI", OrderSide.Buy, 100, 26000);
        Console.WriteLine($"  Ket qua: OrderId={buyResult.OrderId} Status={buyResult.Status}");

        // --- Buoc 3: Dat lenh Limit ban ---
        Console.WriteLine("\n--- Dat lenh LIMIT ban SSI ---");
        var sellResult = await trading.Trading.PlaceLimitOrderAsync(
            SampleConfig.AccountNo, "SSI", OrderSide.Sell, 100, 27000);
        Console.WriteLine($"  Ket qua: OrderId={sellResult.OrderId} Status={sellResult.Status}");

        // --- Buoc 4: Kiem tra lenh vua dat trong so lenh ---
        Console.WriteLine("\n--- So lenh hom nay ---");
        var orders = await trading.Portfolio.GetTodayOrdersAsync(SampleConfig.AccountNo);
        foreach (var order in orders.TakeLast(5))
        {
            Console.WriteLine(
                $"  {order.OrderId} | {order.Symbol} {order.Side} " +
                $"{order.OrderType} | SL: {order.Quantity} @ {order.Price} " +
                $"| Trang thai: {order.Status}");
        }

        // --- Response Summary ---
        Console.WriteLine($"\n[Response] max_buy_qty|buy_status|sell_status");
        Console.WriteLine($"{maxBs.MaxBuyQuantity}|{buyResult.Status}|{sellResult.Status}");
    }
}
