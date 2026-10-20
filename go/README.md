# FastConnect Go SDK Samples

## Cấu hình (`config.json`)

Các file sample trong thư mục Go tự động đọc cấu hình từ `../config.json` hoặc `config.json`.

1. Tạo tệp `config.json` tại thư mục gốc dự án (nếu chưa có):
   ```bash
   cp ../config.example.json ../config.json
   ```
2. Điền các thông số kết nối API (`client_id`, `api_key`, `api_secret`, `private_key`, `equity_account`, `otp`) vào `config.json`.

## Hướng dẫn chạy từng Sample

```bash
cd go

# Sample 01 — Auth & Request/Verify OTP
go run sample_01_auth.go

# Sample 2 — Index list
go run sample_02_index_list.go

# Sample 3 — OHLC
go run sample_03_ohlc.go

# Sample 4 — Securities
go run sample_04_securities.go

# Sample 5 — Balance
go run sample_05_balance.go

# Sample 6 — Limit order
go run sample_06_limit_order.go

# Sample 7 — Market order
go run sample_07_market_order.go

# Sample 8 — Order status
go run sample_08_order_status.go

# Sample 9 — Cancel order
go run sample_09_cancel_order.go

# Sample 10 — WebSocket market data
go run sample_10_websocket_data.go

# Sample 11 — WebSocket trading
go run sample_11_websocket_trading.go

# Sample 12 — MA cross auto trade
go run sample_12_ma_cross_auto_trade.go

# Sample 13 — FCO order
go run sample_13_fco_order.go

# Sample 14 — Batch orders (GỬI LỆNH THẬT, hỏi xác nhận YES trước khi gửi)
go run sample_14_batch_orders.go

# Sample 15 — Positions & PPMMR (T0/T1/T2, chỉ đọc)
go run sample_15_positions_ppmmr.go

# Sample 16 — Streaming nâng cao (board, index, sự kiện kết nối, ListSubscription)
go run sample_16_streaming_advanced.go

# Sample 17 — Trading WebSocket (GỬI LỆNH THẬT, hỏi xác nhận YES)
go run sample_17_trading_ws.go

# Sample 18 — Xử lý lỗi (rate limit, Smart OTP, validation, ServerErrorCode)
go run sample_18_error_handling.go

# Sample 19 — Sự kiện doanh nghiệp (cổ tức, cổ phiếu thưởng, quyền mua; chỉ đọc)
go run sample_19_corporate_actions.go
```

Lưu ý:

- Sample 14 và 17 gửi lệnh mua thật ở giá sàn rồi hủy ngay; chỉ chạy trên tài khoản bạn được phép giao dịch. Cả hai cần `private_key` và `equity_account`.
- Sample 17 dùng package `tradingws` (`github.com/SSI-Securities-Inc/ssi-sdk-go/v3/tradingws`).
- Mọi thời gian trong SDK 3.3.0 tính bằng mili-giây (`*Ms`), ví dụ `Stream.Wait(&timeoutMs)`.
- Các sample thiếu `client_id`, `api_key`, `api_secret` sẽ dừng ngay, không có giá trị mặc định.