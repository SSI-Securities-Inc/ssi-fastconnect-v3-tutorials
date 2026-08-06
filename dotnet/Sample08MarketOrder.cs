// Sample 7 — Dat lenh Market (MP)
// Khop lenh nhanh theo gia thi truong hien tai.

using SsiSdk;

public static class Sample08MarketOrder
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var trading = new TradingClient(auth);

        // --- Buoc 1: Kiem tra suc mua/ban o gia thi truong ---
        var maxBs = await trading.Trading.GetMaxBuySellAtMarketPriceAsync(SampleConfig.AccountNo, "SSI");
        Console.WriteLine($"Max mua (market): {maxBs.MaxBuyQuantity} co phieu");
        Console.WriteLine($"Max ban (market): {maxBs.MaxSellQuantity} co phieu");

        // --- Buoc 2: Dat lenh Market mua ---
        Console.WriteLine("\n--- Dat lenh MARKET mua SSI ---");
        var result = await trading.Trading.PlaceMarketOrderAsync(
            SampleConfig.AccountNo, "SSI", OrderSide.Buy, 100);
        Console.WriteLine($"  Ket qua: OrderId={result.OrderId} Status={result.Status}");

        // --- Buoc 3: Kiem tra trang thai lenh ---
        Console.WriteLine("\n--- So lenh hom nay ---");
        var orders = await trading.Portfolio.GetTodayOrdersAsync(SampleConfig.AccountNo);
        foreach (var order in orders.TakeLast(3))
        {
            Console.WriteLine(
                $"  {order.OrderId} | {order.Symbol} {order.Side} " +
                $"{order.OrderType} | SL: {order.Quantity} " +
                $"| Khop: {order.FilledQuantity} | Trang thai: {order.Status}");
        }

        // --- Buoc 4: Cap nhat lai so du sau khi khop ---
        Console.WriteLine("\n--- So du sau giao dich ---");
        var balance = await trading.Portfolio.GetEquityBalanceAsync(SampleConfig.AccountNo);
        if (balance is not null)
        {
            Console.WriteLine($"  Tien mat kha dung: {balance.AccountBalance,15:N0}");
        }

        // --- Buoc 5: Cap nhat danh muc ---
        Console.WriteLine("\n--- Vi the sau giao dich ---");
        var positions = await trading.Portfolio.GetEquityPositionsAsync(SampleConfig.AccountNo);
        foreach (var pos in positions)
        {
            if (pos.Symbol == "SSI")
                Console.WriteLine($"  SSI | SL: {pos.Quantity} | Gia von: {pos.CostPrice:N0}");
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] max_buy_mkt|max_sell_mkt|buy_status");
        Console.WriteLine($"{maxBs.MaxBuyQuantity}|{maxBs.MaxSellQuantity}|{result.Status}");
    }
}
