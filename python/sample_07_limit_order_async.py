"""
Sample 6 (Async) — Đặt lệnh Limit (LO)
=========================================
Đặt lệnh mua/bán tại mức giá chỉ định (phiên bản async).

Luồng:
  1. Client tạo payload order (symbol, side, quantity, price, timeInForce)
  2. SDK tự gắn Idempotency-Key (clientRequestId) để chống submit trùng
  3. Gửi request tới Trading Orders API (có ký RSA)
  4. API trả về orderId và trạng thái ban đầu (PENDING)
  5. Lưu orderId để theo dõi khớp lệnh
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
        await ensure_auth_async(auth, otp="<OTP>")

        async with AsyncTrading(auth) as trading:
            # --- Bước 1: Kiểm tra sức mua trước ---
            max_bs = await trading.trading.get_max_buy_sell(ACCOUNT_NO, "SSI", 26000)
            print(f"Sức mua tối đa SSI @ 26,000: {max_bs.max_buy_quantity} cổ phiếu")

            if max_bs.max_buy_quantity < 100:
                print("Không đủ sức mua, dừng lại.")
            else:
                # --- Bước 2: Đặt song song lệnh Limit mua + bán ---
                print("\n--- Đặt lệnh LIMIT mua + bán SSI ---")
                buy_task = trading.trading.place_limit_order(
                    account_no=ACCOUNT_NO,
                    symbol="SSI",
                    side=OrderSide.BUY,
                    quantity=100,
                    price=26000,
                )
                sell_task = trading.trading.place_limit_order(
                    account_no=ACCOUNT_NO,
                    symbol="SSI",
                    side=OrderSide.SELL,
                    quantity=100,
                    price=27000,
                )
                buy_result, sell_result = await asyncio.gather(buy_task, sell_task)
                print(f"  Lệnh mua: {buy_result}")
                print(f"  Lệnh bán: {sell_result}")

                # --- Bước 3: Kiểm tra lệnh vừa đặt trong sổ lệnh ---
                print("\n--- Sổ lệnh hôm nay ---")
                orders = await trading.portfolio.get_today_orders(ACCOUNT_NO)
                for order in orders[-5:]:
                    print(
                        f"  {order.order_id} | {order.symbol} {order.side.value} "
                        f"{order.order_type.value} | SL: {order.quantity} @ {order.price} "
                        f"| Trạng thái: {order.status.value}"
                    )


asyncio.run(main())
