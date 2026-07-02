// FastConnect .NET SDK Samples — Entry point
// Chay: dotnet run -- <so_sample>
// VD:   dotnet run -- 01    (Auth)
//       dotnet run -- 10    (WebSocket Data)

var sampleId = args.Length > 0 ? args[0] : "01";

Console.WriteLine($"=== FastConnect .NET SDK — Sample {sampleId} ===\n");

await (sampleId switch
{
    "01" => Sample01Auth.RunAsync(),
    "02" => Sample02IndexList.RunAsync(),
    "03" => Sample03Ohlc.RunAsync(),
    "04" => Sample04Securities.RunAsync(),
    "05" => Sample05Balance.RunAsync(),
    "06" => Sample06LimitOrder.RunAsync(),
    "07" => Sample07MarketOrder.RunAsync(),
    "08" => Sample08OrderStatus.RunAsync(),
    "09" => Sample09CancelOrder.RunAsync(),
    "10" => Sample10WebsocketData.RunAsync(),
    "11" => Sample11WebsocketTrading.RunAsync(),
    "12" => Sample12MaCrossAutoTrade.RunAsync(),
    _ => throw new ArgumentException($"Sample '{sampleId}' khong ton tai. Dung 01-12."),
});
