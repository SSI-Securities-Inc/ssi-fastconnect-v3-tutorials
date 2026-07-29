# FastConnect Node.js Samples

Bộ sample minh họa 13 kịch bản sử dụng **FastConnect Node.js SDK** (`@ssi.developer/ssi-sdk`).

> Node.js là async sẵn nên mỗi kịch bản chỉ có **một file** (dùng top-level `await`), không tách sync/async như Python.

## Yêu cầu

- Node.js ≥ 18
- SDK `@ssi.developer/ssi-sdk` (link tới `../../sdk/ssi-sdk-node`)

## Cài đặt & chạy

```bash
cd node
npm install            # cài SDK qua link local

npm run sample:01      # Auth
npm run sample:02      # Index list
# ...
npm run sample:12      # MA cross auto trade
npm run sample:13      # Lệnh điều kiện FCO

# Hoặc chạy trực tiếp
node sample_01_auth.js
```

## Cấu hình (`config.js`)

Tất cả sample import config từ `config.js` (thay vì nhúng lại trong từng file). Thông tin UAT
được nhúng trực tiếp trong code giống sample python/go — chạy được ngay, không cần set env var:

| Tham số | Giá trị |
|---|---|
| `clientId` | `<CLIENT_ID>` |
| `apiKey` / `apiSecret` | nhúng trong `config.js` |
| `privateKey` | RSA key (Base64 XML) nhúng trong `config.js` — dùng ký lệnh (sample 05–09, 11–13) |
| `ACCOUNT_NO` | `<ACCOUNT_NO>` |
| `OTP` | `<OTP>` (chỉ dùng ở lần authenticate đầu tiên) |

> **Bảo mật:** đây là credential **UAT/sandbox**. Khi chuyển sang production, thay toàn bộ bằng thông
> tin thật do SSI cấp và nên đưa secret/private key vào vault hoặc biến môi trường thay vì commit vào code.

## Token cache

`auth_helper.js` cache token vào `token_cache.json`:

- Lần đầu: `authenticate()` (có thể cần OTP) → lưu token.
- Lần sau: load từ file, tự `refresh()` nếu hết hạn — không cần OTP lại.
- `token_cache.json` đã được `.gitignore` (chứa access/refresh token).

> Chạy `sample:01` trước để tạo token; các sample sau dùng lại từ cache.

## Danh sách sample

| # | File | Mô tả | OTP |
|---|------|-------|-----|
| 01 | `sample_01_auth.js` | Xác thực, lấy access token | – |
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
| 13 | `sample_13_fco_order.js` | Đặt & quản lý lệnh điều kiện FCO (GTD, Stop, Trailing Stop, OCO, Bull Bear) | ✔ |

> Sample 05–09, 11–13 ký lệnh bằng RSA (dùng `privateKey` trong `config.js`). Các sample WebSocket
> (10–12) tự dừng sau 5 phút hoặc khi nhấn `Ctrl+C`.

## API SDK dùng trong sample

| Class | Mục đích | Dùng ở sample |
|---|---|---|
| `Auth` | Xác thực, quản lý token | tất cả |
| `Data` → `marketData` | Index, OHLC, securities | 02, 03, 04, 12 |
| `Trading` → `account` / `trading` / `portfolio` | Tài khoản, đặt/hủy lệnh, số dư, sổ lệnh | 01, 05–09, 12 |
| `Stream` → `streaming` | WebSocket real-time | 10, 11, 12 |
