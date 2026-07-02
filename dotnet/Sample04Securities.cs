// Sample 4 — Lay danh sach co phieu theo san
// Tao watchlist/screener theo tieu chi thi truong.

using SsiSdk;

static class Sample04Securities
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth);

        var data = new DataClient(auth);

        // --- Buoc 1: Lay danh sach co phieu san HOSE ---
        Console.WriteLine("--- Co phieu san HOSE ---");
        var hoseSecurities = await data.MarketData.GetSecuritiesInfoByBoardAsync(Board.HOSE);
        Console.WriteLine($"Tong so ma: {hoseSecurities.Count}\n");

        foreach (var sec in hoseSecurities.Take(10))
        {
            Console.WriteLine(
                $"  {sec.Symbol,-10} | {(sec.SymbolNameVi ?? sec.SymbolNameEn ?? ""),-30} " +
                $"| Lot: {sec.LotSize}");
        }

        // --- Buoc 2: Lay danh sach co phieu san HNX ---
        Console.WriteLine("\n--- Co phieu san HNX ---");
        var hnxSecurities = await data.MarketData.GetSecuritiesInfoByBoardAsync(Board.HNX);
        Console.WriteLine($"Tong so ma: {hnxSecurities.Count}");

        foreach (var sec in hnxSecurities.Take(10))
        {
            Console.WriteLine(
                $"  {sec.Symbol,-10} | {(sec.SymbolNameVi ?? sec.SymbolNameEn ?? ""),-30} " +
                $"| Lot: {sec.LotSize}");
        }

        // --- Buoc 3: Lay theo chi so (index) ---
        Console.WriteLine("\n--- Co phieu thuoc VN30 ---");
        var vn30Securities = await data.MarketData.GetSecuritiesInfoByIndexAsync("VN30");
        Console.WriteLine($"Tong so ma: {vn30Securities.Count}");

        foreach (var sec in vn30Securities)
        {
            Console.WriteLine($"  {sec.Symbol,-10} | {sec.SymbolNameVi ?? sec.SymbolNameEn ?? ""}");
        }

        // --- Buoc 4: Xem thong tin chi tiet mot ma ---
        Console.WriteLine("\n--- Chi tiet ma SSI ---");
        var info = await data.MarketData.GetSecuritiesInfoAsync("SSI");
        if (info is not null)
        {
            Console.WriteLine($"  Ma          : {info.Symbol}");
            Console.WriteLine($"  Ten (VI)    : {info.SymbolNameVi}");
            Console.WriteLine($"  Ten (EN)    : {info.SymbolNameEn}");
            Console.WriteLine($"  San         : {info.Board}");
            Console.WriteLine($"  Lot size    : {info.LotSize}");
            Console.WriteLine($"  ICB Code    : {info.IcbCode}");
            Console.WriteLine($"  ICB Name    : {info.IcbName}");
            Console.WriteLine($"  Listed Shares: {info.ListedShares}");
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] hose_count|hnx_count|vn30_count|symbol|lot_size|listed_shares");
        Console.WriteLine($"{hoseSecurities.Count}|{hnxSecurities.Count}|{vn30Securities.Count}|{info?.Symbol}|{info?.LotSize}|{info?.ListedShares}");
    }
}
