// Sample 15 — Vị thế (T1/T2) & sức mua PPMMR
// Chi DOC du lieu, khong dat lenh:
//   1. Vi the co phieu: so luong cho ve T1/T2 (mua va ban)
//   2. So du tien Equity: mua/ban cho thanh toan T0/T1/T2
//   3. PPMMR Equity (suc mua, ty le ky quy, no)
//   4. Phai sinh (neu config.json co "derivative_account"): so du, vi the mo/dong, PPMMR
//
// Luu y: tren vi the co phieu (EquityPosition) server chi co T1/T2; khong co truong T0.
// Mua/ban T0 chi nam o so du tien (EquityAccountBalance) va PPMMR.

using SsiSdk;
using SsiSdk.Models;

public static class Sample15PositionsPpmmr
{
    public static async Task RunAsync()
    {
        SampleConfig.EnsureConfigured();
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var trading = new TradingClient(auth);
        var account = SampleConfig.AccountNo;
        Console.WriteLine($"Tai khoan Equity: {SampleConfig.MaskAccount(account)}\n");

        // --- Buoc 1: Vi the co phieu, kem so luong cho ve T1/T2 ---
        Console.WriteLine("--- Buoc 1: Vi the co phieu ---");
        // null = tai khoan chua co vi the (khac voi danh sach rong)
        var positions = await trading.Portfolio.GetEquityPositionsAsync(account) ?? [];
        Console.WriteLine($"  {"Ma",-8} {"SL",8} {"Ban duoc",9} {"Mua T1",8} {"Mua T2",8} {"Ban T1",8} {"Ban T2",8} {"Gia von",10}");
        foreach (var p in positions)
        {
            Console.WriteLine(
                $"  {p.Symbol,-8} {p.Quantity,8} {p.SellableQuantity,9} " +
                $"{p.T1BuyQuantity,8} {p.T2BuyQuantity,8} {p.T1SellQuantity,8} {p.T2SellQuantity,8} {p.CostPrice,10:N0}");
        }
        var pendingBuy = positions.Sum(p => (long)p.T1BuyQuantity + p.T2BuyQuantity);
        var pendingSell = positions.Sum(p => (long)p.T1SellQuantity + p.T2SellQuantity);
        Console.WriteLine($"  Tong cho ve: mua {pendingBuy:N0} cp, ban {pendingSell:N0} cp");

        // --- Buoc 2: So du tien Equity theo ngay thanh toan ---
        Console.WriteLine("\n--- Buoc 2: So du tien Equity (T0/T1/T2) ---");
        var balance = await trading.Portfolio.GetEquityBalanceAsync(account);
        if (balance is null)
        {
            Console.WriteLine("  (khong co du lieu so du)");
        }
        else
        {
            Console.WriteLine($"  Tien mat kha dung : {balance.AccountBalance,15:N0}");
            Console.WriteLine($"  Mua  T0/T1/T2     : {balance.BuyT0,13:N0} / {balance.BuyT1,13:N0} / {balance.BuyT2,13:N0}");
            Console.WriteLine($"  Ban  T0/T1/T2     : {balance.SellT0,13:N0} / {balance.SellT1,13:N0} / {balance.SellT2,13:N0}");
            Console.WriteLine($"  Mua/Ban chua khop : {balance.BuyUnmatched:N0} / {balance.SellUnmatched:N0}");
        }

        // --- Buoc 3: PPMMR Equity ---
        Console.WriteLine("\n--- Buoc 3: PPMMR Equity ---");
        var ppmmr = await trading.Portfolio.GetEquityPpmmrAsync(account);
        if (ppmmr is null)
        {
            Console.WriteLine("  (khong co du lieu PPMMR)");
        }
        else
        {
            // Tat ca la decimal, ke ca ty le (MarginRatio)
            Console.WriteLine($"  Suc mua (PurchasingPower) : {ppmmr.PurchasingPower,15:N0}");
            Console.WriteLine($"  Ty le ky quy (MarginRatio): {ppmmr.MarginRatio,15}");
            Console.WriteLine($"  EE / EE50 / EE70          : {ppmmr.EE:N0} / {ppmmr.EE50:N0} / {ppmmr.EE70:N0}");
            Console.WriteLine($"  Tong no / Tong tai san    : {ppmmr.TotalDebt:N0} / {ppmmr.TotalAssets:N0}");
            Console.WriteLine($"  Mua T0/T1/T2              : {ppmmr.BuyT0:N0} / {ppmmr.BuyT1:N0} / {ppmmr.BuyT2:N0}");
            Console.WriteLine($"  Ban T0/T1/T2              : {ppmmr.SellT0:N0} / {ppmmr.SellT1:N0} / {ppmmr.SellT2:N0}");
        }

        // --- Buoc 4: Phai sinh (tuy chon) ---
        var derivAccount = SampleConfig.DerivativeAccount;
        if (derivAccount is null)
        {
            Console.WriteLine("\n--- Buoc 4: bo qua phai sinh (config.json chua co \"derivative_account\") ---");
        }
        else
        {
            Console.WriteLine($"\n--- Buoc 4: Phai sinh {SampleConfig.MaskAccount(derivAccount)} ---");

            var dBalance = await trading.Portfolio.GetDerivativeBalanceAsync(derivAccount);
            if (dBalance is not null)
            {
                // Dung WithdrawableSSI / WithdrawableVSDC (CashWithdrawableSSI/VSDC da loi thoi)
                Console.WriteLine($"  So du: {dBalance.AccountBalance:N0} | Rut duoc SSI/VSDC: {dBalance.WithdrawableSSI:N0} / {dBalance.WithdrawableVSDC:N0}");
                Console.WriteLine($"  Lai/lo: floating {dBalance.FloatingPL:N0} | trading {dBalance.TradingPL:N0} | tong {dBalance.TotalPL:N0}");
            }

            var all = await trading.Portfolio.GetDerivativePositionsAsync(derivAccount);
            var open = all?.OpenPositions ?? [];
            var closed = all?.ClosedPositions ?? [];
            Console.WriteLine($"  Vi the mo: {open.Count} | dong: {closed.Count}");
            foreach (var d in open)
            {
                Console.WriteLine($"    {d.Symbol,-12} Long {d.Long,4} Short {d.Short,4} Net {d.Net,4} | floating PL {d.FloatingPL:N0}");
            }

            var dPpmmr = await trading.Portfolio.GetDerivativePpmmrAsync(derivAccount);
            if (dPpmmr is not null)
            {
                Console.WriteLine($"  PPMMR: tong tai san SSI {dPpmmr.TotalAssetSSI:N0} | EE SSI {dPpmmr.EeSSI:N0} | margin call SSI {dPpmmr.MarginCallSSI:N0}");
            }
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] positions|pending_buy_t1t2|pending_sell_t1t2|purchasing_power");
        Console.WriteLine($"{positions.Count}|{pendingBuy}|{pendingSell}|{ppmmr?.PurchasingPower}");
    }
}
