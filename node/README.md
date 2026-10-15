# FastConnect Node.js Samples

Bộ sample minh họa 18 kịch bản sử dụng **FastConnect Node.js SDK** (`@ssi.developer/ssi-sdk`).

> Node.js là async sẵn nên mỗi kịch bản chỉ có **một file** (dùng top-level `await`), không tách sync/async như Python.

## Yêu cầu

- Node.js ≥ 18
- SDK `@ssi.developer/ssi-sdk` (link tới `../../sdk/ssi-sdk-node`)

## Cài đặt & chạy

```bash
cd node
npm install            # cài SDK

# Chạy các sample từ 01 đến 18
npm run sample:01      # Sample 01 — Auth & OTP
npm run sample:02      # Sample 02 — Index list
npm run sample:03      # Sample 03 — OHLC
npm run sample:04      # Sample 04 — Securities
npm run sample:05      # Sample 05 — Balance
npm run sample:06      # Sample 06 — Limit Order
npm run sample:07      # Sample 07 — Market Order
npm run sample:08      # Sample 08 — Order Status
npm run sample:09      # Sample 09 — Cancel Order
npm run sample:10      # Sample 10 — WebSocket Data
npm run sample:11      # Sample 11 — WebSocket Trading
npm run sample:12      # Sample 12 — MA Cross Auto Trade
npm run sample:13      # Sample 13 — Lệnh điều kiện FCO
npm run sample:14      # Sample 14 — Đặt/hủy lệnh theo lô (batch)
npm run sample:15      # Sample 15 — Vị thế & PPMMR (T0/T1/T2)
npm run sample:16      # Sample 16 — Streaming nâng cao (board/index/kết nối)
npm run sample:17      # Sample 17 — Trading WebSocket (EXPERIMENTAL, cần TRADING_WS_EXPERIMENTAL=1)
npm run sample:18      # Sample 18 — Xử lý lỗi, rate limit, timeout

# Hoặc chạy trực tiếp bằng Node
node sample_01_auth.js
```

## Cấu hình (`config.js`)

Tất cả sample import config từ `config.js`. File này **không chứa credential**: nó đọc `../config.json`
ở gốc dự án tutorials (sao chép từ `config.example.json`, file đã được `.gitignore`):

| Khóa trong `config.json` | Ý nghĩa |
|---|---|
| `client_id`, `api_key`, `api_secret` | Credential do SSI cấp |
| `private_key` | RSA key (Base64 XML) dùng ký lệnh |
| `equity_account` / `derivative_account` | Số tài khoản (sample chỉ in dạng che `***1234`) |
| `otp` | Chỉ dùng ở lần authenticate đầu tiên |
| `log_level` | Tùy chọn: DEBUG/INFO/WARNING/ERROR (SDK che secret trong log) |

Thiếu giá trị bắt buộc thì sample dừng ngay (`assertConfigured`) thay vì gửi request với placeholder.
Biến môi trường tùy chọn: `FCO_OTP` (mã OTP một lần cho sample 13), `STREAM_SECONDS` (sample 16),
`BATCH_DUP_DEMO=1` (sample 14), `TRADING_WS_EXPERIMENTAL=1` / `TRADING_WS_PLACE=1` (sample 17).

> **Bảo mật:** không commit `config.json`, `token_cache.json` hay `shared_token.json`. Production nên đưa
> secret/private key vào vault hoặc biến môi trường.

## Token cache

`auth_helper.js` cache token vào `token_cache.json`:

- Lần đầu: `requestOtp()` / `authenticate()` / Polling Smart OTP → lưu token.
- Lần sau: load từ file, tự `refresh()` nếu hết hạn — không cần OTP lại.
- `token_cache.json` đã được `.gitignore` (chứa access/refresh token).

> Chạy `sample:01` trước để tạo token; các sample sau dùng lại từ cache.

## Danh sách sample

| # | File | Mô tả | OTP |
|---|------|-------|-----|
| 01 | `sample_01_auth.js` | Xác thực, Yêu cầu & Xác thực OTP / Push Smart OTP, lấy token | ✔ |
| 02 | `sample_02_index_list.js` | Danh sách chỉ số (VN-Index, HNX-Index...) | – |
| 03 | `sample_03_ohlc.js` | Dữ liệu nến K-line (OHLC) | – |
| 04 | `sample_04_securities.js` | Danh sách cổ phiếu theo sàn | – |
| 05 | `sample_05_balance.js` | Số dư & sức mua tài khoản | ✔ |
| 06 | `sample_06_limit_order.js` | Đặt lệnh giới hạn (LO) | ✔ |
| 07 | `sample_07_market_order.js` | Đặt lệnh thị trường (MTL) | ✔ |
| 08 | `sample_08_order_status.js` | Polling trạng thái lệnh | ✔ |
| 09 | `sample_09_cancel_order.js` | Hủy lệnh chưa khớp | ✔ |
| 10 | `sample_10_websocket_data.js` | Giá real-time (Trade/Quote/Room) | – |
| 11 | `sample_11_websocket_trading.js` | Trạng thái lệnh & danh mục real-time | ✔ |
| 12 | `sample_12_ma_cross_auto_trade.js` | Tự động giao dịch theo tín hiệu MA5 cắt MA10 | ✔ |
| 13 | `sample_13_fco_order.js` | Đặt & quản lý lệnh điều kiện FCO (GTD, Stop, Trailing Stop, OCO, Bull Bear), lịch sử trạng thái, OTP `FCO_OTP` | ✔ |
| 14 | `sample_14_batch_orders.js` | Đặt/hủy lệnh theo lô, `maxBatchOrders`, kết quả từng lệnh, `clientRequestId` chống trùng | ✔ |
| 15 | `sample_15_positions_ppmmr.js` | Vị thế (T1/T2 bán & mua), PPMMR (buy/sell T0/T1/T2, marginRatio, purchasingPower), phái sinh, mua/bán tối đa | ✔ |
| 16 | `sample_16_streaming_advanced.js` | Board/index, tick chỉ số, `onConnection`, `listSubscription`, `unsubscribe*`, sự kiện lệnh | ✔ |
| 17 | `sample_17_trading_ws.js` | **EXPERIMENTAL** — đặt/hủy lệnh & truy vấn qua trading WebSocket, rate limit | ✔ |
| 18 | `sample_18_error_handling.js` | Phân loại lỗi, `RateLimitError.retryAfterMs`, `ServerErrorCode`, timeout `*Ms`, không lộ secret | – |

> **Cảnh báo:** sample 06, 07, 09, 12, 13, 14 (và 17 khi bật `TRADING_WS_PLACE=1`) **gửi lệnh thật** lên
> tài khoản trong `config.json`. Chỉ chạy trên tài khoản thử nghiệm hoặc khi hiểu rõ rủi ro.
>
> Các sample WebSocket (10, 11, 12, 16) cần ping định kỳ (SDK không tự ping) và tự dừng sau thời gian
> chờ hoặc khi nhấn `Ctrl+C`. Sample 17 dùng API **experimental** (server chưa công bố phát hành), mặc định thoát ngay.
> Sample 03 dùng `getOhlcAll`/`iterOhlc` để lấy hết các trang; khung `1w`/`1M` không được hỗ trợ.

## API SDK dùng trong sample

| Class | Mục đích | Dùng ở sample |
|---|---|---|
| `Auth` | Xác thực, quản lý token | tất cả |
| `Data` → `marketData` | Index, OHLC, securities, master data | 02, 03, 04, 12, 14, 18 |
| `Trading` → `account` / `trading` / `portfolio` | Tài khoản, đặt/hủy lệnh (kể cả batch), số dư, vị thế, PPMMR, sổ lệnh | 01, 05–09, 12–15, 18 |
| `Stream` → `streaming` | WebSocket real-time | 10, 11, 12, 16 |
| `tradingWs` (`@ssi.developer/ssi-sdk/experimental`) | Trading WebSocket (thử nghiệm) | 17 |
