# FastConnect Python SDK Samples

Bộ sample minh họa 19 kịch bản sử dụng **SSI FastConnect Python SDK** (`ssi-sdk`). Hỗ trợ cả hai phiên bản **Synchronous (đồng bộ)** và **Asynchronous (bất đồng bộ)**.

## 1. Cấu hình (`config.json` / `config.py`)

Tất cả sample trong thư mục Python tự động tải cấu hình từ tệp `config.json` ở thư mục gốc dự án (`../config.json`).

1. Tạo tệp `config.json` tại thư mục gốc dự án:
   ```bash
   cp ../config.example.json ../config.json
   ```
2. Điền các tham số API (`client_id`, `api_key`, `api_secret`, `private_key`, `equity_account`, `otp`) vào `config.json`.
3. Hoặc bạn có thể sửa trực tiếp trong `python/config.py`.

---

## 2. Hướng dẫn chạy từng Sample

### Chạy phiên bản Đồng bộ (Sync)
```bash
python sample_01_auth.py                # Sample 01 — Auth & OTP
python sample_02_index_list.py          # Sample 02 — Chỉ số thị trường
python sample_03_ohlc.py                # Sample 03 — Dữ liệu nến OHLC
python sample_04_securities.py          # Sample 04 — Danh sách cổ phiếu
python sample_05_balance.py             # Sample 05 — Số dư tài khoản
python sample_06_limit_order.py         # Sample 06 — Đặt lệnh Limit
python sample_07_market_order.py        # Sample 07 — Đặt lệnh Market
python sample_08_order_status.py        # Sample 08 — Trạng thái lệnh
python sample_09_cancel_order.py        # Sample 09 — Hủy lệnh
python sample_10_websocket_data.py      # Sample 10 — WebSocket Thị trường
python sample_11_websocket_trading.py   # Sample 11 — WebSocket Giao dịch
python sample_12_ma_cross_auto_trade.py # Sample 12 — Chiến thuật MA Cross
python sample_13_fco_order.py          # Sample 13 — Lệnh điều kiện FCO
python sample_14_batch_orders.py       # Sample 14 — Đặt/hủy nhiều lệnh (batch)
python sample_15_positions_ppmmr.py    # Sample 15 — Vị thế T1/T2, PPMMR, max mua/bán
python sample_16_streaming_advanced.py # Sample 16 — Streaming nâng cao
python sample_17_trading_ws.py         # Sample 17 — Trading WebSocket
python sample_18_error_handling.py     # Sample 18 — Xử lý lỗi
python sample_19_corporate_actions.py  # Sample 19 — Sự kiện doanh nghiệp (corporate actions)
```

### Chạy phiên bản Bất đồng bộ (Async)
```bash
python sample_01_auth_async.py
python sample_02_index_list_async.py
python sample_03_ohlc_async.py
python sample_04_securities_async.py
python sample_05_balance_async.py
python sample_06_limit_order_async.py
python sample_07_market_order_async.py
python sample_08_order_status_async.py
python sample_09_cancel_order_async.py
python sample_10_websocket_data_async.py
python sample_11_websocket_trading_async.py
python sample_12_ma_cross_auto_trade_async.py
python sample_13_fco_order_async.py
python sample_14_batch_orders_async.py
python sample_15_positions_ppmmr_async.py
python sample_16_streaming_advanced_async.py
python sample_17_trading_ws_async.py
python sample_18_error_handling_async.py
python sample_19_corporate_actions_async.py
```

---

## auth_helper — Tái sử dụng token (Token Cache)
**Mục tiêu:** Tránh authenticate lại mỗi lần chạy script bằng cách cache token xuống file.

**Luồng xử lý code:**
1. Load token từ file `token_cache.json` (nếu có).
2. Nếu chưa có → gọi quy trình Request OTP & Authenticate lần đầu và lưu token xuống file.
3. Nếu token hết hạn → gọi `refresh()` và lưu token mới.
4. Nếu token còn hạn → dùng trực tiếp, không gọi API.
5. Cả hai phiên bản sync (`ensure_auth`) và async (`ensure_auth_async`) đều được hỗ trợ.

**File:** `python/auth_helper.py`

---

## helpers — Tiện ích dùng chung
**Mục tiêu:** Gom các hàm nhỏ mà mọi sample cần từ SDK 3.3.0.

- `enum_value(x)`: các trường enum (`Account.account_type`, `Order.side/order_type/status`...) có thể là enum, `str` (giá trị server mới mà SDK chưa biết) hoặc `None`. Hàm này lấy chuỗi hiển thị an toàn cho cả 3 trường hợp.
- `mask_account(account_no)`: che số tài khoản khi in, chỉ giữ 3 ký tự cuối.
- `require_account(account_no)`: dừng ngay (fail fast) nếu `config.json` thiếu `equity_account`.
- `confirm_real_orders(what)`: cảnh báo lệnh THẬT và bắt gõ `YES` mới chạy tiếp.

`config.py` còn đọc thêm `derivative_account` (`DERIVATIVE_ACCOUNT_NO`) cho sample 15.

**File:** `python/helpers.py` · `python/config.py`

---

## Sample 1 — Xác thực, Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
**Mục tiêu:** Đăng nhập, yêu cầu/xác thực mã OTP (SMS/Email hoặc Smart OTP Push Notification) và lấy token cho toàn bộ API call sau đó.

**Luồng xử lý code:**
1. Tạo `Config` → `Auth` → gọi `ensure_auth(auth, otp=...)`.
2. `ensure_auth` load token từ cache, refresh nếu hết hạn.
3. Nếu chưa có token hợp lệ:
   - Gọi `request_otp()` gửi yêu cầu mã OTP (SMS/Email/Smart OTP Push).
   - Nếu tài khoản dùng Smart OTP Push Notification: Nhận `transactionId` từ response và gọi `ensure_authenticated(transaction_id=...)` để SDK tự động Polling chờ người dùng bấm Approve trên ứng dụng di động SSI.
   - Nếu tài khoản dùng mã OTP 6 số: Nhập mã OTP và gọi `authenticate(otp=...)`.
4. Auth service trả về `accessToken`, `refreshToken`, `expiresIn` và lưu vào `token_cache.json`.
5. Mọi request sau đó gắn `Authorization: Bearer <accessToken>`.
6. Tự động kiểm tra token bằng cách gọi API lấy danh sách tài khoản.

**File:** `python/sample_01_auth.py` · `python/sample_01_auth_async.py`

**Thành phần kết nối:** Client App → Auth API → `token_cache.json`.

---

## Sample 2 — Lấy danh sách chỉ số thị trường (Index)
**Mục tiêu:** Hiển thị VN-Index, HNX-Index… trên dashboard.

**Luồng xử lý code:**
1. Client gọi endpoint `indexList` với filter (`exchange`, `limit`, `view`).
2. API trả về danh sách chỉ số và dữ liệu giá hiện tại.
3. Code map dữ liệu sang model hiển thị UI.
4. Lấy summary chi tiết cho một chỉ số cụ thể (`get_index_summary`).
5. Lọc chỉ số theo sàn (`get_indexes_by_board`).

**File:** `python/sample_02_index_list.py` · `python/sample_02_index_list_async.py`

**Thành phần kết nối:** UI Dashboard → Market Data API (Index) → Cache State.

---

## Sample 3 — Lấy dữ liệu K-line (OHLC)
**Mục tiêu:** Cung cấp dữ liệu nến cho biểu đồ và phân tích kỹ thuật.

**Luồng xử lý code:**
1. Client gửi `symbol`, `interval`, `startTime`, `endTime`.
2. API trả về mảng OHLCV theo mốc thời gian.
3. Code chuẩn hóa dữ liệu (time/open/high/low/close/volume).
4. Hỗ trợ nhiều timeframe: 1m, 3m, 5m, 15m, 1h, 1d (intraday & historical).
5. Nếu lịch sử dài thì lặp theo paging/window thời gian.

**File:** `python/sample_03_ohlc.py` · `python/sample_03_ohlc_async.py`

**Thành phần kết nối:** Chart/Strategy Module → Market Data API (OHLC) → Chart Engine.

---

## Sample 4 — Lấy danh sách cổ phiếu theo sàn
**Mục tiêu:** Tạo watchlist/screener theo tiêu chí thị trường.

**Luồng xử lý code:**
1. Client gọi `get_securities_info` theo `symbol` hoặc `get_securities_info_by_board`.
2. API trả về danh sách mã + thông tin giao dịch cơ bản.
3. Lấy summary giá hiện tại qua `get_securities_summary`.
4. Lấy lịch sử summary qua `get_securities_summary_historical`.
5. Khi user chọn mã, chuyển sang luồng xem chi tiết/đặt lệnh.

**File:** `python/sample_04_securities.py` · `python/sample_04_securities_async.py`

**Thành phần kết nối:** Screener → Market Data API (Securities) → Watchlist State.

---

## Sample 5 — Lấy số dư tài khoản (Account Balance)
**Mục tiêu:** Kiểm tra khả năng giao dịch trước khi đặt lệnh.

**Luồng xử lý code:**
1. Client gọi `get_equity_balance` hoặc `get_derivative_balance` theo `accountNo`.
2. API trả về `available`, `onHold`, `limits`, `settlement`.
3. Code tính khả năng mua/bán thực tế theo nghiệp vụ.
4. Nếu không đủ điều kiện thì chặn thao tác đặt lệnh.
5. Nếu đủ điều kiện thì cho phép đi tiếp sang order flow.

**File:** `python/sample_05_balance.py` · `python/sample_05_balance_async.py`

**Thành phần kết nối:** Trading Service → Account API → Validation Rule Engine.

---

## Sample 6 — Đặt lệnh Limit
**Mục tiêu:** Đặt lệnh mua/bán tại mức giá chỉ định.

**Luồng xử lý code:**
1. Client tạo payload order (`symbol`, `side`, `quantity`, `price`).
2. Gọi `place_limit_order` với `OrderType.LO`.
3. Gửi request tới Trading Orders API.
4. API trả về `orderId` và trạng thái ban đầu (`PENDING`).
5. Code lưu `orderId` để theo dõi khớp lệnh.

**File:** `python/sample_06_limit_order.py` · `python/sample_06_limit_order_async.py`

**Thành phần kết nối:** Order → Trading API (Create Order) → Order Tracking.

---

## Sample 7 — Đặt lệnh Market
**Mục tiêu:** Khớp lệnh nhanh theo giá thị trường hiện tại.

**Luồng xử lý code:**
1. Client tạo order `MARKET` qua `place_market_order` (không truyền `price`).
2. Gửi request tới Trading Orders API.
3. Hệ thống match theo thanh khoản thị trường tại thời điểm gửi.
4. API trả về trạng thái khớp (`FILLED` hoặc `PARTIALLY_FILLED`).
5. Code cập nhật ngay danh mục/số dư tạm tính.

**File:** `python/sample_07_market_order.py` · `python/sample_07_market_order_async.py`

**Thành phần kết nối:** Quick Trade → Trading Matching Engine (qua API) → Portfolio.

---

## Sample 8 — Kiểm tra trạng thái lệnh
**Mục tiêu:** Theo dõi tiến trình khớp của một lệnh cụ thể.

**Luồng xử lý code:**
1. Client gọi `get_today_orders` hoặc `get_historical_orders` theo `accountNo`.
2. API trả về `status`, `filledQuantity`, `fills`.
3. Code đối chiếu lượng còn lại và trạng thái hiện tại.
4. Nếu chưa hoàn tất thì tiếp tục polling chu kỳ ngắn.
5. Khi `FILLED/CANCELLED/REJECTED` thì đóng vòng theo dõi.

**File:** `python/sample_08_order_status.py` · `python/sample_08_order_status_async.py`

**Thành phần kết nối:** Order Monitor Service → Trading API (Order Detail) → Execution.

---

## Sample 9 — Hủy lệnh
**Mục tiêu:** Dừng phần khối lượng chưa khớp của lệnh đang mở.

**Luồng xử lý code:**
1. User bấm hủy trên lệnh đang `PENDING/PARTIALLY_FILLED`.
2. Client gọi `cancel_order` kèm `clientRequestId` hoặc `cancel_order_by_order_id`.
3. API xác thực quyền và trạng thái lệnh hiện tại.
4. Nếu hợp lệ, hệ thống cập nhật `CANCELLED` cho phần chưa khớp.
5. Code đồng bộ lại sổ lệnh và số lượng còn treo.

**File:** `python/sample_09_cancel_order.py` · `python/sample_09_cancel_order_async.py`

**Thành phần kết nối:** Open Orders → Trading API (Cancel Order) → Order Book State.

---

## Sample 10 — WebSocket dữ liệu thị trường real-time
**Mục tiêu:** Nhận tick data (giá khớp, bảng giá, room nước ngoài) tức thời.

**Luồng xử lý code:**
1. Client mở kết nối WebSocket bằng token hợp lệ (không cần OTP).
2. Subscribe stream theo symbol (`subscribe_symbol`) và index (`subscribe_index`).
3. Server push event khi có giao dịch mới hoặc bảng giá thay đổi.
4. Callback `on_data` phân loại message: `TradeMessage`, `QuoteMessage`, `ForeignRoomMessage`.
5. Khi mất kết nối, chạy cơ chế reconnect exponential backoff.

**File:** `python/sample_10_websocket_data.py` · `python/sample_10_websocket_data_async.py`

**Thành phần kết nối:** Client ↔ Streaming API ↔ Market Data Event.

---

## Sample 11 — WebSocket trading real-time (trạng thái lệnh & danh mục)
**Mục tiêu:** Nhận cập nhật tức thời về lệnh khớp và danh mục tài khoản.

**Luồng xử lý code:**
1. Client mở kết nối WebSocket bằng token hợp lệ (cần OTP).
2. Subscribe stream lệnh (`subscribe_order_status`) và danh mục (`subscribe_portfolio`) theo `accountNo`.
3. Server push event khi trạng thái lệnh hoặc danh mục thay đổi.
4. Callback `on_trading` phân loại message: `OrderStatusMessage`, `PortfolioMessage`.
5. Khi mất kết nối, chạy cơ chế reconnect exponential backoff.

**File:** `python/sample_11_websocket_trading.py` · `python/sample_11_websocket_trading_async.py`

**Thành phần kết nối:** Client ↔ Streaming API ↔ Trading Event.

---

## Sample 12 — MA Cross Signal + Auto Place & Monitor
**Mục tiêu:** Tự động giao dịch khi MA5 cắt MA10.

**Luồng xử lý code:**
1. Lấy OHLC theo chu kỳ để tính MA(5), MA(10).
2. Kiểm tra điều kiện giao cắt tại nến hiện tại so với nến trước.
3. Có tín hiệu thì kiểm tra balance/risk rule trước khi vào lệnh.
4. Tự động đặt lệnh (thường MARKET để ưu tiên khớp nhanh).
5. Theo dõi trạng thái đến khi `FILLED`, timeout thì hủy lệnh.
6. Ghi log kết quả giao dịch và tính P&L cơ bản.

**File:** `python/sample_12_ma_cross_auto_trade.py` · `python/sample_12_ma_cross_auto_trade_async.py`

**Thành phần kết nối:** Strategy → Market Data API → Account API → Trading API → Monitor/Logger.

---

## Sample 13 — Đặt Lệnh Điều Kiện (FCO)
**Mục tiêu:** Đặt và quản lý các loại lệnh điều kiện nâng cao (Fast Conditional Orders).

**Luồng xử lý code:**
1. Khởi tạo `Trading` service từ `Auth` đã xác thực.
2. Đặt các loại lệnh FCO: `place_fco_gtd`, `place_fco_stop`, `place_fco_stop_limit`, `place_fco_trailing_stop`, `place_fco_trailing_stop_limit`, `place_fco_oco`, `place_fco_bull_bear`.
3. Truy vấn danh sách lệnh điều kiện qua `get_fco_by_account_no`.
4. Hủy lệnh điều kiện FCO qua `cancel_fco(fco_id)`.

**File:** `python/sample_13_fco_order.py` · `python/sample_13_fco_order_async.py`

**Thành phần kết nối:** Client App → Trading API (FCO Endpoints).

---

## Cập nhật SDK 3.3.0 ở các sample cũ
- **Mọi thời gian tính bằng mili-giây** (`*_ms`): `wait(timeout_ms=...)`, `ping(interval_ms=...)`, `poll_interval_ms`, `poll_max_retries`. Tên cũ báo `TypeError`.
- **Sample 03:** thêm `get_ohlc_all` (tự phân trang) và `iter_ohlc` (đọc dần). Nến tuần/tháng KHÔNG được hỗ trợ (lỗi 400210): `get_ohlc_1week_historical` / `get_ohlc_1month_historical` raise `ValidationError` ngay phía client.
- **Sample 05:** hiển thị `purchasing_power` (thay `purchase_power` đã deprecated; có thể `None` với tài khoản phái sinh) và KL chờ về T1/T2 mua/bán của vị thế.
- **Sample 10/11:** `subscribe_index(["VN30"])` nhận dữ liệu CÁC MÃ THÀNH PHẦN; muốn tick của chính chỉ số dùng `subscribe_index_ticks` (`IndexTickMessage`). Thêm `OrderMatchMessage`; `PortfolioMessage` chỉ có với tài khoản phái sinh và không còn `total_asset/cash_balance/stock_value`.
- **Sample 13:** ngày FCO tính theo hôm nay (tối đa 31 ngày), thêm `get_fco_status_history` và tham số OTP `code` tùy chọn cho lệnh/hủy FCO.
- Các trường enum có thể là `str` hoặc `None`: dùng `helpers.enum_value`.

---

## Sample 14 — Đặt / Hủy Nhiều Lệnh (Batch Orders)
**Mục tiêu:** Gửi nhiều lệnh trong MỘT request đã ký RSA và xử lý kết quả từng lệnh.

**Luồng xử lý code:**
1. Cảnh báo lệnh thật, yêu cầu gõ `YES`. Lệnh là LO MUA tại giá sàn (xa thị trường, chỉ nằm chờ).
2. Dựng danh sách `BatchPlaceOrderItem`, mỗi lệnh có `client_request_id` (tối đa 20 ký tự) làm khóa idempotency; kiểm tra `<= Config.max_batch_orders` (mặc định 20).
3. `place_batch_orders` → duyệt `BatchOrderResult` (`success`, `order_id`, `error_code`, `error_message`).
4. `cancel_batch_orders` với `BatchCancelOrderItem` (mỗi item chỉ có MỘT trong `order_id` / `client_request_id`).
5. Xử lý `ValidationError`, `DuplicateRequestError` (409), `APIError`.

**File:** `python/sample_14_batch_orders.py` · `python/sample_14_batch_orders_async.py`

**Thành phần kết nối:** Client App → Market Data API (giá sàn) → Trading API (batch).

---

## Sample 15 — Vị thế, PPMMR và Max mua/bán
**Mục tiêu:** Đọc vị thế T1/T2, sức mua / margin (PPMMR) và khối lượng mua/bán tối đa. Chỉ đọc, không đặt lệnh.

**Luồng xử lý code:**
1. `get_equity_positions`: số lượng, bán được, `t1/t2_buy_quantity` và `t1/t2_sell_quantity` (KHỐI LƯỢNG chờ về).
2. `get_equity_ppmmr`: `buy_t0/t1/t2`, `sell_t0/t1/t2` (SỐ TIỀN chờ thanh toán), `margin_ratio`, `purchasing_power`.
3. Nếu có `derivative_account` trong `config.json`: `get_derivative_positions` và `get_derivative_ppmmr`.
4. `get_max_buy_sell` và `get_max_buy_sell_at_market_price`.

Lưu ý: REST chỉ trả danh sách vị thế theo mã; dòng tổng hợp (`query_summary`) chỉ có ở trading WebSocket (sample 17).

**File:** `python/sample_15_positions_ppmmr.py` · `python/sample_15_positions_ppmmr_async.py`

---

## Sample 16 — Streaming nâng cao
**Mục tiêu:** Theo dõi kết nối và dùng các stream/callback mới của SDK 3.3.0.

**Luồng xử lý code:**
1. Gắn callback trước khi `connect`: `on_connection` (`connected` / `reconnected` / `disconnected` / `failed`), `on_data`, `on_market_flag`, `on_index_summary`, `on_order`, `on_order_match`, `on_portfolio`.
2. Đọc `is_connected`, `reconnect_count`, `last_error`.
3. Subscribe: `subscribe_trades` / `subscribe_quotes`, `subscribe_board`, `subscribe_index`, `subscribe_index_ticks`, `subscribe_index_summary`, `subscribe_session_flag`, `subscribe_candles` (chỉ `tick`, `1m`, `5m`) và `subscribe_orders` / `subscribe_positions` khi có tài khoản.
4. `wait(timeout_ms=...)`, `list_subscription`, `unsubscribe_*`, `unsubscribe_all`.

Callback chạy trên thread/task của SDK nên sample chỉ đếm message, không in từng message.

**File:** `python/sample_16_streaming_advanced.py` · `python/sample_16_streaming_advanced_async.py`

---

## Sample 17 — Trading WebSocket
**Mục tiêu:** Đặt/hủy lệnh và truy vấn qua WebSocket trading thay vì REST.

**Lưu ý:** `trading_ws` / `async_trading_ws` import trực tiếp từ `ssi_sdk`; `Trading`/`Stream` không tự kết nối trading WebSocket. Token cần scope `tradingws:*`, thiếu scope server trả 403.

**Luồng xử lý code:**
1. Cảnh báo lệnh thật, yêu cầu gõ `YES`. Kết nối `trading_ws(auth)` / `async_trading_ws(auth)`.
2. Gắn `client.on_event`, `subscribe_order_events(account)`.
3. `place_order` (LO MUA tại giá sàn, nhận ack `PD`) rồi `cancel_order`.
4. `query_position(query_summary=True)` và `query_order_book`.
5. Đọc `client.rate_limit`; bắt `TradingWSError` (`status`, `error_code`, `server_message`); luôn `close()`.

**File:** `python/sample_17_trading_ws.py` · `python/sample_17_trading_ws_async.py`

---

## Sample 18 — Xử lý lỗi
**Mục tiêu:** Bắt đúng loại exception và phản ứng phù hợp.

**Luồng xử lý code:**
1. Phần A (offline): cấu hình theo `*_ms`; tên cũ báo `TypeError`.
2. Phần B (offline): hướng xử lý cho `RateLimitError` (`retry_after_ms`), `ValidationError`, `DuplicateRequestError`, `APIError` (`code` so với `ServerErrorCode.from_value`), `SmartOTPPendingError`, `SmartOTPRejectedError`, `ReauthenticationRequired`.
3. Phần C (gọi API đọc): retry có kiểm soát theo `retry_after_ms` (chỉ request GET, không bao giờ cho đặt/hủy lệnh), bắt `ValidationError` của nến tuần.
4. Không in `headers`, `response_body` hay cấu hình.

**File:** `python/sample_18_error_handling.py` · `python/sample_18_error_handling_async.py`

---

## Sample 19 — Sự kiện doanh nghiệp (Corporate actions)
**Mục tiêu:** Xem cổ tức, cổ phiếu thưởng, quyền mua của các mã. Chỉ đọc dữ liệu, không đặt lệnh.

**Lưu ý:** SDK tự lấy hết các trang. Mọi trường là chuỗi đúng như server gửi: ngày dạng `YYYY/MM/DD`, `ratio` không được parse, trường server để trống là `""`. Ngày truyền vào phải đúng dạng `YYYY/MM/DD`; mỗi lần gọi `get_corporate_actions_historical` chỉ được tối đa 5 ngày.

**Luồng xử lý code:**
1. `get_corporate_actions()`: sự kiện của hôm nay (theo ngày máy bạn).
2. `get_corporate_actions_historical(from_date, to_date)`: 5 ngày gần nhất.
3. Lọc `symbol="SSI"`, `language="en"`.
4. Khoảng 30 ngày bị SDK từ chối bằng `ValidationError` trước khi gọi server; chia khoảng ngày thành các đoạn <= 5 ngày để lấy dài hơn.

**File:** `python/sample_19_corporate_actions.py` · `python/sample_19_corporate_actions_async.py`

---

- **Manual flow:** Auth (token cache) → Market Data → Account Check → Place Order → Track/Cancel → Realtime Update.
- **Auto flow:** Auth (token cache) → OHLC/Indicator → Signal Engine → Risk Check → Auto Order → Monitor → P&L Logging.
 