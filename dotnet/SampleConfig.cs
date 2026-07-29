// Cấu hình kết nối dùng chung cho toàn bộ sample
// Tự động đọc từ config.json ở gốc dự án (nếu có).

using System;
using System.IO;
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

    public static string AccountNo => GetVal("equity_account", "<ACCOUNT_NO>");
    public static string Otp => GetVal("otp", "<OTP>");
}
