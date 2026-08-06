using System.Text.Json;
using SsiSdk;

/// <summary>
/// Sample 02 — Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
/// </summary>
static class Sample02Otp
{
    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);

        Console.WriteLine("=== Bước 1: Yêu cầu OTP (Request OTP) ===");
        var otpRes = await auth.RequestOtpAsync();
        Console.WriteLine($"Request OTP Response: {otpRes}");

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
            Console.WriteLine($"\n[Smart OTP] Đã nhận Transaction ID: {transactionId}");
            Console.WriteLine("Vui lòng mở ứng dụng SSI trên điện thoại và bấm APPROVE (Xác nhận)...");
            Console.WriteLine("SDK đang Polling chờ bạn bấm phê duyệt...");

            try
            {
                var accessToken = await auth.EnsureAuthenticatedAsync(
                    transactionId: transactionId,
                    pollInterval: TimeSpan.FromSeconds(5),
                    pollMaxRetries: 6
                );
                Console.WriteLine("\n[THÀNH CÔNG] Đã xác thực Smart OTP!");
                Console.WriteLine($"Access Token: {accessToken[..Math.Min(40, accessToken.Length)]}...");
            }
            catch (Exception ex)
            {
                Console.WriteLine($"\n[LỖI/TIMEOUT] Phê duyệt Smart OTP thất bại: {ex.Message}");
            }
        }
        else
        {
            Console.Write("\n[OTP Thường / Smart OTP lấy trực tiếp trên App] Vui lòng nhập mã OTP 6 số: ");
            var userOtp = Console.ReadLine()?.Trim() ?? "";
            if (!string.IsNullOrEmpty(userOtp))
            {
                var token = await auth.AuthenticateAsync(otp: userOtp);
                Console.WriteLine("\n[THÀNH CÔNG] Đã xác thực mã OTP!");
                Console.WriteLine($"Access Token: {token.AccessToken[..Math.Min(40, token.AccessToken.Length)]}...");
            }
        }
    }
}
