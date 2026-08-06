// FastConnect .NET SDK Samples — Entry point
// Chay: dotnet run -- <so_sample>
// VD:   dotnet run -- 01    (Auth)
//       dotnet run -- 10    (WebSocket Data)

var sampleId = args.Length > 0 ? args[0] : "01";

Console.WriteLine($"=== FastConnect .NET SDK — Sample {sampleId} ===\n");

await (sampleId switch
{
    "01" => Sample01Auth.RunAsync(),
    "02" => Sample02Otp.RunAsync(),
    "03" => Sample03IndexList.RunAsync(),
    "04" => Sample04Ohlc.RunAsync(),
    "05" => Sample05Securities.RunAsync(),
    "06" => Sample06Balance.RunAsync(),
    "07" => Sample07LimitOrder.RunAsync(),
    "08" => Sample08MarketOrder.RunAsync(),
    "09" => Sample09OrderStatus.RunAsync(),
    "10" => Sample10CancelOrder.RunAsync(),
    "11" => Sample11WebsocketData.RunAsync(),
    "12" => Sample12WebsocketTrading.RunAsync(),
    "13" => Sample13MaCrossAutoTrade.RunAsync(),
    "14" => Sample14FcoOrder.RunAsync(),
    _ => throw new ArgumentException($"Sample '{sampleId}' khong ton tai. Dung 01-14."),
});
