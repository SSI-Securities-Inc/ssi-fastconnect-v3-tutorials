"""
Sample 8 — Kiểm tra trạng thái lệnh
======================================
Theo dõi tiến trình khớp của một lệnh cụ thể.

Luồng:
  1. Client gọi GET order by orderId
  2. API trả về status, filledQuantity, fills
  3. Đối chiếu lượng còn lại và trạng thái hiện tại
  4. Nếu chưa hoàn tất thì tiếp tục polling chu kỳ ngắn
  5. Khi FILLED/CANCELLED/REJECTED thì đóng vòng theo dõi
"""

import time

from ssi_sdk import Auth, Trading, Config
from ssi_sdk.enums import OrderSide, OrderStatus
from auth_helper import ensure_auth

config = Config(
    client_id="<CLIENT_ID>",
    api_key="<API_KEY>",
    api_secret="<API_SECRET>",
    private_key=("<PRIVATE_KEY_CONTENT>"),
    log_level="DEBUG",
)

ACCOUNT_NO = "<ACCOUNT_NO>"

# Trạng thái kết thúc — không cần polling thêm
TERMINAL_STATUSES = {
    OrderStatus.FILLED,
    OrderStatus.CANCELLED,
    OrderStatus.REJECTED,
    OrderStatus.EXPIRED,
    OrderStatus.PARTIAL_CANCELLED,
}

with Auth(config) as auth:
    ensure_auth(auth, otp="<OTP>")

    with Trading(auth) as trading:
        # --- Bước 1: Đặt một lệnh để theo dõi ---
        print("Đặt lệnh Limit mua SSI @ 26000...")
        result = trading.trading.place_limit_order(
            account_no=ACCOUNT_NO,
            symbol="SSI",
            side=OrderSide.BUY,
            quantity=100,
            price=26000,
        )
        print(f"  Kết quả đặt lệnh: {result}")

        # --- Bước 2-5: Polling trạng thái ---
        print("\n--- Bắt đầu theo dõi trạng thái lệnh ---")
        max_polls = 10
        poll_interval = 3  # giây

        for i in range(1, max_polls + 1):
            orders = trading.portfolio.get_today_orders(ACCOUNT_NO)

            # Tìm lệnh vừa đặt (lệnh mới nhất)
            if not orders:
                print(f"  Poll {i}: Chưa có lệnh trong sổ.")
                time.sleep(poll_interval)
                continue

            latest = orders[-1]
            remaining = latest.quantity - latest.filled_quantity - latest.cancel_quantity

            print(
                f"  Poll {i}: OrderID={latest.order_id} | "
                f"Status={latest.status.value} | "
                f"Khớp={latest.filled_quantity}/{latest.quantity} | "
                f"Còn lại={remaining}"
            )

            if latest.status in TERMINAL_STATUSES:
                print(f"\n→ Lệnh đã kết thúc với trạng thái: {latest.status.value}")
                if latest.filled_quantity > 0:
                    print(f"  Đã khớp: {latest.filled_quantity} cổ phiếu @ trung bình {latest.avg_price:,.0f}")
                break

            time.sleep(poll_interval)
        else:
            print(f"\nHết {max_polls} lần poll, lệnh vẫn đang mở.")
