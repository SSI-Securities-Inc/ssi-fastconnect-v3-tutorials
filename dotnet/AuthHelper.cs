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

        if (cached is not null)
        {
            auth.TokenManager.SetToken(cached);
            if (!auth.TokenManager.IsTokenExpired)
            {
                Console.WriteLine("Token con han, dung token tu file.");
                return;
            }

            Console.WriteLine("Token da het han, dang refresh...");
            try
            {
                cached = await auth.RefreshAsync();
                SaveToken(cached);
                Console.WriteLine("Refresh token thanh cong.");
                return;
            }
            catch (Exception ex)
            {
                Console.WriteLine($"Refresh token that bai ({ex.Message}), tien hanh xac thuc lai...");
            }
        }

        Console.WriteLine("Khong tim thay token hop le, dang thuc hien quy trinh xac thuc & OTP...");
        if (!string.IsNullOrEmpty(otp))
        {
            cached = await auth.AuthenticateAsync(otp);
        }
        else
        {
            Console.WriteLine("=== Yeu cau OTP (Request OTP) ===");
            var otpRes = await auth.RequestOtpAsync();
            string? transactionId = null;
            if (otpRes.ValueKind == JsonValueKind.Object)
            {
                if (otpRes.TryGetProperty("data", out var dataObj) && dataObj.ValueKind == JsonValueKind.Object)
                {
                    if (dataObj.TryGetProperty("transactionId", out var tid))
                        transactionId = tid.GetString();
                }
                if (string.IsNullOrEmpty(transactionId) && otpRes.TryGetProperty("transactionId", out var tid2))
                {
                    transactionId = tid2.GetString();
                }
            }

            if (!string.IsNullOrEmpty(transactionId))
            {
                Console.WriteLine($"[Smart OTP] Transaction ID: {transactionId}");
                Console.WriteLine("Vui long mo ung dung SSI tren dien thoai va bam APPROVE (Xac nhan)...");
                Console.WriteLine("SDK dang Polling cho ban bam phe duyet...");
                var accessToken = await auth.EnsureAuthenticatedAsync(
                    transactionId: transactionId,
                    pollInterval: TimeSpan.FromSeconds(5),
                    pollMaxRetries: 6
                );
                cached = auth.TokenManager.Token;
            }
            else
            {
                Console.Write("Vui long nhap ma OTP 6 so: ");
                var userOtp = Console.ReadLine()?.Trim() ?? "";
                if (!string.IsNullOrEmpty(userOtp))
                {
                    cached = await auth.AuthenticateAsync(otp: userOtp);
                }
            }
        }

        if (auth.TokenManager.Token is not null)
        {
            SaveToken(auth.TokenManager.Token);
            Console.WriteLine("Authenticate thanh cong, token da luu.");
        }
    }
}
