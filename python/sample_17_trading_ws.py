"""
Sample 17 — Trading WebSocket (THỬ NGHIỆM / EXPERIMENTAL)
============================================================
Đặt/hủy lệnh và truy vấn qua WebSocket trading (/ws/v3/trading) thay vì REST.

!!! EXPERIMENTAL !!!
  - API nằm trong `ssi_sdk.experimental`, KHÔNG thuộc facade ổn định (Trading/Stream) và có thể
    đổi mà không báo trước.
  - Phía server chưa được công bố là đã phát hành: chỉ chạy trên môi trường bạn được xác nhận
    có hỗ trợ trading WebSocket. Nếu không, bước kết nối sẽ lỗi (TradingWSError/WebSocketError).
  - Token cần các scope tradingws:order:post / :delete và tradingws:position:get /
    :orderBook:get. Thiếu scope → server trả 403.

CẢNH BÁO: sample này gửi lệnh THẬT (LO MUA tại GIÁ SÀN, xa thị trường, rồi hủy ngay) và yêu cầu
gõ YES trước khi gửi.

Luồng:
  1. Kết nối trading_ws(auth) (cùng token và private key với REST)
  2. Gắn client.on_event để nhận sự kiện lệnh (frame {"channel","topic","data"})
  3. subscribe_order_events(account) → đặt lệnh LO → nhận ack PD → hủy lệnh
  4. query_position (có dòng tổng hợp querySummary) và query_order_book
  5. Đọc client.rate_limit; xử lý TradingWSError (429/403/401...)

Lưu ý:
  - place_order trả về ack trạng thái PD; kết quả thật đến qua sự kiện lệnh, không phải qua ack.
  - SDK tự chờ khi quota về 0 (respect_rate_limit, tối đa max_rate_limit_wait_ms). Quota: 50 lệnh /
    cửa sổ cho lệnh, 100 cho truy vấn.
  - on_event chạy trên thread đọc của SDK: giữ ngắn, KHÔNG in nguyên frame (topic chứa số tài khoản).
"""

import time
import uuid
from decimal import Decimal

from ssi_sdk import Auth, Data
from ssi_sdk.enums import OrderSide, OrderType
from ssi_sdk.exceptions import ValidationError, WebSocketError
from ssi_sdk.experimental import TradingWSError, trading_ws
from auth_helper import ensure_auth
from helpers import confirm_real_orders, enum_value, mask_account, require_account
from config import config, ACCOUNT_NO, OTP

SYMBOL = "SSI"
QUANTITY = 100

events: list[dict] = []


def on_event(frame: dict):
    # Chỉ giữ loại sự kiện và trạng thái, không giữ topic/số tài khoản
    data = frame.get("data") if isinstance(frame.get("data"), dict) else {}
    events.append({"eventType": data.get("eventType"), "status": data.get("orderStatus")})
    print(f"  [EVENT] {data.get('eventType')} | status={data.get('orderStatus')}")


def describe(result) -> str:
    if isinstance(result, dict):
        return "dict: " + ", ".join(
            f"{k}[{len(v)}]" if isinstance(v, list) else str(k) for k, v in result.items()
        )
    if isinstance(result, list):
        return f"list[{len(result)}]"
    return type(result).__name__


def show_rate_limit(ws, label: str):
    rl = ws.client.rate_limit
    print(f"  [RATE] {label}: limit={rl.get('limit')} remaining={rl.get('remaining')} resetMs={rl.get('resetMs')}")


require_account(ACCOUNT_NO)
confirm_real_orders(f"1 lệnh LO mua {SYMBOL} qua trading WebSocket (EXPERIMENTAL)")

with Auth(config) as auth:
    ensure_auth(auth, otp=OTP)

    # Giá sàn qua REST để đặt lệnh nằm chờ, xa thị trường
    with Data(auth) as data:
        master = next((m for m in data.market_data.get_master_data() if m.symbol == SYMBOL), None)
    if master is None or master.floor <= 0:
        raise SystemExit(f"Không có giá sàn của {SYMBOL}, dừng để an toàn.")
    floor_price = Decimal(str(master.floor))

    # --- Bước 1: Kết nối ---
    print("Đang kết nối trading WebSocket (EXPERIMENTAL)...")
    try:
        ws = trading_ws(auth)  # tùy chọn: respect_rate_limit, max_rate_limit_wait_ms, request_timeout_ms...
    except (TradingWSError, WebSocketError) as e:
        raise SystemExit(f"Không kết nối được trading WS (môi trường có hỗ trợ chưa?): {type(e).__name__}: {e}")

    try:
        # --- Bước 2: Nhận sự kiện lệnh ---
        ws.client.on_event = on_event
        ws.subscribe_order_events(ACCOUNT_NO)
        print(f"Đã subscribe sự kiện lệnh tài khoản {mask_account(ACCOUNT_NO)}")

        # --- Bước 3: Đặt lệnh rồi hủy ---
        try:
            ack = ws.place_order(
                account_no=ACCOUNT_NO,
                symbol=SYMBOL,
                side=OrderSide.BUY,
                quantity=QUANTITY,
                price=floor_price,
                order_type=OrderType.LO,
                client_request_id=f"ws{uuid.uuid4().hex[:12]}",  # <= 20 ký tự
            )
            print(f"Ack đặt lệnh: orderId={ack.order_id} status={enum_value(ack.status)} (PD = đang chờ xử lý)")
            show_rate_limit(ws, "sau place_order")

            time.sleep(2)  # chờ sự kiện lệnh về on_event
            cancel = ws.cancel_order(account_no=ACCOUNT_NO, order_id=ack.order_id)
            print(f"Ack hủy lệnh: status={enum_value(cancel.status)}")
            time.sleep(2)
        except ValidationError as e:
            print(f"Dữ liệu không hợp lệ (chưa gửi gì): {e}")
        except TradingWSError as e:
            # status: 401 token hết hạn | 403 thiếu scope | 429 hết quota | 5xx lỗi server
            print(f"Server từ chối lệnh: status={e.status} code={e.error_code} msg={e.server_message}")
            if e.status == 429:
                print("  Hết quota: chờ resetMs trong client.rate_limit rồi mới gửi tiếp.")

        # --- Bước 4: Truy vấn ---
        try:
            pos = ws.query_position(ACCOUNT_NO, query_summary=True)
            print(f"\nquery_position: status={pos.status} result={describe(pos.result)}")
            book = ws.query_order_book(ACCOUNT_NO, symbol=SYMBOL, page_size=10)
            print(f"query_order_book: status={book.status} result={describe(book.result)}")
            show_rate_limit(ws, "sau truy vấn")
        except TradingWSError as e:
            print(f"Truy vấn bị từ chối: status={e.status} code={e.error_code} msg={e.server_message}")

        print(f"\nTổng số sự kiện nhận được: {len(events)}")
        ws.unsubscribe_all_events()
    finally:
        ws.close()  # luôn đóng socket
