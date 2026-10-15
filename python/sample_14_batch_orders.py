"""
Sample 14 — Đặt / hủy nhiều lệnh trong 1 request (Batch Orders)
=================================================================
SDK 3.3.0: place_batch_orders và cancel_batch_orders gửi nhiều lệnh trong MỘT request đã ký RSA.

CẢNH BÁO: sample này gửi lệnh THẬT. Để an toàn, mọi lệnh là LO MUA tại GIÁ SÀN
(xa thị trường, chỉ nằm chờ) với khối lượng nhỏ, và script yêu cầu gõ YES trước khi gửi.

Luồng:
  1. Lấy giá sàn từ master data, dựng danh sách BatchPlaceOrderItem
  2. Kiểm tra số lệnh <= Config.max_batch_orders (mặc định 20)
  3. place_batch_orders → xử lý kết quả TỪNG lệnh (BatchOrderResult.success/error_code)
  4. cancel_batch_orders cho các lệnh vừa đặt thành công (BatchCancelOrderItem)

Lưu ý quan trọng:
  - Server validate cả batch trước: 1 lệnh sai → cả batch bị từ chối (ValidationError/APIError).
  - client_request_id (<= 20 ký tự) là khóa idempotency của TỪNG lệnh. Tự đặt để còn tra cứu/hủy.
  - batch bị gửi lại với id đã dùng trong ngày → DuplicateRequestError (HTTP 409). SDK không
    tự retry request POST/DELETE; lệnh gốc có thể đã nằm trên sàn, hãy kiểm tra sổ lệnh trước.
"""

import uuid
from decimal import Decimal

from ssi_sdk import Auth, Data, Trading
from ssi_sdk.enums import OrderSide, OrderType
from ssi_sdk.exceptions import APIError, DuplicateRequestError, ValidationError
from ssi_sdk.models import BatchCancelOrderItem, BatchPlaceOrderItem
from auth_helper import ensure_auth
from helpers import confirm_real_orders, enum_value, mask_account, require_account
from config import config, ACCOUNT_NO, OTP

SYMBOL = "SSI"
QUANTITY = 100
NUM_ORDERS = 3

require_account(ACCOUNT_NO)
confirm_real_orders(f"{NUM_ORDERS} lệnh LO mua {SYMBOL} (batch)")

with Auth(config) as auth:
    ensure_auth(auth, otp=OTP)

    # --- Bước 1: Lấy giá sàn làm giá đặt (lệnh nằm chờ, xa thị trường) ---
    with Data(auth) as data:
        master = next((m for m in data.market_data.get_master_data() if m.symbol == SYMBOL), None)
    if master is None or master.floor <= 0:
        raise SystemExit(f"Không có giá sàn của {SYMBOL}, dừng để an toàn.")
    floor_price = Decimal(str(master.floor))
    print(f"Giá sàn {SYMBOL}: {floor_price}")

    with Trading(auth) as trading:
        # --- Bước 2: Dựng batch ---
        batch_tag = uuid.uuid4().hex[:8]
        place_items: list[BatchPlaceOrderItem] = [
            {
                "account_no": ACCOUNT_NO,
                "symbol": SYMBOL,
                "side": OrderSide.BUY,
                "quantity": QUANTITY,
                "price": floor_price,
                "order_type": OrderType.LO,
                "client_request_id": f"bt{batch_tag}{i:02d}",  # <= 20 ký tự
            }
            for i in range(1, NUM_ORDERS + 1)
        ]
        if len(place_items) > config.max_batch_orders:
            raise SystemExit(f"Tối đa {config.max_batch_orders} lệnh / batch (Config.max_batch_orders).")

        # --- Bước 3: Đặt batch ---
        print(f"\n--- place_batch_orders: {len(place_items)} lệnh, tài khoản {mask_account(ACCOUNT_NO)} ---")
        try:
            response = trading.trading.place_batch_orders(place_items)
        except ValidationError as e:
            raise SystemExit(f"Batch không hợp lệ (chưa gửi gì): {e}")
        except DuplicateRequestError as e:
            raise SystemExit(f"Batch id đã dùng trong ngày (409), kiểm tra sổ lệnh trước khi gửi lại: {e}")
        except APIError as e:
            raise SystemExit(f"Server từ chối batch: code={e.code} status={e.status_code} msg={e.message}")

        print(f"Batch ID: {response.batch_request_id}")
        placed_ok: list[str] = []
        for r in response.results:
            if r.success:
                placed_ok.append(r.order_id)
                print(f"  OK   {r.client_request_id} -> orderId={r.order_id} status={enum_value(r.status)}")
            else:
                print(f"  FAIL {r.client_request_id} -> code={r.error_code} msg={r.error_message}")

        # --- Bước 4: Hủy batch các lệnh vừa đặt ---
        if not placed_ok:
            print("\nKhông có lệnh nào được đặt thành công, bỏ qua bước hủy.")
        else:
            # Mỗi item chỉ được có ĐÚNG MỘT trong order_id / client_request_id
            cancel_items: list[BatchCancelOrderItem] = [
                {"account_no": ACCOUNT_NO, "order_id": oid} for oid in placed_ok
            ]
            print(f"\n--- cancel_batch_orders: {len(cancel_items)} lệnh ---")
            try:
                cancel_resp = trading.trading.cancel_batch_orders(cancel_items)
            except (ValidationError, DuplicateRequestError, APIError) as e:
                raise SystemExit(f"Hủy batch thất bại: {type(e).__name__}: {e}")
            for r in cancel_resp.results:
                state = "OK  " if r.success else "FAIL"
                print(
                    f"  {state} orderId={r.order_id} cancelId={r.client_cancel_id} "
                    f"status={enum_value(r.status)} {r.error_code} {r.error_message}".rstrip()
                )
