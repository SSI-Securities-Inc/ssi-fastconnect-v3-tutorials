// Sample 1 — Xác thực, Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
// Đăng nhập, xử lý luồng OTP (Smart OTP Push Polling / OTP 6 số) và kiểm tra tài khoản qua Trading API.

using SsiSdk;

static class Sample01Auth
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);

        // --- Bước 1-3: Xác thực + Yêu cầu & Xác thực OTP / Smart OTP ---
        await AuthHelper.EnsureAuthAsync(auth);

        var accessToken = auth.AccessToken;
        Console.WriteLine($"\n--- Thông tin Token ---");
        Console.WriteLine($"Access Token : {(string.IsNullOrEmpty(accessToken) ? "N/A" : accessToken[..Math.Min(40, accessToken.Length)])}...");

        // --- Bước 4: Kiểm tra token hết hạn & refresh ---
        if (auth.TokenManager.IsTokenExpired)
        {
            Console.WriteLine("\nToken hết hạn, đang refresh...");
            var newToken = await auth.RefreshAsync();
            Console.WriteLine($"Token mới    : {newToken.AccessToken[..Math.Min(40, newToken.AccessToken.Length)]}...");
        }

        // --- Bước 5: Xác nhận token hoạt động bằng cách gọi API ---
        var trading = new TradingClient(auth);
        var accounts = await trading.Account.GetAccountInfoAsync();
        Console.WriteLine($"\nXác thực thành công! Tìm thấy {accounts.Count} tài khoản:");
        foreach (var acc in accounts)
        {
            Console.WriteLine($"  - {acc.AccountNo} ({acc.AccountType})");
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] token_type|expires_at|account_count");
        Console.WriteLine($"Bearer|{auth.TokenManager.Token?.ExpiresAt}|{accounts.Count}");
        Console.WriteLine("[Response:account] account_no|account_type");
        foreach (var acc in accounts)
        {
            Console.WriteLine($"{acc.AccountNo}|{acc.AccountType}");
        }
    }
}

