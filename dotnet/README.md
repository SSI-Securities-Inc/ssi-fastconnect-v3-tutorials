# FastConnect .NET SDK Samples

## Cấu hình (`config.json` / `SampleConfig.cs`)

Tự động đọc thông tin từ `config.json` tại thư mục gốc dự án (`../../config.json` hoặc từ `AppContext.BaseDirectory`).

1. Tạo tệp `config.json` tại thư mục gốc dự án (nếu chưa có):
   ```bash
   cp ../config.example.json ../config.json
   ```
2. Điền các tham số API vào `config.json` hoặc sửa trực tiếp trong `dotnet/SampleConfig.cs`.

## Chạy từng Sample

```bash
cd dotnet

# Chay sample cu the
dotnet run -- 01    # Auth & Request/Verify OTP
dotnet run -- 02    # Index List
dotnet run -- 03    # OHLC
dotnet run -- 04    # Securities
dotnet run -- 05    # Balance
dotnet run -- 06    # Limit Order
dotnet run -- 07    # Market Order
dotnet run -- 08    # Order Status
dotnet run -- 09    # Cancel Order
dotnet run -- 10    # WebSocket Data
dotnet run -- 11    # WebSocket Trading
dotnet run -- 12    # MA Cross Auto Trade
dotnet run -- 13    # Lệnh điều kiện FCO
dotnet run -- 14    # Đặt/hủy nhiều lệnh (batch)
dotnet run -- 15    # Vị thế T1/T2 & PPMMR
dotnet run -- 16    # Streaming nâng cao
dotnet run -- 17    # Trading WebSocket (EXPERIMENTAL)
dotnet run -- 18    # Xử lý lỗi
```

Sample 06, 07, 09, 13, 14 đặt lệnh THẬT: chỉ chạy trên tài khoản thử nghiệm.

Khóa config tùy chọn (không bắt buộc, thiếu thì sample tương ứng bỏ qua phần đó): `fco_code` (mã xác nhận cho lệnh FCO, Sample 13), `derivative_account` (tài khoản phái sinh, Sample 15).

## Danh sach sample

| # | File | Mo ta |
|---|------|-------|
| 01 | Sample01Auth.cs | Xac thuc, yeu cau & xac thuc OTP (Push Smart OTP / OTP 6 so), lay token & kiem tra tai khoan |
| 02 | Sample02IndexList.cs | Lay danh sach chi so thi truong |
| 03 | Sample03Ohlc.cs | Lay du lieu K-line (OHLC) nhieu timeframe, lay het cac trang (IterOhlcAsync / GetOhlcAllAsync); 1week/1month khong ho tro |
| 04 | Sample04Securities.cs | Lay danh sach co phieu theo san/index & Master Data |
| 05 | Sample05Balance.cs | So du tai khoan, suc mua, vi the |
| 06 | Sample06LimitOrder.cs | Dat lenh Limit (LO) mua/ban |
| 07 | Sample07MarketOrder.cs | Dat lenh Market (MP) |
| 08 | Sample08OrderStatus.cs | Theo doi trang thai lenh (polling) |
| 09 | Sample09CancelOrder.cs | Huy lenh dang mo |
| 10 | Sample10WebsocketData.cs | WebSocket du lieu thi truong real-time |
| 11 | Sample11WebsocketTrading.cs | WebSocket trang thai lenh real-time |
| 12 | Sample12MaCrossAutoTrade.cs | MA Cross + WebSocket auto trade |
| 13 | Sample13FcoOrder.cs | Dat & quan ly lenh dieu kien FCO (GTD, Stop, Trailing Stop, OCO, Bull Bear), lich su trang thai, `code` OTP tuy chon |
| 14 | Sample14BatchOrders.cs | Dat/huy nhieu lenh mot lan (batch), ket qua tung lenh, idempotency theo clientRequestId |
| 15 | Sample15PositionsPpmmr.cs | Vi the co phieu T1/T2, so du T0/T1/T2, PPMMR Equity & phai sinh (chi doc) |
| 16 | Sample16StreamingAdvanced.cs | Streaming nang cao: OnConnection, handler theo loai, quy tac topic, ping, ListSubscription/UnsubscribeAll |
| 17 | Sample17TradingWs.cs | Trading WebSocket (EXPERIMENTAL, SsiSdk.Experimental): query + su kien lenh, chi doc |
| 18 | Sample18ErrorHandling.cs | Phan loai exception, ServerErrorCode, retry theo RetryAfterMs, 409 duplicate |
