// Sample 8 — Kiem tra trang thai lenh
// Theo doi tien trinh khop cua mot lenh cu the.

using SsiSdk;

public static class Sample09OrderStatus
{
    private static readonly HashSet<string> TerminalStatuses = new()
    {
        OrderStatus.Filled,
        OrderStatus.Cancelled,
        OrderStatus.Rejected,
        OrderStatus.Expired,
        OrderStatus.PartialCancelled,
    };

    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var trading = new TradingClient(auth);

        // --- Buoc 1: Dat mot lenh de theo doi ---
        Console.WriteLine("Dat lenh Limit mua SSI @ 26000...");
        var result = await trading.Trading.PlaceLimitOrderAsync(
            SampleConfig.AccountNo, "SSI", OrderSide.Buy, 100, 26000);
        Console.WriteLine($"  Ket qua dat lenh: OrderId={result.OrderId} Status={result.Status}");

        // --- Buoc 2-5: Polling trang thai ---
        Console.WriteLine("\n--- Bat dau theo doi trang thai lenh ---");
        const int maxPolls = 10;
        const int pollIntervalMs = 3000;

        for (var i = 1; i <= maxPolls; i++)
        {
            var orders = await trading.Portfolio.GetTodayOrdersAsync(SampleConfig.AccountNo);

            if (orders.Count == 0)
            {
                Console.WriteLine($"  Poll {i}: Chua co lenh trong so.");
                await Task.Delay(pollIntervalMs);
                continue;
            }

            var latest = orders[^1];
            var remaining = latest.Quantity - latest.FilledQuantity - latest.CancelQuantity;

            Console.WriteLine(
                $"  Poll {i}: OrderID={latest.OrderId} | " +
                $"Status={latest.Status} | " +
                $"Khop={latest.FilledQuantity}/{latest.Quantity} | " +
                $"Con lai={remaining}");

            if (TerminalStatuses.Contains(latest.Status))
            {
                Console.WriteLine($"\n-> Lenh da ket thuc voi trang thai: {latest.Status}");
                if (latest.FilledQuantity > 0)
                {
                    Console.WriteLine($"  Da khop: {latest.FilledQuantity} co phieu @ trung binh {latest.AvgPrice:N0}");
                }
                return;
            }

            await Task.Delay(pollIntervalMs);
        }

        Console.WriteLine($"\nHet {maxPolls} lan poll, lenh van dang mo.");
    }
}
