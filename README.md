# SSI FastConnect API Sample Codes

Tài liệu hướng dẫn và mã nguồn mẫu tích hợp hệ thống giao dịch chứng khoán của SSI (Saigon Securities Incorporation) qua **SSI FastConnect API**.

### Ngôn ngữ hỗ trợ
- [x] **Go** (bao gồm các mẫu từ Sample 01 đến 13)
- [x] **Python** (bao gồm cả phiên bản đồng bộ - synchronous và bất đồng bộ - asynchronous)
- [x] **Node.js** (bao gồm các mẫu từ Sample 01 đến 13)
- [x] **.NET** (bao gồm các mẫu từ Sample 01 đến 13)

## Mục lục
- [Giới thiệu](#giới-thiệu)
- [Tính năng cốt lõi](#tính-năng-cốt-lõi)
- [Hướng dẫn cấu hình](#hướng-dẫn-cấu-hình)
- [Cách chạy mẫu Go](#cách-chạy-mẫu-go)
- [Cách chạy mẫu Python](#cách-chạy-mẫu-python)
- [Cách chạy mẫu Node.js](#cách-chạy-mẫu-nodejs)
- [Cách chạy mẫu .NET](#cách-chạy-mẫu-dotnet)
- [Luồng hoạt động chính](#luồng-hoạt-động-chính)
- [Lưu ý quan trọng](#lưu-ý-quan-trọng)

---

## Giới thiệu
Bộ mã nguồn này cung cấp các ví dụ thực tế giúp nhà phát triển nhanh chóng nắm bắt cách thức giao tiếp với cổng API của SSI để thực hiện:
* Truy vấn thông tin tài khoản và số dư.
* Lấy dữ liệu thị trường (chỉ số, thông tin cổ phiếu, dữ liệu nến lịch sử và intraday).
* Thực hiện giao dịch (đặt lệnh LO, lệnh MP, hủy lệnh, theo dõi trạng thái lệnh, đặt & quản lý lệnh điều kiện FCO).
* Đăng ký nhận dữ liệu thời gian thực (real-time stream) qua kết nối WebSocket bảo mật.
* Triển khai chiến thuật giao dịch tự động hóa cơ bản (ví dụ: giao cắt đường trung bình động MA).

---


## Tính năng cốt lõi

### 1. Cơ chế quản lý Token (Token Cache)
Tránh việc gọi lại luồng xác thực (`authenticate`) và OTP ở mỗi lần khởi chạy thông qua cơ chế lưu trữ cục bộ:
* **Tệp lưu trữ:** `token_cache.json` (tự động tạo sau lần đăng nhập đầu tiên).
* **Quy trình:**
  1. Kiểm tra và tải token đã lưu từ cache cục bộ.
  2. Nếu chưa có token, gọi hàm xác thực lần đầu (có thể yêu cầu OTP) rồi lưu xuống đĩa.
  3. Nếu token còn hạn, sử dụng trực tiếp mà không cần tương tác qua mạng.
  4. Nếu token đã hết hạn nhưng Refresh Token còn hạn, hệ thống tự động làm mới để lấy Access Token mới và cập nhật tệp cache.

### 2. Danh sách các mã nguồn mẫu (Samples)

Dưới đây là chi tiết các mẫu kiểm thử tương ứng theo từng ngôn ngữ:

| Stt | Tính năng | Tệp nguồn Go | Tệp nguồn Python | Tệp nguồn Node.js | Tệp nguồn .NET | Mô tả chi tiết |
|---|---|---|---|---|---|---|
| **01** | Xác thực & OTP | `sample_01_auth.go` | `sample_01_auth.py`<br>`sample_01_auth_async.py` | `sample_01_auth.js` | `Sample01Auth.cs` | Thiết lập kết nối ban đầu, yêu cầu/xác thực OTP (SMS/Email hoặc Smart OTP Push Polling), lấy token và ghi nhận tài khoản khả dụng. |
| **02** | Chỉ số thị trường | `sample_02_index_list.go` | `sample_02_index_list.py`<br>`sample_02_index_list_async.py` | `sample_02_index_list.js` | `Sample02IndexList.cs` | Lấy danh sách chỉ số (VN-Index, HNX-Index...) và chi tiết từng chỉ số theo sàn. |
| **03** | Dữ liệu nến (OHLC) | `sample_03_ohlc.go` | `sample_03_ohlc.py`<br>`sample_03_ohlc_async.py` | `sample_03_ohlc.js` | `Sample03Ohlc.cs` | Lấy mảng dữ liệu giá (Mở, Cao, Thấp, Đóng, Khối lượng) theo các khung thời gian linh hoạt (1m, 5m, 1h, 1d...). |
| **04** | Danh sách cổ phiếu | `sample_04_securities.go` | `sample_04_securities.py`<br>`sample_04_securities_async.py` | `sample_04_securities.js` | `Sample04Securities.cs` | Truy vấn thông tin chi tiết của một mã hoặc danh sách mã theo sàn giao dịch & Master Data. |
| **05** | Số dư tài khoản | `sample_05_balance.go` | `sample_05_balance.py`<br>`sample_05_balance_async.py` | `sample_05_balance.js` | `Sample05Balance.cs` | Kiểm tra số dư khả dụng, tiền ký quỹ cho tiểu khoản thường (equity) hoặc phái sinh (derivative). |
| **06** | Đặt lệnh giới hạn | `sample_06_limit_order.go` | `sample_06_limit_order.py`<br>`sample_06_limit_order_async.py` | `sample_06_limit_order.js` | `Sample06LimitOrder.cs` | Gửi lệnh mua hoặc bán cổ phiếu với mức giá giới hạn (LO) mong muốn. |
| **07** | Đặt lệnh thị trường | `sample_07_market_order.go` | `sample_07_market_order.py`<br>`sample_07_market_order_async.py` | `sample_07_market_order.js` | `Sample07MarketOrder.cs` | Gửi lệnh mua hoặc bán theo giá thị trường (MP/MTL...) nhằm ưu tiên khớp ngay lập tức. |
| **08** | Trạng thái lệnh | `sample_08_order_status.go` | `sample_08_order_status.py`<br>`sample_08_order_status_async.py` | `sample_08_order_status.js` | `Sample08OrderStatus.cs` | Kiểm tra lịch sử đặt lệnh trong ngày hoặc quá khứ của một tài khoản cụ thể. |
| **09** | Hủy lệnh | `sample_09_cancel_order.go` | `sample_09_cancel_order.py`<br>`sample_09_cancel_order_async.py` | `sample_09_cancel_order.js` | `Sample09CancelOrder.cs` | Hủy phần khối lượng chưa khớp của lệnh giới hạn đang ở trạng thái chờ khớp. |
| **10** | WebSocket Thị trường | `sample_10_websocket_data.go` | `sample_10_websocket_data.py`<br>`sample_10_websocket_data_async.py` | `sample_10_websocket_data.js` | `Sample10WebsocketData.cs` | Nhận luồng dữ liệu thời gian thực (giá khớp, thông tin bảng giá, room khối ngoại) bằng kết nối WebSocket. |
| **11** | WebSocket Giao dịch | `sample_11_websocket_trading.go` | `sample_11_websocket_trading.py`<br>`sample_11_websocket_trading_async.py` | `sample_11_websocket_trading.js` | `Sample11WebsocketTrading.cs` | Lắng nghe các thay đổi tức thời về trạng thái khớp lệnh và danh mục tài sản của người dùng. |
| **12** | Chiến thuật MA Cross | `sample_12_ma_cross_auto_trade.go` | `sample_12_ma_cross_auto_trade.py`<br>`sample_12_ma_cross_auto_trade_async.py` | `sample_12_ma_cross_auto_trade.js` | `Sample12MaCrossAutoTrade.cs` | Mô phỏng hệ thống tự động hóa hoàn chỉnh: Tính toán MA5/MA10, tạo tín hiệu giao dịch, kiểm tra điều kiện rủi ro, đặt lệnh và theo dõi. |
| **13** | Lệnh điều kiện (FCO) | `sample_13_fco_order.go` | `sample_13_fco_order.py`<br>`sample_13_fco_order_async.py` | `sample_13_fco_order.js` | `Sample13FcoOrder.cs` | Đặt và quản lý toàn bộ các loại lệnh điều kiện FCO (GTD, Stop, Stop Limit, Trailing Stop, Trailing Stop Limit, OCO, Bull Bear, truy vấn & hủy lệnh FCO). |

---

## Hướng dẫn cấu hình

Trước khi vận hành bất kỳ mẫu kiểm thử nào, bạn cần điền thông tin tài khoản được cấp từ hệ thống SSI FastConnect vào phần cấu hình.

### Cách 1: Sử dụng tệp `config.json` chung (Khuyến nghị)
Tất cả các ngôn ngữ (Python, .NET, Node.js, Go) đều được thiết lập tự động đọc thông số từ tệp `config.json` đặt tại **thư mục gốc** của dự án (`ssi-fastconnect-v3-tutorials/config.json`).

1. Tạo tệp `config.json` bằng cách copy từ tệp mẫu `config.example.json`:
   ```bash
   cp config.example.json config.json
   ```
2. Mở `config.json` và thay đổi các tham số tương ứng:
   ```json
   {
     "client_id": "YOUR_CLIENT_ID",
     "api_key": "YOUR_API_KEY",
     "api_secret": "YOUR_API_SECRET",
     "private_key": "-----BEGIN RSA PRIVATE KEY-----\nYOUR_PRIVATE_KEY_HERE\n-----END RSA PRIVATE KEY-----",
     "equity_account": "YOUR_EQUITY_ACCOUNT_NO",
     "derivative_account": "YOUR_DERIVATIVE_ACCOUNT_NO",
     "otp": "YOUR_OTP_CODE",
     "log_level": "INFO"
   }
   ```

### Cách 2: Cấu hình trực tiếp trong từng ngôn ngữ
Nếu không dùng `config.json`, bạn có thể cập nhật tham số trực tiếp trong file config của từng ngôn ngữ:
* **Python:** Cập nhật `python/config.py`
* **.NET:** Cập nhật `dotnet/SampleConfig.cs`
* **Node.js:** Cập nhật `node/config.js`
* **Go:** Cập nhật hàm `loadConfig()` hoặc struct cấu hình trong các file `go/sample_*.go`

### Giải thích các tham số cấu hình
* **`client_id`**: Định danh khách hàng đăng ký dịch vụ FastConnect API.
* **`api_key`**: Khóa API được cấp từ SSI Developer Portal.
* **`api_secret`**: Secret key dùng để ký hoặc xác thực thông điệp API.
* **`private_key`**: Chuỗi RSA Private Key (định dạng PEM hoặc XML Base64) dùng để ký giao dịch/lệnh.
* **`equity_account`**: Số tài khoản chứng khoán cơ sở (ví dụ: `123C456789`).
* **`derivative_account`**: Số tài khoản phái sinh (nếu giao dịch phái sinh).
* **`otp`**: Mã OTP 6 số (chỉ sử dụng cho lần đầu đăng nhập nếu tài khoản không dùng Smart OTP Push Notification).

> [!WARNING]
> File `config.json` và `token_cache.json` đã được đưa vào `.gitignore`. Không bao giờ commit các thông tin bí mật cấu hình (`api_secret`, `private_key`) lên kho mã nguồn công khai.

---

## Cách chạy từng mẫu kiểm thử (Samples)

### 1. Python
```bash
cd python

# Sample 01 — Xác thực & OTP (SMS/Email / Push Smart OTP)
python sample_01_auth.py             # Phiên bản đồng bộ
python sample_01_auth_async.py       # Phiên bản bất đồng bộ

# Sample 02 — Chỉ số thị trường (Index)
python sample_02_index_list.py
python sample_02_index_list_async.py

# Sample 03 — Dữ liệu nến (OHLC)
python sample_03_ohlc.py
python sample_03_ohlc_async.py

# Sample 04 — Danh sách cổ phiếu
python sample_04_securities.py
python sample_04_securities_async.py

# Sample 05 — Số dư tài khoản (Account Balance)
python sample_05_balance.py
python sample_05_balance_async.py

# Sample 06 — Đặt lệnh Limit (LO)
python sample_06_limit_order.py
python sample_06_limit_order_async.py

# Sample 07 — Đặt lệnh Market (MP)
python sample_07_market_order.py
python sample_07_market_order_async.py

# Sample 08 — Trạng thái lệnh
python sample_08_order_status.py
python sample_08_order_status_async.py

# Sample 09 — Hủy lệnh
python sample_09_cancel_order.py
python sample_09_cancel_order_async.py

# Sample 10 — WebSocket dữ liệu thị trường real-time
python sample_10_websocket_data.py
python sample_10_websocket_data_async.py

# Sample 11 — WebSocket trading real-time
python sample_11_websocket_trading.py
python sample_11_websocket_trading_async.py

# Sample 12 — Tự động giao dịch theo MA Cross
python sample_12_ma_cross_auto_trade.py
python sample_12_ma_cross_auto_trade_async.py

# Sample 13 — Đặt & quản lý lệnh điều kiện FCO
python sample_13_fco_order.py
python sample_13_fco_order_async.py
```

### 2. .NET
```bash
cd dotnet

dotnet run -- 01    # Sample 01 — Xác thực & OTP
dotnet run -- 02    # Sample 02 — Chỉ số thị trường
dotnet run -- 03    # Sample 03 — Dữ liệu nến (OHLC)
dotnet run -- 04    # Sample 04 — Danh sách cổ phiếu
dotnet run -- 05    # Sample 05 — Số dư tài khoản
dotnet run -- 06    # Sample 06 — Đặt lệnh Limit
dotnet run -- 07    # Sample 07 — Đặt lệnh Market
dotnet run -- 08    # Sample 08 — Trạng thái lệnh
dotnet run -- 09    # Sample 09 — Hủy lệnh
dotnet run -- 10    # Sample 10 — WebSocket Thị trường
dotnet run -- 11    # Sample 11 — WebSocket Giao dịch
dotnet run -- 12    # Sample 12 — Chiến thuật MA Cross
dotnet run -- 13    # Sample 13 — Lệnh điều kiện FCO
```

### 3. Node.js
```bash
cd node
npm install

npm run sample:01   # Sample 01 — Xác thực & OTP
npm run sample:02   # Sample 02 — Chỉ số thị trường
npm run sample:03   # Sample 03 — Dữ liệu nến (OHLC)
npm run sample:04   # Sample 04 — Danh sách cổ phiếu
npm run sample:05   # Sample 05 — Số dư tài khoản
npm run sample:06   # Sample 06 — Đặt lệnh Limit
npm run sample:07   # Sample 07 — Đặt lệnh Market
npm run sample:08   # Sample 08 — Trạng thái lệnh
npm run sample:09   # Sample 09 — Hủy lệnh
npm run sample:10   # Sample 10 — WebSocket Thị trường
npm run sample:11   # Sample 11 — WebSocket Giao dịch
npm run sample:12   # Sample 12 — Chiến thuật MA Cross
npm run sample:13   # Sample 13 — Lệnh điều kiện FCO

# Hoặc chạy trực tiếp bằng Node:
node sample_01_auth.js
```

### 4. Go
```bash
cd go

go run sample_01_auth.go                # Sample 01 — Xác thực & OTP
go run sample_02_index_list.go          # Sample 02 — Chỉ số thị trường
go run sample_03_ohlc.go                # Sample 03 — Dữ liệu nến (OHLC)
go run sample_04_securities.go          # Sample 04 — Danh sách cổ phiếu
go run sample_05_balance.go             # Sample 05 — Số dư tài khoản
go run sample_06_limit_order.go         # Sample 06 — Đặt lệnh Limit
go run sample_07_market_order.go        # Sample 07 — Đặt lệnh Market
go run sample_08_order_status.go        # Sample 08 — Trạng thái lệnh
go run sample_09_cancel_order.go        # Sample 09 — Hủy lệnh
go run sample_10_websocket_data.go      # Sample 10 — WebSocket Thị trường
go run sample_11_websocket_trading.go   # Sample 11 — WebSocket Giao dịch
go run sample_12_ma_cross_auto_trade.go # Sample 12 — Chiến thuật MA Cross
go run sample_13_fco_order.go          # Sample 13 — Lệnh điều kiện FCO
```

---

## Luồng hoạt động chính

### Quy trình thủ công (Manual Flow)
Quy trình giao dịch thông qua các tương tác tuần tự của người dùng:
```
[Auth & Token Cache] ──> [Market Data Lookup] ──> [Account Balance Check] ──> [Place Order] ──> [Monitor Status / Cancel Order]
```

### Quy trình tự động (Algorithmic Trading Flow)
Luồng xử lý tự động liên tục được kiểm soát bởi chiến thuật giao dịch:
```
[Auth & Token Cache] ──> [Real-time/OHLC Feeds] ──> [Signal Generation] ──> [Risk Management Check] ──> [Execution (LO/MP)] ──> [Order Tracking]
```

---

## Lưu ý quan trọng
* **Cơ chế tái kết nối (Reconnect):** Các tệp mẫu WebSocket (`sample_10`, `sample_11`) tích hợp sẵn cơ chế Exponential Backoff để tự động thiết lập lại kết nối khi đường truyền mạng gặp sự cố ngắt quãng.
* **Thời gian hết hạn của token:** Access Token mặc định có thời hạn sử dụng hữu hạn. Sử dụng `auth_helper` sẽ giúp mã nguồn của bạn hoạt động ổn định trong thời gian dài mà không bị gián đoạn do lỗi hết hạn phiên làm việc.
* **Quản lý rủi ro:** Luôn sử dụng môi trường giả lập (UAT/Sandbox) trước khi chính thức áp dụng các mã nguồn tự động hóa giao dịch trên tài khoản thật.

