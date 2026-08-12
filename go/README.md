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