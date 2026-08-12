// Sample 05 — Lấy số dư tài khoản (Account Balance)
// Kiem tra kha nang giao dich truoc khi dat lenh.

using SsiSdk;

public static class Sample05Balance
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var trading = new TradingClient(auth);

        // --- Buoc 1: Lay danh sach tai khoan ---
        var accounts = await trading.Account.GetAccountInfoAsync();
        Console.WriteLine("Danh sach tai khoan:");
        foreach (var acc in accounts)
        {
            Console.WriteLine($"  - {acc.AccountNo} ({acc.AccountType})");
        }

        // --- Buoc 2: Lay so du tai khoan Equity ---
        Console.WriteLine($"\n--- So du tai khoan Equity: {SampleConfig.AccountNo} ---");
        var balance = await trading.Portfolio.GetEquityBalanceAsync(SampleConfig.AccountNo);
        if (balance is not null)
        {
            Console.WriteLine($"  Tien mat kha dung  : {balance.AccountBalance,15:N0}");
            Console.WriteLine($"  Tong no            : {balance.TotalDebt,15:N0}");
            Console.WriteLine($"  Mua T0/T1/T2       : {balance.BuyT0,12:N0} / {balance.BuyT1,12:N0} / {balance.BuyT2,12:N0}");
            Console.WriteLine($"  Ban T0/T1/T2       : {balance.SellT0,12:N0} / {balance.SellT1,12:N0} / {balance.SellT2,12:N0}");
        }

        // --- Buoc 3: Kiem tra suc mua toi da cho mot ma ---
        Console.WriteLine("\n--- Suc mua/ban toi da: SSI ---");
        var maxBs = await trading.Trading.GetMaxBuySellAsync(SampleConfig.AccountNo, "SSI", 26000);
        Console.WriteLine($"  Max mua : {maxBs.MaxBuyQuantity,10} co phieu");
        Console.WriteLine($"  Max ban : {maxBs.MaxSellQuantity,10} co phieu");
        Console.WriteLine($"  Suc mua : {maxBs.PurchasePower,15}");

        // --- Buoc 4: Logic kiem tra truoc khi dat lenh ---
        var desiredQuantity = 100;
        var desiredPrice = 26000.0;
        var requiredAmount = desiredQuantity * desiredPrice;

        if (balance is not null && balance.AccountBalance >= requiredAmount)
        {
            Console.WriteLine($"\n  Du dieu kien: can {requiredAmount:N0}, co {balance.AccountBalance:N0}");
            Console.WriteLine("  -> Cho phep dat lenh mua.");
        }
        else
        {
            Console.WriteLine($"\n  Khong du: can {requiredAmount:N0}, chi co {balance?.AccountBalance:N0}");
            Console.WriteLine("  -> Chan dat lenh.");
        }

        // --- Buoc 5: Xem vi the hien co ---
        Console.WriteLine($"\n--- Vi the co phieu ({SampleConfig.AccountNo}) ---");
        var positions = await trading.Portfolio.GetEquityPositionsAsync(SampleConfig.AccountNo);
        foreach (var pos in positions)
        {
            Console.WriteLine(
                $"  {pos.Symbol,-10} | SL: {pos.Quantity,8} | " +
                $"Ban duoc: {pos.SellableQuantity,8} | Gia von: {pos.CostPrice,10:N0}");
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] accounts|avail_cash|max_buy_qty|max_sell_qty|positions");
        Console.WriteLine($"{accounts.Count}|{balance?.AccountBalance}|{maxBs.MaxBuyQuantity}|{maxBs.MaxSellQuantity}|{positions.Count}");
        if (positions.Count > 0)
        {
            var p = positions[0];
            Console.WriteLine("[Response:first_pos] symbol|quantity|sellable|cost_price");
            Console.WriteLine($"{p.Symbol}|{p.Quantity}|{p.SellableQuantity}|{p.CostPrice}");
        }
    }
}
