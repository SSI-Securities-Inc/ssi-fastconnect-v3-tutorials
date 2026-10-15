// Sample 03 — Lấy dữ liệu K-line (OHLC)
// Cung cap du lieu nen cho bieu do va phan tich ky thuat.

using SsiSdk;
using SsiSdk.Models;

public static class Sample03Ohlc
{
    private const string Symbol = "SSI";
    private const string DateFmt = "yyyy/MM/dd HH:mm:ss";

    private static void PrintBar(OhlcData bar) =>
        Console.WriteLine(
            $"  {bar.TradingDate} | " +
            $"O:{bar.OpenPrice,10} H:{bar.HighPrice,10} " +
            $"L:{bar.LowPrice,10} C:{bar.ClosePrice,10} " +
            $"V:{bar.Volume,12}");

    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth);

        var data = new DataClient(auth);

        // Khoang thoi gian tinh tu hom nay de sample khong bi loi thoi
        var now = DateTime.Now;
        var to = now.ToString(DateFmt);

        // --- Buoc 1: Lay OHLC ngay (30 ngay gan nhat) ---
        Console.WriteLine($"--- OHLC 1 ngay gan nhat ({Symbol}) ---");
        var daily = await data.MarketData.GetOhlc1DayHistoricalAsync(
            Symbol, now.AddDays(-30).ToString(DateFmt), to);

        foreach (var bar in daily.Take(5)) PrintBar(bar);

        // --- Buoc 2: Lay OHLC lich su theo khoang thoi gian, co phan trang thu cong ---
        Console.WriteLine($"\n--- OHLC 1 ngay lich su ({Symbol}) ---");
        var hist = await data.MarketData.GetOhlc1DayHistoricalAsync(
            Symbol, now.AddDays(-90).ToString(DateFmt), to, page: 1, size: 20);

        foreach (var bar in hist) PrintBar(bar);

        // --- Buoc 3: Lay OHLC theo timeframe khac (1h) ---
        Console.WriteLine($"\n--- OHLC 1 gio gan nhat ({Symbol}) ---");
        var hourly = await data.MarketData.GetOhlc1HourAsync(Symbol);
        foreach (var bar in hourly.Take(5)) PrintBar(bar);

        // --- Buoc 4: Lay het du lieu lon bang IterOhlcAsync (SDK tu phan trang) ---
        // Khong can viet vong lap page++ nua: SDK goi trang tiep theo den khi gap trang ngan.
        // Thu tu la thu tu cua server (moi nhat truoc).
        Console.WriteLine($"\n--- IterOhlcAsync: OHLC 1 phut {Symbol}, 5 ngay gan nhat ---");
        var totalBars = 0;
        await foreach (var bar in data.MarketData.IterOhlcAsync(
            Symbol, Timeframe.Minute1, now.AddDays(-5).ToString(DateFmt), to, size: 100))
        {
            totalBars++;
            if (totalBars <= 3) PrintBar(bar);
        }
        Console.WriteLine($"  Tong cong {totalBars} nen 1 phut duoc tai.");

        // --- Buoc 5: GetOhlcAllAsync — gom het vao List, co the sap xep cu -> moi ---
        Console.WriteLine($"\n--- GetOhlcAllAsync: OHLC 1 ngay {Symbol}, 60 ngay, cu -> moi ---");
        var allDaily = await data.MarketData.GetOhlcAllAsync(
            Symbol, Timeframe.Day1, now.AddDays(-60).ToString(DateFmt), to, ascending: true);
        Console.WriteLine($"  {allDaily.Count} nen ngay");
        foreach (var bar in allDaily.Take(3)) PrintBar(bar);

        // --- Buoc 6: Timeframe khong ho tro ---
        // Server tu choi 1w / 1M (loi 400210) nen SDK nem ValidationException ngay, khong goi mang.
        // Muon nen tuan/thang: tu gop tu nen ngay.
        Console.WriteLine("\n--- Timeframe 1 tuan (khong ho tro) ---");
        try
        {
            // Ham nay da [Obsolete]; goi co chu y de minh hoa loi, dung dung trong code that
#pragma warning disable CS0618
            await data.MarketData.GetOhlc1WeekHistoricalAsync(Symbol, now.AddDays(-30).ToString(DateFmt), to);
#pragma warning restore CS0618
        }
        catch (ValidationException ex)
        {
            Console.WriteLine($"  Bi tu choi dung nhu mong doi: {ex.Message}");
        }
        Console.WriteLine($"  Timeframe duoc ho tro: {string.Join(", ", Timeframe.Allowed.Order())}");

        // --- Response Summary ---
        Console.WriteLine("\n[Response] daily_bars|hourly_bars|iter_1min");
        Console.WriteLine($"{daily.Count}|{hourly.Count}|{totalBars}");
        if (daily.Count > 0)
        {
            var c = daily[0];
            Console.WriteLine("[Response:first_daily] date|open|high|low|close|volume");
            Console.WriteLine($"{c.TradingDate}|{c.OpenPrice}|{c.HighPrice}|{c.LowPrice}|{c.ClosePrice}|{c.Volume}");
        }
    }
}
