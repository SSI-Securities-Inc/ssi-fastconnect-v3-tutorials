// Cấu hình kết nối dùng chung cho toàn bộ sample
// Tự động đọc từ config.json ở gốc dự án (nếu có).

using System;
using System.IO;
using System.Linq;
using System.Text.Json;
using SsiSdk;

static class SampleConfig
{
    private static readonly JsonElement JsonConfig = LoadJsonConfig();

    private static JsonElement LoadJsonConfig()
    {
        var path = Path.Combine(AppContext.BaseDirectory, "..", "..", "..", "..", "config.json");
        if (File.Exists(path))
        {
            try
            {
                var txt = File.ReadAllText(path);
                using var doc = JsonDocument.Parse(txt);
                return doc.RootElement.Clone();
            }
            catch (Exception ex)
            {
                Console.WriteLine($"[Config Warning] Không thể đọc config.json: {ex.Message}");
            }
        }
        return default;
    }

    private static string GetVal(string key, string fallback)
    {
        if (JsonConfig.ValueKind == JsonValueKind.Object && JsonConfig.TryGetProperty(key, out var v))
        {
            return v.GetString() ?? fallback;
        }
        return fallback;
    }

    public static Config Create() => new()
    {
        ClientId = GetVal("client_id", "<CLIENT_ID>"),
        ApiKey = GetVal("api_key", "<API_KEY>"),
        ApiSecret = GetVal("api_secret", "<API_SECRET>"),
        PrivateKey = GetVal("private_key", "<PRIVATE_KEY_CONTENT>"),
    };

    // Khoa tuy chon: tra ve null neu config.json khong co hoac de trong
    private static string? GetOptional(string key)
    {
        if (JsonConfig.ValueKind == JsonValueKind.Object && JsonConfig.TryGetProperty(key, out var v)
            && v.ValueKind == JsonValueKind.String && !IsPlaceholder(v.GetString()))
        {
            return v.GetString();
        }
        return null;
    }

    // Gia tri trong hoac con la mau (<...> / YOUR_...) coi nhu chua cau hinh
    private static bool IsPlaceholder(string? value) =>
        string.IsNullOrWhiteSpace(value) || value.StartsWith('<') || value.StartsWith("YOUR_", StringComparison.Ordinal);

    // Fail fast: dung ngay neu config.json thieu khoa bat buoc (con gia tri <PLACEHOLDER>)
    public static void EnsureConfigured()
    {
        var missing = new[]
        {
            ("client_id", GetVal("client_id", "<>")), ("api_key", GetVal("api_key", "<>")),
            ("api_secret", GetVal("api_secret", "<>")), ("private_key", GetVal("private_key", "<>")),
            ("equity_account", GetVal("equity_account", "<>")),
        }.Where(t => IsPlaceholder(t.Item2)).Select(t => t.Item1).ToList();
        if (missing.Count > 0)
            throw new InvalidOperationException($"config.json thieu khoa bat buoc: {string.Join(", ", missing)}");
    }

    // Che so tai khoan khi in ra man hinh: chi giu 3 ky tu cuoi
    public static string MaskAccount(string? accountNo) =>
        string.IsNullOrEmpty(accountNo) ? "(trong)"
        : accountNo.Length <= 3 ? "***" : new string('*', accountNo.Length - 3) + accountNo[^3..];

    public static string AccountNo => GetVal("equity_account", "<ACCOUNT_NO>");
    public static string Otp => GetVal("otp", "<OTP>");

    // Ma OTP gui kem lenh dieu kien (FCO) khi server yeu cau. Dat trong config.json khoa "fco_code"; khong co thi bo qua.
    public static string? FcoCode => GetOptional("fco_code");

    // Tai khoan phai sinh (tuy chon, config.json "derivative_account"); null thi bo qua phan phai sinh
    public static string? DerivativeAccount => GetOptional("derivative_account");
}
