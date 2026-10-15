// Sample 14 — Đặt / hủy nhiều lệnh một lần (Batch Orders)
// PlaceBatchOrdersAsync + CancelBatchOrdersAsync: toi da Config.MaxBatchOrders lenh / request,
// ket qua tra ve RIENG TUNG LENH (mot lenh loi khong lam hong ca batch sau khi da gui).
//
// CANH BAO: sample nay DAT LENH THAT (LO mua SSI, gia thap hon gia dong cua gan nhat ~5% de lenh
// nam cho, khong khop) roi huy lai ngay. Chi chay tren tai khoan thu nghiem. Neu sample dung
// giua chung, tu kiem tra va huy cac lenh con lai tren so lenh.

using SsiSdk;
using SsiSdk.Models;

public static class Sample14BatchOrders
{
    private const string Symbol = "SSI";
    private const int QuantityPerOrder = 100;
    private const int OrderCount = 3;

    // Buoc gia san HOSE: <10.000 -> 10, 10.000-49.950 -> 50, >=50.000 -> 100. Lam tron XUONG.
    private static decimal FloorToTick(decimal price)
    {
        var tick = price < 10_000m ? 10m : price < 50_000m ? 50m : 100m;
        return Math.Floor(price / tick) * tick;
    }

    private static void PrintResults(string title, BatchOrderResponse response)
    {
        Console.WriteLine($"  {title}: batchRequestId={response.BatchRequestId}, {response.Results.Count} ket qua");
        foreach (var r in response.Results)
        {
            // Success=false -> doc ErrorCode/ErrorMessage cua chinh lenh do
            var detail = r.Success
                ? $"OrderId={r.OrderId} Status={r.Status}"
                : $"LOI {r.ErrorCode}: {r.ErrorMessage}";
            Console.WriteLine($"    - clientRequestId={r.ClientRequestId} | {(r.Success ? "OK " : "FAIL")} | {detail}");
        }
    }

    public static async Task RunAsync()
    {
        SampleConfig.EnsureConfigured();
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var trading = new TradingClient(auth);
        var data = new DataClient(auth);
        var account = SampleConfig.AccountNo;

        Console.WriteLine($"Tai khoan: {SampleConfig.MaskAccount(account)} | Gioi han batch: {config.MaxBatchOrders} lenh/request\n");

        // --- Buoc 1: Gioi han batch duoc kiem tra TRUOC khi gui mang ---
        // Vuot Config.MaxBatchOrders -> ValidationException ngay tai client, khong co request nao di.
        Console.WriteLine("--- Buoc 1: Batch vuot gioi han ---");
        var oversize = Enumerable.Range(0, config.MaxBatchOrders + 1)
            .Select(i => new BatchPlaceOrderItem(account, Symbol, OrderSide.Buy, QuantityPerOrder, 1m, OrderType.LO))
            .ToList();
        try
        {
            await trading.Trading.PlaceBatchOrdersAsync(oversize);
            Console.WriteLine("  (khong mong doi: batch vuot gioi han van duoc chap nhan)");
        }
        catch (ValidationException ex)
        {
            Console.WriteLine($"  Bi tu choi tai client nhu mong doi: {ex.Message}");
        }

        // --- Buoc 2: Chon gia xa thi truong (5% duoi gia dong cua gan nhat) ---
        var now = DateTime.Now;
        var bars = await data.MarketData.GetOhlc1DayHistoricalAsync(
            Symbol, now.AddDays(-14).ToString("yyyy/MM/dd HH:mm:ss"), now.ToString("yyyy/MM/dd HH:mm:ss"), size: 5);
        if (bars.Count == 0)
        {
            Console.WriteLine("Khong lay duoc gia tham chieu, dung lai (khong dat lenh).");
            return;
        }
        var basePrice = FloorToTick(bars[0].ClosePrice * 0.95m);
        Console.WriteLine($"\n--- Buoc 2: Gia dat = {basePrice:N0} (95% gia dong cua gan nhat {bars[0].ClosePrice:N0}) ---");

        // Kiem tra suc mua cho toan bo batch truoc khi dat
        var maxBs = await trading.Trading.GetMaxBuySellAsync(account, Symbol, basePrice);
        if (maxBs.MaxBuyQuantity < QuantityPerOrder * OrderCount)
        {
            Console.WriteLine($"Khong du suc mua ({maxBs.MaxBuyQuantity} < {QuantityPerOrder * OrderCount}), dung lai.");
            return;
        }

        // Xac nhan thu cong truoc khi dat lenh that (mac dinh tu choi)
        Console.Write($"Se dat {OrderCount} lenh LO mua {Symbol} x{QuantityPerOrder} THAT. Go YES de tiep tuc: ");
        if (!string.Equals(Console.ReadLine()?.Trim(), "YES", StringComparison.Ordinal))
        {
            Console.WriteLine("Da huy, khong dat lenh.");
            return;
        }

        // --- Buoc 3: Dat batch voi clientRequestId do minh dat ---
        // clientRequestId la khoa idempotency: gui lai cung id thi server khong tao lenh thu hai.
        // Dat id theo logic cua minh (ngan, SDK gioi han do dai) de sau nay hoi/huy theo id.
        var runId = Guid.NewGuid().ToString("N")[..8];
        var items = Enumerable.Range(0, OrderCount)
            .Select(i => new BatchPlaceOrderItem(
                account, Symbol, OrderSide.Buy, QuantityPerOrder,
                basePrice - i * 50m, OrderType.LO,
                ClientRequestId: $"t14-{runId}-{i}"))
            .ToList();

        Console.WriteLine("\n--- Buoc 3: PlaceBatchOrdersAsync ---");
        var placed = new List<BatchOrderResult>();
        BatchOrderResponse placeResp;
        try
        {
            placeResp = await trading.Trading.PlaceBatchOrdersAsync(items);
        }
        catch (SsiException ex)
        {
            // Loi o muc request (xac thuc, chu ky, rate limit...): chua co lenh nao chac chan duoc tao
            Console.WriteLine($"  Batch that bai: {ex.GetType().Name} [{ex.Code}] {ex.Message}");
            return;
        }
        PrintResults("Ket qua dat", placeResp);
        placed.AddRange(placeResp.Results);

        // --- Buoc 4: Gui lai CUNG batch (mo phong mat ket noi roi retry) ---
        // Moi lan goi SDK tao batchRequestId moi, nen idempotency nam o clientRequestId cua tung lenh.
        Console.WriteLine("\n--- Buoc 4: Retry cung clientRequestId (idempotency) ---");
        try
        {
            var retryResp = await trading.Trading.PlaceBatchOrdersAsync(items);
            PrintResults("Ket qua retry", retryResp);
            placed.AddRange(retryResp.Results);
        }
        catch (DuplicateRequestException ex)
        {
            // HTTP 409: server da thay clientRequestId nay -> khong tao them lenh
            Console.WriteLine($"  Server tu choi trung lap (409) nhu mong doi: {ex.Message}");
        }
        catch (SsiException ex)
        {
            Console.WriteLine($"  Retry loi: {ex.GetType().Name} [{ex.Code}] {ex.Message}");
        }

        // --- Buoc 5: Huy tat ca lenh vua dat bang CancelBatchOrdersAsync ---
        // Moi item: AccountNo + DUNG MOT trong OrderId / ClientRequestId.
        var cancelItems = placed
            .Where(r => r.Success)
            .Select(r => !string.IsNullOrEmpty(r.OrderId)
                ? new BatchCancelOrderItem(account, OrderId: r.OrderId)
                : new BatchCancelOrderItem(account, ClientRequestId: r.ClientRequestId))
            .DistinctBy(c => c.OrderId ?? c.ClientRequestId)
            .ToList();

        Console.WriteLine($"\n--- Buoc 5: CancelBatchOrdersAsync ({cancelItems.Count} lenh) ---");
        if (cancelItems.Count == 0)
        {
            Console.WriteLine("  Khong co lenh nao dat thanh cong, khong can huy.");
            return;
        }
        try
        {
            var cancelResp = await trading.Trading.CancelBatchOrdersAsync(cancelItems);
            PrintResults("Ket qua huy", cancelResp);
        }
        catch (SsiException ex)
        {
            Console.WriteLine($"  Huy that bai: {ex.GetType().Name} [{ex.Code}] {ex.Message}");
            Console.WriteLine("  !!! Hay vao so lenh va huy tay cac lenh con lai.");
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] placed_ok|placed_fail|cancel_requested");
        Console.WriteLine($"{placed.Count(r => r.Success)}|{placed.Count(r => !r.Success)}|{cancelItems.Count}");
    }
}
