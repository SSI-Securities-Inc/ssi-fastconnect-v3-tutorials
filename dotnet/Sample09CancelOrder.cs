// Sample 09 — Hủy lệnh
// Dung phan khoi luong chua khop cua lenh dang mo.

using SsiSdk;

public static class Sample09CancelOrder
{
    private static readonly HashSet<string> CancellableStatuses = new()
    {
        OrderStatus.PendingApproval,
        OrderStatus.Ready,
        OrderStatus.Sent,
        OrderStatus.Queued,
        OrderStatus.PartialFilled,
    };

    public static async Task RunAsync()
    {
        var config = SampleConfig.Create();
        using var auth = new AuthClient(config);
        await AuthHelper.EnsureAuthAsync(auth, SampleConfig.Otp);

        var trading = new TradingClient(auth);

        // --- Buoc 1: Lay so lenh, tim lenh dang mo ---
        Console.WriteLine("--- So lenh hom nay ---");
        var orders = await trading.Portfolio.GetTodayOrdersAsync(SampleConfig.AccountNo);

        var openOrders = orders.Where(o => CancellableStatuses.Contains(o.Status)).ToList();
        Console.WriteLine($"Tong lenh: {orders.Count} | Lenh dang mo: {openOrders.Count}\n");

        if (openOrders.Count == 0)
        {
            Console.WriteLine("Khong co lenh nao dang mo de huy.");

            // --- Response Summary ---
            Console.WriteLine("\n[Response] open_count|cancel_status");
            Console.WriteLine("0|N/A");
            return;
        }

        foreach (var order in openOrders)
        {
            var remaining = order.Quantity - order.FilledQuantity;
            Console.WriteLine(
                $"  OrderID: {order.OrderId} | {order.Symbol} {order.Side} " +
                $"{order.OrderType} | SL: {order.Quantity} @ {order.Price} " +
                $"| Khop: {order.FilledQuantity} | Con: {remaining} " +
                $"| Status: {order.Status}");
        }

        // --- Buoc 2: Huy lenh dau tien trong danh sach ---
        var target = openOrders[0];
        Console.WriteLine($"\n--- Huy lenh: {target.OrderId} ---");

        var result = await trading.Trading.CancelOrderByOrderIdAsync(
            SampleConfig.AccountNo, target.OrderId);
        Console.WriteLine($"  Ket qua huy: OrderId={result.OrderId} Status={result.Status}");

        // --- Buoc 3: Xac nhan trang thai sau huy ---
        Console.WriteLine("\n--- Kiem tra so lenh sau huy ---");
        var ordersAfter = await trading.Portfolio.GetTodayOrdersAsync(SampleConfig.AccountNo);
        foreach (var order in ordersAfter)
        {
            if (order.OrderId == target.OrderId)
            {
                Console.WriteLine(
                    $"  OrderID: {order.OrderId} | " +
                    $"Status: {order.Status} | " +
                    $"Khop: {order.FilledQuantity} | " +
                    $"Da huy: {order.CancelQuantity}");
                break;
            }
        }

        // --- Response Summary ---
        Console.WriteLine("\n[Response] open_count|cancel_status");
        Console.WriteLine($"{openOrders.Count}|{result.Status}");

        // --- Buoc 4: Cap nhat lai so du ---
        Console.WriteLine("\n--- So du sau huy ---");
        var balance = await trading.Portfolio.GetEquityBalanceAsync(SampleConfig.AccountNo);
        if (balance is not null)
        {
            Console.WriteLine($"  Tien mat kha dung: {balance.AccountBalance,15:N0}");
        }
    }
}
