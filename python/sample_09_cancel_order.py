"""
Sample 9 — Hủy lệnh
=====================
Dừng phần khối lượng chưa khớp của lệnh đang mở.

Luồng:
  1. User bấm hủy trên lệnh đang PENDING/PARTIALLY_FILLED
  2. Client gửi DELETE order kèm thông tin account/symbol
  3. API xác thực quyền và trạng thái lệnh hiện tại
  4. Nếu hợp lệ, hệ thống cập nhật CANCELLED cho phần chưa khớp
  5. Đồng bộ lại sổ lệnh và số lượng còn treo
"""

from ssi_sdk import Auth, Trading, Config
from ssi_sdk.enums import OrderStatus
from auth_helper import ensure_auth

config = Config(
    client_id="<CLIENT_ID>",
    api_key="<API_KEY>",
    api_secret="<API_SECRET>",
    private_key=("<PRIVATE_KEY_CONTENT>"),
    log_level="DEBUG",
)

ACCOUNT_NO = "<ACCOUNT_NO>"
# Trạng thái có thể hủy
CANCELLABLE_STATUSES = {
    OrderStatus.PENDING_APPROVAL,
    OrderStatus.READY,
    OrderStatus.SENT,
    OrderStatus.QUEUED,
    OrderStatus.PARTIAL_FILLED,
}

with Auth(config) as auth:
    ensure_auth(auth, otp="222222")

    with Trading(auth) as trading:
        # --- Bước 1: Lấy sổ lệnh, tìm lệnh đang mở ---
        print("--- Sổ lệnh hôm nay ---")
        orders = trading.portfolio.get_today_orders(ACCOUNT_NO)

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

            # --- Bước 2: Hủy lệnh đầu tiên trong danh sách ---
            target = open_orders[0]
            print(f"\n--- Hủy lệnh: {target.order_id} ---")

            result = trading.trading.cancel_order_by_order_id(
                account_no=ACCOUNT_NO,
                order_id=target.order_id,
            )
            print(f"  Kết quả hủy: {result}")

            # --- Bước 3: Xác nhận trạng thái sau hủy ---
            print("\n--- Kiểm tra sổ lệnh sau hủy ---")
            orders_after = trading.portfolio.get_today_orders(ACCOUNT_NO)
            for order in orders_after:
                if order.order_id == target.order_id:
                    print(
                        f"  OrderID: {order.order_id} | "
                        f"Status: {order.status.value} | "
                        f"Khớp: {order.filled_quantity} | "
                        f"Đã hủy: {order.cancel_quantity}"
                    )
                    break

    # --- Bước 4: Cập nhật lại số dư ---
    print("\n--- Số dư sau hủy ---")
    balance = trading.portfolio.get_equity_balance(ACCOUNT_NO)
    print(f"  Tiền mặt khả dụng: {balance.available_cash:>15,.0f}")
