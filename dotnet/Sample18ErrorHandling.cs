// Sample 18 — Xử lý lỗi
//   1. Cay exception cua SDK va cach phan loai (SsiException -> ...)
//   2. Ma loi server (ServerErrorCode) va Code/StatusCode/Headers
//   3. Loi nhap lieu duoc chan ngay tai client (ValidationException)
//   4. Retry an toan: RateLimitException cho theo RetryAfterMs; DuplicateRequestException = lenh da duoc nhan
//   5. Mot loi that tu server (doc, khong tac dong)
//
// Buoc 1-4 chay offline bang exception gia lap; chi buoc 5 goi mang (doc du lieu).

using SsiSdk;

public static class Sample18ErrorHandling
{
    // Phan loai mot loi thanh hanh dong. Thu tu case: lop con truoc, lop cha sau.
    private static string Classify(SsiException ex) => ex switch
    {
        SmartOtpPendingException => "cho nguoi dung bam duyet Smart OTP -> goi lai cung transactionId",
        SmartOtpRejectedException => "Smart OTP bi tu choi/khong tim thay -> bat dau lai quy trinh OTP",
        ReauthenticationRequiredException => "refresh token het hieu luc -> xac thuc lai voi OTP moi",
        AuthenticationException => "loi xac thuc -> kiem tra credential / token",
        DuplicateRequestException => "clientRequestId da duoc server nhan (409) -> KHONG gui lai, tra cuu lenh theo id",
        RateLimitException r => $"vuot han muc -> cho {r.RetryAfterMs?.ToString() ?? "(khong co Retry-After)"} ms roi thu lai",
        ValidationException => "tham so sai tai client -> sua input, khong retry",
        TradingWsError w => $"trading WS tu choi {w.ErrorCode} -> xem ma loi",
        WebSocketException => "loi ket noi WebSocket -> SDK tu reconnect, theo doi OnConnection",
        ApiException => "server tra loi loi -> doc Code/StatusCode, khong retry mu quang",
        _ => "loi SDK khac",
    };

    private static void Describe(SsiException ex)
    {
        Console.WriteLine($"  {ex.GetType().Name,-34} code={ex.Code,-9} http={ex.StatusCode}");
        Console.WriteLine($"    -> {Classify(ex)}");
        // Ma loi server la so: doi chieu voi ServerErrorCode
        if (int.TryParse(ex.Code, out var code))
        {
            var note = ServerErrorCode.IsKnown(code)
                ? code switch
                {
                    ServerErrorCode.FromAfterTo => "fromDate sau toDate",
                    ServerErrorCode.TimeframeNotSupported => "timeframe khong ho tro (1w/1M)",
                    ServerErrorCode.DateFormatInvalid => "sai dinh dang ngay (yyyy/MM/dd HH:mm:ss)",
                    ServerErrorCode.SignatureMismatch => "chu ky khong khop",
                    ServerErrorCode.ReplayWindowExceeded => "request qua cu (lech dong ho?)",
                    ServerErrorCode.ModifyPriceAndQuantity => "chi sua gia HOAC so luong",
                    _ => "ma da biet trong ServerErrorCode",
                }
                : "ma chua co trong ServerErrorCode";
            Console.WriteLine($"    -> ServerErrorCode: {note}");
        }
    }

    // Retry: chi retry loi tam thoi; lenh phai mang clientRequestId co dinh de retry khong tao lenh thu hai
    private static async Task<T> WithRetryAsync<T>(Func<Task<T>> operation, int maxAttempts = 3)
    {
        for (var attempt = 1; ; attempt++)
        {
            try
            {
                return await operation();
            }
            catch (RateLimitException ex) when (attempt < maxAttempts)
            {
                var waitMs = (int)Math.Min(ex.RetryAfterMs ?? 1000 * attempt, 10_000);
                Console.WriteLine($"    lan {attempt}: bi gioi han toc do, cho {waitMs} ms");
                await Task.Delay(waitMs);
            }
        }
    }

    public static async Task RunAsync()
    {
        // --- Buoc 1+2: Phan loai exception (gia lap, khong goi mang) ---
        Console.WriteLine("--- Buoc 1: Phan loai exception ---");
        SsiException[] samples =
        [
            new SmartOtpPendingException("pending"),
            new SmartOtpRejectedException("rejected", code: "401115"),
            new ReauthenticationRequiredException("refresh het han", code: "401103"),
            new DuplicateRequestException("trung clientRequestId"),
            new RateLimitException("qua nhieu request", retryAfterMs: 1500),
            new ValidationException("symbol rong"),
            new ApiException("khong ho tro timeframe", code: "400210", statusCode: 400),
            new ApiException("ma la", code: "999999", statusCode: 400),
        ];
        foreach (var ex in samples) Describe(ex);

        // --- Buoc 2: ValidationException tai client ---
        Console.WriteLine("\n--- Buoc 2: Chan loi nhap lieu truoc khi gui mang ---");
        var checks = new (string Name, Func<object?> Probe)[]
        {
            ("Topic sai interval", () => Topic.Build("trade", "VNM", "1w")),
            ("Topic danh sach dau phay", () => Topic.Build("trade", "VNM,FPT")),
            ("Interval chi hop le cho trade", () => Topic.Build("quote", "VN30", "1m")),
        };
        foreach (var (name, probe) in checks)
        {
            try
            {
                probe();
                Console.WriteLine($"  {name}: (khong mong doi) khong loi");
            }
            catch (ValidationException ex)
            {
                Console.WriteLine($"  {name}: ValidationException - {ex.Message}");
            }
        }

        // --- Buoc 4: Retry an toan (gia lap) ---
        Console.WriteLine("\n--- Buoc 3: Retry khi bi gioi han toc do ---");
        var calls = 0;
        var result = await WithRetryAsync(() =>
        {
            calls++;
            if (calls < 3) throw new RateLimitException("gia lap", retryAfterMs: 200);
            return Task.FromResult($"thanh cong o lan {calls}");
        });
        Console.WriteLine($"  {result}");

        Console.WriteLine("\n--- Buoc 4: Duplicate = da nhan, khong gui lai ---");
        var clientRequestId = "demo-" + Guid.NewGuid().ToString("N")[..8];
        try
        {
            // Mo phong: lan retry thu hai bi server tra 409 vi lan dau thuc ra da toi noi
            throw new DuplicateRequestException($"clientRequestId {clientRequestId} da ton tai");
        }
        catch (DuplicateRequestException)
        {
            Console.WriteLine($"  Coi lenh {clientRequestId} la da gui; tra cuu bang GetTodayOrdersAsync, KHONG dat lai.");
        }

        // --- Buoc 5: Mot loi that tu server (chi doc) ---
        Console.WriteLine("\n--- Buoc 5: Loi that tu server (doc so du tai khoan khong ton tai) ---");
        SampleConfig.EnsureConfigured();
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);
        var trading = new TradingClient(auth);
        try
        {
            await trading.Portfolio.GetEquityBalanceAsync("0000000");
            Console.WriteLine("  (server khong bao loi)");
        }
        catch (SsiException ex)
        {
            Describe(ex);
        }

        Console.WriteLine("\n[Response] error_handling_demo_done");
    }
}
