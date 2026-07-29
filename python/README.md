## auth_helper — Tái sử dụng token (Token Cache)
**Mục tiêu:** Tránh authenticate lại mỗi lần chạy script bằng cách cache token xuống file.

**Luồng xử lý code:**
1. Load token từ file `token_cache.json` (nếu có).
2. Nếu chưa có → gọi `authenticate()` lần đầu và lưu token xuống file.
3. Nếu token hết hạn → gọi `refresh()` và lưu token mới.
4. Nếu token còn hạn → dùng trực tiếp, không gọi API.
5. Cả hai phiên bản sync (`ensure_auth`) và async (`ensure_auth_async`) đều được hỗ trợ.

**File:** `python/auth_helper.py`

---

## Sample 1 — Xác thực và lấy Access Token
**Mục tiêu:** Đăng nhập và lấy token cho toàn bộ API call sau đó.

**Luồng xử lý code:**
1. Tạo `Config` → `Auth` → gọi `ensure_auth(auth, otp=...)`.
2. `ensure_auth` load token từ cache, authenticate nếu chưa có, refresh nếu hết hạn.
3. Auth service trả về `accessToken`, `refreshToken`, `expiresIn` và lưu vào `token_cache.json`.
4. Mọi request sau đó gắn `Authorization: Bearer <accessToken>`.
5. Nếu token hết hạn ở lần sau thì tự động refresh, không cần OTP lại.

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
- **Manual flow:** Auth (token cache) → Market Data → Account Check → Place Order → Track/Cancel → Realtime Update.
- **Auto flow:** Auth (token cache) → OHLC/Indicator → Signal Engine → Risk Check → Auto Order → Monitor → P&L Logging.
 