// Sample 19 — Su kien doanh nghiep (Corporate actions)
// Xem co tuc, co phieu thuong, quyen mua... cua cac ma. Chi DOC du lieu, khong dat lenh.
//
// Luong:
//   1. GetCorporateActionsAsync(): su kien cua HOM NAY
//   2. GetCorporateActionsHistoricalAsync(from, to): su kien trong mot khoang ngay o qua khu
//      (server cho toi da 5 ngay moi lan goi)
//   3. Loc theo ma (symbol) va ngon ngu ("vn" / "en")
//   4. Khoang ngay qua rong bi SDK chan tai client (ValidationException), khong goi server
//
// Luu y: SDK tu lay het cac trang. Moi truong la chuoi dung nhu server gui (ngay YYYY/MM/DD,
// ratio khong duoc parse, truong de trong la ""). Ngay truyen vao phai dung dang YYYY/MM/DD.

using SsiSdk;
using SsiSdk.Models;

public static class Sample19CorporateActions
{
    private const string DateFmt = "yyyy/MM/dd";

    private static void Show(List<CorporateAction> actions, int limit = 10)
    {
        static string Dash(string s) => string.IsNullOrEmpty(s) ? "-" : s;
        foreach (var a in actions.Take(limit))
        {
            Console.WriteLine(
                $"  {a.Symbol,-8} | {a.Exchange,-6} | GDKHQ: {Dash(a.ExrightDate),-10} " +
                $"| Chot: {Dash(a.RecordDate),-10} | Phat hanh: {Dash(a.IssueDate),-10} " +
                $"| Ty le: {Dash(a.Ratio),-8} | {a.EventTitle}");
        }
        if (actions.Count > limit) Console.WriteLine($"  ... con {actions.Count - limit} su kien");
    }

    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth);

        var data = new DataClient(auth);

        // --- Buoc 1: Su kien cua hom nay ---
        Console.WriteLine("--- Su kien doanh nghiep hom nay ---");
        var todayActions = await data.MarketData.GetCorporateActionsAsync();
        Console.WriteLine($"Tong so su kien: {todayActions.Count}");
        Show(todayActions);

        // --- Buoc 2: Su kien trong 5 ngay gan nhat (toi da 5 ngay cho moi lan goi) ---
        var now = DateTime.Now;
        var fromDate = now.AddDays(-4).ToString(DateFmt);
        var toDate = now.ToString(DateFmt);
        Console.WriteLine($"\n--- Su kien tu {fromDate} den {toDate} ---");
        var rangeActions = await data.MarketData.GetCorporateActionsHistoricalAsync(fromDate, toDate);
        Console.WriteLine($"Tong so su kien: {rangeActions.Count}");
        Show(rangeActions);

        // --- Buoc 3: Loc theo ma, tra tieu de bang tieng Anh ---
        Console.WriteLine("\n--- Su kien cua ma SSI (tieng Anh) ---");
        var ssiActions = await data.MarketData.GetCorporateActionsHistoricalAsync(fromDate, toDate, "SSI", "en");
        Console.WriteLine($"Tong so su kien: {ssiActions.Count}");
        Show(ssiActions);

        // --- Buoc 4: Khoang ngay qua rong bi chan tai SDK ---
        Console.WriteLine("\n--- Khoang 30 ngay (vuot gioi han 5 ngay) ---");
        try
        {
            await data.MarketData.GetCorporateActionsHistoricalAsync(now.AddDays(-30).ToString(DateFmt), toDate);
        }
        catch (ValidationException ex)
        {
            Console.WriteLine($"  SDK tu choi truoc khi goi server: {ex.Message}");
            Console.WriteLine("  -> chia khoang ngay thanh cac doan <= 5 ngay roi goi tung doan.");
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] today_count|range_count|ssi_range_count");
        Console.WriteLine($"{todayActions.Count}|{rangeActions.Count}|{ssiActions.Count}");
    }
}
