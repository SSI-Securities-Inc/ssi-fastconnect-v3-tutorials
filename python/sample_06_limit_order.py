"""
Sample 6 — Đặt lệnh Limit (LO)
=================================
Đặt lệnh mua/bán tại mức giá chỉ định.

Luồng:
  1. Client tạo payload order (symbol, side, quantity, price, timeInForce)
  2. SDK tự gắn Idempotency-Key (clientRequestId) để chống submit trùng
  3. Gửi request tới Trading Orders API (có ký RSA)
  4. API trả về orderId và trạng thái ban đầu (PENDING)
  5. Lưu orderId để theo dõi khớp lệnh
"""

from ssi_sdk import Auth, Trading, Config
from auth_helper import ensure_auth
from ssi_sdk.enums import OrderSide

config = Config(
    client_id="<CLIENT_ID>",
    api_key="<API_KEY>",
    api_secret="<API_SECRET>",
    private_key=("<PRIVATE_KEY_CONTENT>"),
    log_level="DEBUG",
)

ACCOUNT_NO = "<ACCOUNT_NO>"

with Auth(config) as auth:
    ensure_auth(auth, otp="222222")

    with Trading(auth) as trading:
        # --- Bước 1: Kiểm tra sức mua trước ---
        max_bs = trading.trading.get_max_buy_sell(ACCOUNT_NO, "SSI", 26000)
        print(f"Sức mua tối đa SSI @ 26,000: {max_bs.max_buy_quantity} cổ phiếu")

        if max_bs.max_buy_quantity < 100:
            print("Không đủ sức mua, dừng lại.")
        else:
            # --- Bước 2: Đặt lệnh Limit mua ---
            print("\n--- Đặt lệnh LIMIT mua SSI ---")
            result = trading.trading.place_limit_order(
                account_no=ACCOUNT_NO,
                symbol="SSI",
                side=OrderSide.BUY,
                quantity=100,
                price=26000,
            )
            print(f"  Kết quả: {result}")

            # --- Bước 3: Đặt lệnh Limit bán ---
            print("\n--- Đặt lệnh LIMIT bán SSI ---")
            result = trading.trading.place_limit_order(
                account_no=ACCOUNT_NO,
                symbol="SSI",
                side=OrderSide.SELL,
                quantity=100,
                price=27000,
            )
            print(f"  Kết quả: {result}")

            # --- Bước 4: Kiểm tra lệnh vừa đặt trong sổ lệnh ---
            print("\n--- Sổ lệnh hôm nay ---")
            orders = trading.portfolio.get_today_orders(ACCOUNT_NO)
            for order in orders[-5:]:
                print(
                    f"  {order.order_id} | {order.symbol} {order.side.value} "
                    f"{order.order_type.value} | SL: {order.quantity} @ {order.price} "
                    f"| Trạng thái: {order.status.value}"
                )
