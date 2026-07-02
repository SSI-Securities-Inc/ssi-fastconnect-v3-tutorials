"""
Sample 9 (Async) — Hủy lệnh
==============================
Dừng phần khối lượng chưa khớp của lệnh đang mở (phiên bản async).

Luồng:
  1. User bấm hủy trên lệnh đang PENDING/PARTIALLY_FILLED
  2. Client gửi DELETE order kèm thông tin account/symbol
  3. API xác thực quyền và trạng thái lệnh hiện tại
  4. Nếu hợp lệ, hệ thống cập nhật CANCELLED cho phần chưa khớp
  5. Đồng bộ lại sổ lệnh và số lượng còn treo
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncTrading, Config
from ssi_sdk.enums import OrderStatus
from auth_helper import ensure_auth_async

config = Config(
    client_id="<CLIENT_ID>",
    api_key="<API_KEY>",
    api_secret="<API_SECRET>",
    private_key=("<PRIVATE_KEY_CONTENT>"),
    log_level="DEBUG",
)

ACCOUNT_NO = "<ACCOUNT_NO>"
CANCELLABLE_STATUSES = {
    OrderStatus.PENDING_APPROVAL,
    OrderStatus.READY,
    OrderStatus.SENT,
    OrderStatus.QUEUED,
    OrderStatus.PARTIAL_FILLED,
}


async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth, otp="<OTP>")

        async with AsyncTrading(auth) as trading:
            # --- Bước 1: Lấy sổ lệnh, tìm lệnh đang mở ---
            print("--- Sổ lệnh hôm nay ---")
            orders = await trading.portfolio.get_today_orders(ACCOUNT_NO)

            open_orders = [o for o in orders if o.status in CANCELLABLE_STATUSES]
            print(f"Tổng lệnh: {len(orders)} | Lệnh đang mở: {len(open_orders)}\n")

            if not open_orders:
                print("Không có lệnh nào đang mở để hủy.")
            else:
                for order in open_orders:
                    remaining = order.quantity - order.filled_quantity
                    print(
                        f"  OrderID: {order.order_id} | {order.symbol} {order.side.value} "
                        f"{order.order_type.value} | SL: {order.quantity} @ {order.price} "
                        f"| Khớp: {order.filled_quantity} | Còn: {remaining} "
                        f"| Status: {order.status.value}"
                    )

                # --- Bước 2: Hủy tất cả lệnh đang mở (song song) ---
                print(f"\n--- Hủy {len(open_orders)} lệnh đang mở ---")
                cancel_tasks = [
                    trading.trading.cancel_order_by_order_id(
                        account_no=ACCOUNT_NO,
                        order_id=order.order_id,
                    )
                    for order in open_orders
                ]
                results = await asyncio.gather(*cancel_tasks, return_exceptions=True)

                for order, result in zip(open_orders, results):
                    if isinstance(result, Exception):
                        print(f"  {order.order_id}: LỖI - {result}")
                    else:
                        print(f"  {order.order_id}: OK - {result}")

                # --- Bước 3: Xác nhận + cập nhật số dư (song song) ---
                orders_after, balance = await asyncio.gather(
                    trading.portfolio.get_today_orders(ACCOUNT_NO),
                    trading.portfolio.get_equity_balance(ACCOUNT_NO),
                )

                print("\n--- Kiểm tra sổ lệnh sau hủy ---")
                for order in orders_after:
                    if order.order_id in {o.order_id for o in open_orders}:
                        print(
                            f"  OrderID: {order.order_id} | "
                            f"Status: {order.status.value} | "
                            f"Khớp: {order.filled_quantity} | "
                            f"Đã hủy: {order.cancel_quantity}"
                        )

                print("\n--- Số dư sau hủy ---")
                print(f"  Tiền mặt khả dụng: {balance.available_cash:>15,.0f}")


asyncio.run(main())
