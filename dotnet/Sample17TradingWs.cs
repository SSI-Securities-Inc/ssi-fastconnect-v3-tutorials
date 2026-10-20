// Sample 17 — Trading WebSocket
// Truy van va nhan su kien lenh qua mot socket giao dich rieng (khong phai REST, khong phai StreamClient).
//
// Sample nay CHI DOC (query + subscribe su kien), khong dat/huy lenh.
// Ket noi bang token hien co, nen can xac thuc truoc.

using SsiSdk;

public static class Sample17TradingWs
{
    // Chi liet ke ten truong, khong in noi dung (co the chua so tai khoan)
    private static string Shape(TradingWsResponse r)
    {
        var keys = r.Result.ValueKind == System.Text.Json.JsonValueKind.Object
            ? string.Join(",", r.Result.EnumerateObject().Select(p => p.Name).Take(8))
            : r.Result.ValueKind.ToString();
        return $"status={r.Status} method={r.Method} result=[{keys}]";
    }

    public static async Task RunAsync()
    {
        SampleConfig.EnsureConfigured();
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var account = SampleConfig.AccountNo;
        Console.WriteLine($"Trading WS | tai khoan {SampleConfig.MaskAccount(account)}\n");

        var options = new TradingWsOptions
        {
            RequestTimeoutMs = 10_000,   // mili-giay
            PingIntervalMs = 30_000,
            RespectRateLimit = true,     // SDK tu cho khi server bao het han muc, toi da MaxRateLimitWaitMs
            MaxRateLimitWaitMs = 2_000,
        };

        try
        {
            // --- Buoc 1: Mo socket ---
            await using var ws = await TradingWs.ConnectAsync(auth, options);

            var events = 0;
            ws.Client.OnEvent = _ => Interlocked.Increment(ref events);
            Console.WriteLine($"Da ket noi: {ws.Client.IsConnected}");

            // --- Buoc 2: Truy van (CHI DOC) ---
            Console.WriteLine("\n--- Buoc 2: Query ---");
            Console.WriteLine($"  accountBalance : {Shape(await ws.QueryAccountBalanceAsync(account))}");
            Console.WriteLine($"  position       : {Shape(await ws.QueryPositionAsync(account))}");
            Console.WriteLine($"  maxBuySell SSI : {Shape(await ws.QueryMaxBuySellAsync(account, "SSI"))}");
            Console.WriteLine($"  ppmmrAccount   : {Shape(await ws.QueryPpmmrAccountAsync(account))}");

            // --- Buoc 3: Nhan su kien lenh cua dung 1 tai khoan ---
            // Socket nay tu choi wildcard va danh sach: moi lan mot tai khoan
            Console.WriteLine("\n--- Buoc 3: Subscribe su kien lenh 30s ---");
            await ws.SubscribeOrderEventsAsync(account);
            await Task.Delay(30_000);
            await ws.UnsubscribeOrderEventsAsync(account);
            Console.WriteLine($"  Nhan {events} su kien");

            // Lenh that qua socket nay (PlaceOrderAsync / AmendOrderAsync / CancelOrderAsync / BatchNewOrdersAsync ...)
            // co san nhung KHONG duoc goi trong sample nay.
            Console.WriteLine("\n[Response] trading_ws_ok");
        }
        catch (TradingWsError ex)
        {
            // Server tu choi lenh hoac socket: ErrorCode vi du 401006/401008/401013 la loi chu ky
            Console.WriteLine($"Trading WS bi tu choi: status={ex.Status} code={ex.ErrorCode} {ex.ServerMessage}");
        }
        catch (SsiException ex)
        {
            Console.WriteLine($"Loi: {ex.GetType().Name} [{ex.Code}] {ex.Message}");
        }
    }
}
