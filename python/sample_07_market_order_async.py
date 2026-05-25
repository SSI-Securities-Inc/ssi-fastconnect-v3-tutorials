"""
Sample 7 (Async) — Đặt lệnh Market (MP)
==========================================
Khớp lệnh nhanh theo giá thị trường hiện tại (phiên bản async).

Luồng:
  1. Client tạo order MARKET (không gửi price)
  2. Gửi request tới Trading Orders API
  3. Hệ thống match theo thanh khoản thị trường tại thời điểm gửi
  4. API trả về trạng thái khớp (FILLED hoặc PARTIALLY_FILLED)
  5. Cập nhật ngay danh mục/số dư tạm tính
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncTrading, Config
from ssi_sdk.enums import OrderSide
from auth_helper import ensure_auth_async

config = Config(
    client_id="<CLIENT_ID>",
    api_key="<API_KEY>",
    api_secret="<API_SECRET>",
    private_key=("<PRIVATE_KEY_CONTENT>"),
    log_level="DEBUG",
)

ACCOUNT_NO = "<ACCOUNT_NO>"


async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth, otp="222222")

        async with AsyncTrading(auth) as trading:
            # --- Bước 1: Kiểm tra sức mua/bán ở giá thị trường ---
            max_bs = await trading.trading.get_max_buy_sell_at_market_price(ACCOUNT_NO, "SSI")
            print(f"Max mua (market): {max_bs.max_buy_quantity} cổ phiếu")
            print(f"Max bán (market): {max_bs.max_sell_quantity} cổ phiếu")

            # --- Bước 2: Đặt lệnh Market mua ---
            print("\n--- Đặt lệnh MARKET mua SSI ---")
            result = await trading.trading.place_market_order(
                account_no=ACCOUNT_NO,
                symbol="SSI",
                side=OrderSide.BUY,
                quantity=100,
            )
            print(f"  Kết quả: {result}")

            # --- Bước 3: Kiểm tra trạng thái + cập nhật số dư (song song) ---
            orders_task = trading.portfolio.get_today_orders(ACCOUNT_NO)
            balance_task = trading.portfolio.get_equity_balance(ACCOUNT_NO)
            positions_task = trading.portfolio.get_equity_positions(ACCOUNT_NO)
            orders, balance, positions = await asyncio.gather(
                orders_task, balance_task, positions_task
            )

            print("\n--- Sổ lệnh hôm nay ---")
            for order in orders[-3:]:
                print(
                    f"  {order.order_id} | {order.symbol} {order.side.value} "
                    f"{order.order_type.value} | SL: {order.quantity} "
                    f"| Khớp: {order.filled_quantity} | Trạng thái: {order.status.value}"
                )

            # --- Bước 4: Số dư sau giao dịch ---
            print("\n--- Số dư sau giao dịch ---")
            print(f"  Tiền mặt khả dụng: {balance.available_cash:>15,.0f}")

            # --- Bước 5: Danh mục sau giao dịch ---
            print("\n--- Vị thế sau giao dịch ---")
            for pos in positions:
                if pos.symbol == "SSI":
                    print(f"  SSI | SL: {pos.quantity} | Giá vốn: {pos.cost_price:,.0f}")
        print("\n--- Vị thế sau giao dịch ---")
        for pos in positions:
            if pos.symbol == "SSI":
                print(f"  SSI | SL: {pos.quantity} | Giá vốn: {pos.cost_price:,.0f}")


asyncio.run(main())
