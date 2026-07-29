// Sample 13 — Đặt Lệnh Điều Kiện (FCO)
// Thể hiện đầy đủ các loại lệnh điều kiện (Fast Conditional Orders - FCO):
//   1. GTD (Good-Till-Date / Lệnh chờ theo ngày)
//   2. Stop (Lệnh dừng giá thị trường)
//   3. Stop Limit (Lệnh dừng giá giới hạn)
//   4. Trailing Stop (Lệnh dừng xu hướng)
//   5. Trailing Stop Limit (Lệnh dừng xu hướng giới hạn)
//   6. OCO (One-Cancels-the-Other / Lệnh Chốt lời & Cắt lỗ)
//   7. Bull Bear (Lệnh Hai đầu)
//   8. Truy vấn danh sách & Hủy lệnh FCO

using SsiSdk;

static class Sample13FcoOrder
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var trading = new TradingClient(auth);
        var symbol = "SSI";
        var fromDate = "2026/08/01 00:00:00";
        var toDate = "2026/08/30 23:59:59";

        Console.WriteLine("=== FASTCONNECT .NET SDK — SAMPLE 13: LỆNH ĐIỀU KIỆN (FCO) ===\n");

        try
        {
            // --- 1. Lệnh GTD (Good-Till-Date) ---
            Console.WriteLine("--- 1. Đặt lệnh GTD ---");
            var gtdRes = await trading.Trading.PlaceFcoGtdAsync(
                SampleConfig.AccountNo, symbol, OrderSide.Buy, 100, 26000, 0, fromDate, toDate);
            Console.WriteLine($"  GTD Result: FcoID={gtdRes.FCOID}");

            // --- 2. Lệnh Stop (Stop Market) ---
            Console.WriteLine("\n--- 2. Đặt lệnh Stop ---");
            var stopRes = await trading.Trading.PlaceFcoStopAsync(
                SampleConfig.AccountNo, symbol, OrderSide.Buy, 100, 27000, FCOOperator.GreaterOrEqual, fromDate, toDate);
            Console.WriteLine($"  Stop Result: FcoID={stopRes.FCOID}");

            // --- 3. Lệnh Stop Limit ---
            Console.WriteLine("\n--- 3. Đặt lệnh Stop Limit ---");
            var stopLimitRes = await trading.Trading.PlaceFcoStopLimitAsync(
                SampleConfig.AccountNo, symbol, OrderSide.Buy, 100, 27500, 0, 27000, FCOOperator.GreaterOrEqual, fromDate, toDate);
            Console.WriteLine($"  Stop Limit Result: FcoID={stopLimitRes.FCOID}");

            // --- 4. Lệnh Trailing Stop ---
            Console.WriteLine("\n--- 4. Đặt lệnh Trailing Stop ---");
            var trailingRes = await trading.Trading.PlaceFcoTrailingStopAsync(
                SampleConfig.AccountNo, symbol, OrderSide.Sell, 100, 28000, 1000, fromDate, toDate);
            Console.WriteLine($"  Trailing Stop Result: FcoID={trailingRes.FCOID}");

            // --- 5. Lệnh Trailing Stop Limit ---
            Console.WriteLine("\n--- 5. Đặt lệnh Trailing Stop Limit ---");
            var trailingLimitRes = await trading.Trading.PlaceFcoTrailingStopLimitAsync(
                SampleConfig.AccountNo, symbol, OrderSide.Sell, 100, 28000, 1000, 500, fromDate, toDate);
            Console.WriteLine($"  Trailing Stop Limit Result: FcoID={trailingLimitRes.FCOID}");

            // --- 6. Lệnh OCO (One-Cancels-the-Other) ---
            Console.WriteLine("\n--- 6. Đặt lệnh OCO ---");
            var ocoRes = await trading.Trading.PlaceFcoOcoAsync(
                SampleConfig.AccountNo, symbol, OrderSide.Sell, 100, 30000, 24000, 30000, 24000, 0, 0, fromDate, toDate);
            Console.WriteLine($"  OCO Result: FcoID={ocoRes.FCOID}");

            // --- 7. Lệnh Bull Bear ---
            Console.WriteLine("\n--- 7. Đặt lệnh Bull Bear ---");
            var bbRes = await trading.Trading.PlaceFcoBullBearAsync(
                SampleConfig.AccountNo, symbol, OrderSide.Buy, 100, 26000, 0, 30000, 24000, 30000, 24000, 0, 0, fromDate, toDate);
            Console.WriteLine($"  Bull Bear Result: FcoID={bbRes.FCOID}");

            // --- 8. Truy vấn danh sách lệnh FCO ---
            Console.WriteLine("\n--- 8. Danh sách lệnh FCO ---");
            var fcoList = await trading.Trading.GetFcoByAccountNoAsync(SampleConfig.AccountNo, 1, 10);
            Console.WriteLine($"  Tổng số lệnh FCO: {fcoList.ItemsCount}");
            foreach (var item in fcoList.FCOList.Take(5))
            {
                Console.WriteLine($"  FCO ID: {item.FCOID} | Mã: {item.Symbol} | Loại: {item.Type} | Trạng thái: {item.Status}");
            }

            // --- 9. Hủy lệnh FCO vừa tạo nếu có ---
            if (!string.IsNullOrEmpty(gtdRes.FCOID))
            {
                Console.WriteLine($"\n--- 9. Hủy lệnh FCO ID: {gtdRes.FCOID} ---");
                var cancelRes = await trading.Trading.CancelFcoAsync(gtdRes.FCOID);
                Console.WriteLine($"  Hủy FCO Result: FcoID={cancelRes.FCOID}");
            }

            Console.WriteLine("\n[Response] sample_13_fco_completed");
        }
        catch (Exception ex)
        {
            Console.WriteLine($"Lỗi khi thực thi Sample 13: {ex.Message}");
        }
    }
}
