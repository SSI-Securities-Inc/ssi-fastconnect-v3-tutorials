"""
Sample 7 — Đặt lệnh Market (MP)
==================================
Khớp lệnh nhanh theo giá thị trường hiện tại.

Luồng:
  1. Client tạo order MARKET (không gửi price)
  2. Gửi request tới Trading Orders API
  3. Hệ thống match theo thanh khoản thị trường tại thời điểm gửi
  4. API trả về trạng thái khớp (FILLED hoặc PARTIALLY_FILLED)
  5. Cập nhật ngay danh mục/số dư tạm tính
"""

from ssi_sdk import Auth, Trading
from auth_helper import ensure_auth
from ssi_sdk.enums import OrderSide
from config import config, ACCOUNT_NO, OTP

with Auth(config) as auth:
    ensure_auth(auth, otp=OTP)

    with Trading(auth) as trading:
        # --- Bước 1: Kiểm tra sức mua/bán ở giá thị trường ---
        max_bs = trading.trading.get_max_buy_sell_at_market_price(ACCOUNT_NO, "SSI")
        print(f"Max mua (market): {max_bs.max_buy_quantity} cổ phiếu")
        print(f"Max bán (market): {max_bs.max_sell_quantity} cổ phiếu")

        # --- Bước 2: Đặt lệnh Market mua ---
        print("\n--- Đặt lệnh MARKET mua SSI ---")
        result = trading.trading.place_market_order(
            account_no=ACCOUNT_NO,
            symbol="SSI",
            side=OrderSide.BUY,
            quantity=100,
        )
        print(f"  Kết quả: {result}")

        # --- Bước 3: Kiểm tra trạng thái lệnh ---
        print("\n--- Sổ lệnh hôm nay ---")
        orders = trading.portfolio.get_today_orders(ACCOUNT_NO)
        for order in orders[-3:]:
            print(
                f"  {order.order_id} | {order.symbol} {order.side.value} "
                f"{order.order_type.value} | SL: {order.quantity} "
                f"| Khớp: {order.filled_quantity} | Trạng thái: {order.status.value}"
            )

        # --- Bước 4: Cập nhật lại số dư sau khi khớp ---
        print("\n--- Số dư sau giao dịch ---")
        balance = trading.portfolio.get_equity_balance(ACCOUNT_NO)
        print(f"  Tiền mặt khả dụng: {balance.account_balance:>15,.0f}")

    # --- Bước 5: Cập nhật danh mục ---
    print("\n--- Vị thế sau giao dịch ---")
    positions = trading.portfolio.get_equity_positions(ACCOUNT_NO)
    for pos in positions:
        if pos.symbol == "SSI":
            print(f"  SSI | SL: {pos.quantity} | Giá vốn: {pos.cost_price:,.0f}")
