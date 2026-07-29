// Token cache de tai su dung token giua cac lan chay
// Load tu file -> neu het han thi refresh -> neu chua co thi authenticate

using System.Text.Json;
using SsiSdk;
using SsiSdk.Models;

static class AuthHelper
{
    private static readonly string[] SearchTokenFiles = new[]
    {
        Path.Combine(AppContext.BaseDirectory, "..", "..", "..", "token_cache.json"),
        Path.Combine(Directory.GetCurrentDirectory(), "token_cache.json"),
        Path.Combine(Directory.GetCurrentDirectory(), "..", "python", "token_cache.json"),
        Path.Combine(Directory.GetCurrentDirectory(), "..", "go", "token_cache.json"),
        Path.Combine(Directory.GetCurrentDirectory(), "..", "node", "token_cache.json"),
        Path.Combine(Directory.GetCurrentDirectory(), "..", "shared_token.json")
    };

    private static Token? LoadToken()
    {
        foreach (var file in SearchTokenFiles)
        {
            if (!File.Exists(file)) continue;
            try
            {
                using var doc = JsonDocument.Parse(File.ReadAllText(file));
                var r = doc.RootElement;
                var accessToken = r.TryGetProperty("accessToken", out var at) ? at.GetString() ?? "" : "";
                if (string.IsNullOrEmpty(accessToken)) continue;
                return new Token
                {
                    AccessToken = accessToken,
                    TokenType = r.TryGetProperty("tokenType", out var tt) ? tt.GetString() ?? "Bearer" : "Bearer",
                    ExpiresAt = r.TryGetProperty("expiresAt", out var ea) ? ea.GetInt64() : 0,
                    RefreshToken = r.TryGetProperty("refreshToken", out var rt) ? rt.GetString() ?? "" : "",
                    RefreshTokenExpiresAt = r.TryGetProperty("refreshExpiresAt", out var rea) ? rea.GetInt64() : 0,
                };
            }
            catch { continue; }
        }
        return null;
    }

    private static void SaveToken(Token token)
    {
        var dict = new Dictionary<string, object>
        {
            ["accessToken"] = token.AccessToken,
            ["tokenType"] = token.TokenType,
            ["expiresAt"] = token.ExpiresAt,
            ["refreshToken"] = token.RefreshToken,
            ["refreshExpiresAt"] = token.RefreshTokenExpiresAt,
        };
        var json = JsonSerializer.Serialize(dict, new JsonSerializerOptions { WriteIndented = true });
        var targetFile = Path.Combine(Directory.GetCurrentDirectory(), "token_cache.json");
        File.WriteAllText(targetFile, json);
        Console.WriteLine($"Token da luu vao {targetFile}");
    }

    public static async Task EnsureAuthAsync(AuthClient auth, string? otp = null)
    {
        var cached = LoadToken();

        if (cached is null)
        {
            Console.WriteLine("Khong tim thay file token, dang authenticate...");
            cached = string.IsNullOrEmpty(otp)
                ? await auth.AuthenticateAsync()
                : await auth.AuthenticateAsync(otp);
            SaveToken(cached);
            Console.WriteLine("Authenticate thanh cong, token da luu.");
            return;
        }

        auth.TokenManager.SetToken(cached);

        if (auth.TokenManager.IsTokenExpired)
        {
            Console.WriteLine("Token da het han, dang refresh...");
            cached = await auth.RefreshAsync();
            SaveToken(cached);
            Console.WriteLine("Refresh token thanh cong.");
        }
        else
        {
            Console.WriteLine("Token con han, dung token tu file.");
        }
    }
}
