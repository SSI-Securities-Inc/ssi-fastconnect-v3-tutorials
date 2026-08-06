// Sample 3 — Lay du lieu K-line (OHLC)
// Cung cap du lieu nen cho bieu do va phan tich ky thuat.

using SsiSdk;

public static class Sample04Ohlc
{
    private const string Symbol = "SSI";

    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth);

        var data = new DataClient(auth);

        // --- Buoc 1: Lay OHLC ngay gan nhat ---
        Console.WriteLine($"--- OHLC 1 ngay gan nhat ({Symbol}) ---");
        var daily = await data.MarketData.GetOhlc1DayHistoricalAsync(
            Symbol, "2026/03/01 00:00:00", "2026/03/27 23:59:59");

        foreach (var bar in daily.Take(5))
        {
            Console.WriteLine(
                $"  {bar.TradingDate} | " +
                $"O:{bar.OpenPrice,10} H:{bar.HighPrice,10} " +
                $"L:{bar.LowPrice,10} C:{bar.ClosePrice,10} " +
                $"V:{bar.Volume,12}");
        }

        // --- Buoc 2: Lay OHLC lich su theo khoang thoi gian ---
        Console.WriteLine($"\n--- OHLC 1 ngay lich su ({Symbol}) ---");
        var hist = await data.MarketData.GetOhlc1DayHistoricalAsync(
            Symbol, "2026/01/01 00:00:00", "2026/03/27 23:59:59", page: 1, size: 20);

        foreach (var bar in hist)
        {
            Console.WriteLine(
                $"  {bar.TradingDate} | " +
                $"O:{bar.OpenPrice,10} H:{bar.HighPrice,10} " +
                $"L:{bar.LowPrice,10} C:{bar.ClosePrice,10} " +
                $"V:{bar.Volume,12}");
        }

        // --- Buoc 3: Lay OHLC theo timeframe khac (1h) ---
        Console.WriteLine($"\n--- OHLC 1 gio gan nhat ({Symbol}) ---");
        var hourly = await data.MarketData.GetOhlc1HourAsync(Symbol);
        foreach (var bar in hourly.Take(5))
        {
            Console.WriteLine(
                $"  {bar.TradingDate} | " +
                $"O:{bar.OpenPrice,10} H:{bar.HighPrice,10} " +
                $"L:{bar.LowPrice,10} C:{bar.ClosePrice,10} " +
                $"V:{bar.Volume,12}");
        }

        // --- Buoc 4: Phan trang cho du lieu lon ---
        Console.WriteLine($"\n--- Paging OHLC 1 phut lich su ({Symbol}) ---");
        var page = 1;
        var totalBars = 0;
        while (true)
        {
            var bars = await data.MarketData.GetOhlc1MinuteHistoricalAsync(
                Symbol, "2026/03/25 09:00:00", "2026/03/25 15:00:00", page, 100);

            if (bars.Count == 0) break;
            totalBars += bars.Count;
            Console.WriteLine($"  Trang {page}: {bars.Count} nen (tong: {totalBars})");
            page++;
        }

        Console.WriteLine($"\nTong cong {totalBars} nen 1 phut duoc tai.");

        // --- Response Summary ---
        Console.WriteLine("\n[Response] daily_bars|hourly_bars|paging_1min");
        Console.WriteLine($"{daily.Count}|{hourly.Count}|{totalBars}");
        if (daily.Count > 0)
        {
            var c = daily[0];
            Console.WriteLine("[Response:first_daily] date|open|high|low|close|volume");
            Console.WriteLine($"{c.TradingDate}|{c.OpenPrice}|{c.HighPrice}|{c.LowPrice}|{c.ClosePrice}|{c.Volume}");
        }
    }
}
