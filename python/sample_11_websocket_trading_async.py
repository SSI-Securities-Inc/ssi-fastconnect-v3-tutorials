"""
Sample 10b (Async) — WebSocket trading real-time (trạng thái lệnh & danh mục)
===============================================================================
Nhận cập nhật tức thời về lệnh khớp và danh mục tài khoản (phiên bản async).

Luồng:
  1. Client mở kết nối WebSocket bằng token hợp lệ (cần OTP)
  2. Subscribe stream order_status và portfolio theo accountId
  3. Server push event khi trạng thái lệnh hoặc danh mục thay đổi
  4. Parse message theo loại (OrderStatus / Portfolio)
  5. Khi mất kết nối, chạy cơ chế reconnect exponential backoff
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncStream, Config
from auth_helper import ensure_auth_async
from ssi_sdk.models.streaming import (
    OrderStatusMessage,
    PortfolioMessage,
    HeartbeatMessage,
)

config = Config(
    client_id="<CLIENT_ID>",
    api_key="<API_KEY>",
    api_secret="<API_SECRET>",
    private_key=("<PRIVATE_KEY_CONTENT>"),
    log_level="DEBUG",
)

ACCOUNT_NO = "<ACCOUNT_NO>"

# --- Callback xử lý sự kiện trading ---
def on_trading_event(msg):
    if isinstance(msg, OrderStatusMessage):
        print(
            f"  [ORDER] {msg.symbol} {msg.side} | "
            f"OrderID: {msg.order_id} | Status: {msg.status} | "
            f"Khớp: {msg.filled_quantity}/{msg.quantity}"
        )
    elif isinstance(msg, PortfolioMessage):
        print(
            f"  [PORTFOLIO] Account: {msg.account_no} | "
            f"Tổng TS: {msg.total_asset} | Cash: {msg.cash_balance}"
        )
    else:
        print(f"  [TRADING] {msg}")


# --- Callback heartbeat ---
def on_heartbeat(msg: HeartbeatMessage):
    print(f"  [HEARTBEAT] {msg.status} - {msg.message}")


async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth, otp="<OTP>")

        async with AsyncStream(auth) as stream:
            # --- Bước 1: Mở kết nối WebSocket ---
            print("Đang kết nối WebSocket...")
            await stream.streaming.connect()
            print("Đã kết nối!\n")

            # --- Bước 2: Đăng ký callback ---
            stream.streaming.on_trading = on_trading_event
            stream.streaming.on_heartbeat = on_heartbeat

            # --- Bước 3: Subscribe trạng thái lệnh real-time ---
            print("Subscribing trạng thái lệnh...")
            await stream.streaming.subscribe_order_status(ACCOUNT_NO)

            # # --- Bước 4: Subscribe danh mục tài khoản real-time ---
            # print("Subscribing danh mục tài khoản...")
            # await stream.streaming.subscribe_portfolio(ACCOUNT_NO)

            # --- Bước 5: Lắng nghe liên tục ---
            print("\nĐang lắng nghe... (Ctrl+C để dừng)\n")
            try:
                await stream.streaming.wait(timeout=300)  # 5 phút
            except KeyboardInterrupt:
                print("\nNgắt kết nối...")


asyncio.run(main())
