// Sample 02 — Lấy danh sách chỉ số thị trường (Index)
// Hiển thị VN-Index, HNX-Index... trên dashboard.

using SsiSdk;

public static class Sample02IndexList
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth);

        var data = new DataClient(auth);

        // --- Buoc 1: Lay toan bo chi so ---
        var allIndexes = await data.MarketData.GetIndexesAsync();
        Console.WriteLine($"Tong so chi so: {allIndexes.Count}\n");

        foreach (var idx in allIndexes)
        {
            Console.WriteLine($"  {idx.Index,-15} | {idx.IndexName,-30} | San: {idx.Board}");
        }

        // --- Buoc 2: Loc chi so theo san HOSE ---
        Console.WriteLine("\n--- Chi so san HOSE ---");
        var hoseIndexes = await data.MarketData.GetIndexesByBoardAsync(Board.HOSE);
        foreach (var idx in hoseIndexes)
        {
            Console.WriteLine($"  {idx.Index,-15} | {idx.IndexName}");
        }

        // --- Buoc 3: Lay chi tiet summary cho mot chi so cu the ---
        Console.WriteLine("\n--- VN-Index Summary ---");
        var summary = await data.MarketData.GetIndexSummaryAsync("VNINDEX");
        if (summary is not null)
        {
            Console.WriteLine($"  Gia tri Index    : {summary.IndexValue}");
            Console.WriteLine($"  Thay doi         : {summary.IndexChange} ({summary.IndexChangePercent}%)");
            Console.WriteLine($"  Tong KL khop     : {summary.TotalMatch}");
            Console.WriteLine($"  Tong GT khop     : {summary.TotalMatchValue}");
            Console.WriteLine($"  Tang / Giam / Dung: {summary.TotalAdvanceStock} / {summary.TotalDeclineStock} / {summary.TotalSteadyStock}");
            Console.WriteLine($"  Tran / San       : {summary.TotalCeilingStock} / {summary.TotalFloorStock}");
        }
        else
        {
            Console.WriteLine("Khong lay duoc summary cho VN-Index.");
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] total_indexes|hose_indexes|first_index");
        var first = allIndexes.Count > 0 ? allIndexes[0].Index : "";
        Console.WriteLine($"{allIndexes.Count}|{hoseIndexes.Count}|{first}");
    }
}
