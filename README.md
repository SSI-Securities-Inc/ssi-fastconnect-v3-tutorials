# SSI FastConnect API Sample Codes

Tài liệu hướng dẫn và mã nguồn mẫu tích hợp hệ thống giao dịch chứng khoán của SSI (Saigon Securities Incorporation) qua **SSI FastConnect API**.

### Ngôn ngữ hỗ trợ
- [x] **Go** (bao gồm các mẫu từ Sample 01 đến 12)
- [x] **Python** (bao gồm cả phiên bản đồng bộ - synchronous và bất đồng bộ - asynchronous)
- [ ] **Các ngôn ngữ khác** (sẽ được cập nhật thêm trong tương lai)

## Mục lục
- [Giới thiệu](#giới-thiệu)
- [Tính năng cốt lõi](#tính-năng-cốt-lõi)
- [Hướng dẫn cấu hình](#hướng-dẫn-cấu-hình)
- [Cách chạy mẫu Go](#cách-chạy-mẫu-go)
- [Cách chạy mẫu Python](#cách-chạy-mẫu-python)
- [Luồng hoạt động chính](#luồng-hoạt-động-chính)
- [Lưu ý quan trọng](#lưu-ý-quan-trọng)

---

## Giới thiệu
Bộ mã nguồn này cung cấp các ví dụ thực tế giúp nhà phát triển nhanh chóng nắm bắt cách thức giao tiếp với cổng API của SSI để thực hiện:
* Truy vấn thông tin tài khoản và số dư.
* Lấy dữ liệu thị trường (chỉ số, thông tin cổ phiếu, dữ liệu nến lịch sử và intraday).
* Thực hiện giao dịch (đặt lệnh LO, lệnh MP, hủy lệnh, theo dõi trạng thái lệnh).
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

| Stt | Tính năng | Tệp nguồn Go | Tệp nguồn Python | Mô tả chi tiết |
|---|---|---|---|---|
| **01** | Xác thực & Lấy Token | `sample_01_auth.go` | `sample_01_auth.py`<br>`sample_01_auth_async.py` | Thiết lập kết nối ban đầu, lấy token và ghi nhận tài khoản khả dụng. |
| **02** | Chỉ số thị trường | `sample_02_index_list.go` | `sample_02_index_list.py`<br>`sample_02_index_list_async.py` | Lấy danh sách chỉ số (VN-Index, HNX-Index...) và chi tiết từng chỉ số theo sàn. |
| **03** | Dữ liệu nến (OHLC) | `sample_03_ohlc.go` | `sample_03_ohlc.py`<br>`sample_03_ohlc_async.py` | Lấy mảng dữ liệu giá (Mở, Cao, Thấp, Đóng, Khối lượng) theo các khung thời gian linh hoạt (1m, 5m, 1h, 1d...). |
| **04** | Danh sách cổ phiếu | `sample_04_securities.go` | `sample_04_securities.py`<br>`sample_04_securities_async.py` | Truy vấn thông tin chi tiết của một mã hoặc danh sách mã theo sàn giao dịch. |
| **05** | Số dư tài khoản | `sample_05_balance.go` | `sample_05_balance.py`<br>`sample_05_balance_async.py` | Kiểm tra số dư khả dụng, tiền ký quỹ cho tiểu khoản thường (equity) hoặc phái sinh (derivative). |
| **06** | Đặt lệnh giới hạn | `sample_06_limit_order.go` | `sample_06_limit_order.py`<br>`sample_06_limit_order_async.py` | Gửi lệnh mua hoặc bán cổ phiếu với mức giá giới hạn (LO) mong muốn. |
| **07** | Đặt lệnh thị trường | `sample_07_market_order.go` | `sample_07_market_order.py`<br>`sample_07_market_order_async.py` | Gửi lệnh mua hoặc bán theo giá thị trường (MP/MTL...) nhằm ưu tiên khớp ngay lập tức. |
| **08** | Trạng thái lệnh | `sample_08_order_status.go` | `sample_08_order_status.py`<br>`sample_08_order_status_async.py` | Kiểm tra lịch sử đặt lệnh trong ngày hoặc quá khứ của một tài khoản cụ thể. |
| **09** | Hủy lệnh | `sample_09_cancel_order.go` | `sample_09_cancel_order.py`<br>`sample_09_cancel_order_async.py` | Hủy phần khối lượng chưa khớp của lệnh giới hạn đang ở trạng thái chờ khớp. |
| **10** | WebSocket Thị trường | `sample_10_websocket_data.go` | `sample_10_websocket_data.py`<br>`sample_10_websocket_data_async.py` | Nhận luồng dữ liệu thời gian thực (giá khớp, thông tin bảng giá, room khối ngoại) bằng kết nối WebSocket. |
| **11** | WebSocket Giao dịch | `sample_11_websocket_trading.go` | `sample_11_websocket_trading.py`<br>`sample_11_websocket_trading_async.py` | Lắng nghe các thay đổi tức thời về trạng thái khớp lệnh và danh mục tài sản của người dùng. |
| **12** | Chiến thuật MA Cross | `sample_12_ma_cross_auto_trade.go` | `sample_12_ma_cross_auto_trade.py`<br>`sample_12_ma_cross_auto_trade_async.py` | Mô phỏng hệ thống tự động hóa hoàn chỉnh: Tính toán MA5/MA10, tạo tín hiệu giao dịch, kiểm tra điều kiện rủi ro, đặt lệnh và theo dõi. |

---

## Hướng dẫn cấu hình

Trước khi vận hành bất kỳ mẫu kiểm thử nào, bạn cần điền thông tin tài khoản được cấp từ hệ thống SSI FastConnect vào phần cấu hình của mã nguồn.

### Tham số cấu hình cần thiết
* **`client_id`**: Định danh của khách hàng đăng ký sử dụng dịch vụ API.
* **`api_key`**: Khóa truy cập được cấp bởi cổng thông tin phát triển SSI.
* **`api_secret`**: Khóa bí mật đi kèm để ký các yêu cầu API.
* **`private_key`**: Đường dẫn tệp khóa riêng tư hoặc nội dung khóa riêng tư định dạng PEM dùng để ký số xác thực.

> [!WARNING]
> Không chia sẻ hoặc đưa các thông tin bí mật cấu hình (`api_secret`, `private_key`) lên các kho mã nguồn công khai như GitHub để tránh rủi ro mất an toàn tài khoản.

---

## Cách chạy mẫu Python

### Yêu cầu hệ thống
* Python phiên bản `3.9` trở lên.

### Cài đặt SDK
Cài đặt thư viện `ssi-sdk` chính thức từ PyPI:
```bash
pip install ssi-sdk
```

### Các bước thực hiện
1. Di chuyển vào thư mục Python:
   ```bash
   cd python
   ```
2. Cập nhật thông tin xác thực tại tệp cấu hình hoặc trực tiếp trong từng file sample.
3. Chạy thử nghiệm các tệp mẫu (Ví dụ chạy phiên bản đồng bộ hoặc bất đồng bộ của Sample 01):
   ```bash
   # Phiên bản đồng bộ
   python sample_01_auth.py

   # Phiên bản bất đồng bộ
   python sample_01_auth_async.py
   ```

---

## Cách chạy mẫu Go

### Yêu cầu hệ thống
* Go phiên bản `1.22` trở lên.
* Module `github.com/SSI-Securities-Inc/ssi-sdk-go/v3` (được thiết lập liên kết cục bộ tới thư mục SDK song song `../ssi-sdk-go`).

### Các bước thực hiện
1. Thực hiện tải về các thư viện phụ thuộc và thiết lập module:
   ```bash
   go mod tidy
   ```
2. Chạy thử nghiệm các tệp mẫu mong muốn trực tiếp tại thư mục hiện tại:
   ```bash
   # Ví dụ chạy Sample 01 xác thực tài khoản
   go run sample_01_auth.go

   # Ví dụ chạy Sample 12 chiến thuật giao dịch tự động
   go run sample_12_ma_cross_auto_trade.go
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
