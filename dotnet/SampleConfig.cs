// Cau hinh ket noi dung chung cho toan bo sample (moi truong UAT)

using SsiSdk;

static class SampleConfig
{
    public static Config Create() => new()
    {
        ClientId = "<CLIENT_ID>",
        ApiKey = "<API_KEY>",
        ApiSecret = "<API_SECRET>",
        PrivateKey = "<PRIVATE_KEY_CONTENT>",
    };

    public const string AccountNo = "<ACCOUNT_NO>";
    public const string Otp = "<OTP>";
}
