// Sample 1 — Xac thuc va lay Access Token
// Dang nhap, lay token, kiem tra hoat dong bang cach goi API account info.

using SsiSdk;

static class Sample01Auth
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);

        // --- Buoc 1-2: Xac thuc, nhan accessToken + refreshToken ---
        await AuthHelper.EnsureAuthAsync(auth);

        var accessToken = auth.AccessToken;
        Console.WriteLine($"Access Token : {accessToken[..Math.Min(40, accessToken.Length)]}...");

        // --- Buoc 3: Token da duoc SDK luu tu dong ---

        // --- Buoc 4: Kiem tra token het han & refresh ---
        if (auth.TokenManager.IsTokenExpired)
        {
            Console.WriteLine("\nToken het han, dang refresh...");
            var newToken = await auth.RefreshAsync();
            Console.WriteLine($"Token moi    : {newToken.AccessToken[..40]}...");
        }

        // --- Xac nhan token hoat dong bang cach goi API ---
        var trading = new TradingClient(auth);
        var accounts = await trading.Account.GetAccountInfoAsync();
        Console.WriteLine($"\nXac thuc thanh cong! Tim thay {accounts.Count} tai khoan:");
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
