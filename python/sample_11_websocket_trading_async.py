"""
Sample 11 (Async) — WebSocket trading real-time (trạng thái lệnh & danh mục)
===============================================================================
Nhận cập nhật tức thời về lệnh khớp và danh mục tài khoản (phiên bản async).

Luồng:
  1. Client mở kết nối WebSocket bằng token hợp lệ (cần OTP)
  2. Subscribe stream order_status và portfolio theo accountId
  3. Server push event khi trạng thái lệnh hoặc danh mục thay đổi
  4. Parse message theo loại (OrderStatus / OrderMatch / Portfolio)
  5. Khi mất kết nối, chạy cơ chế reconnect exponential backoff
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncStream
from auth_helper import ensure_auth_async
from helpers import enum_value
from config import config, ACCOUNT_NO, OTP
from ssi_sdk.models.streaming import (
    OrderStatusMessage,
    OrderMatchMessage,
    PortfolioMessage,
    HeartbeatMessage,
)

# --- Callback xử lý sự kiện trading ---
def on_trading_event(msg):
    if isinstance(msg, OrderStatusMessage):
        print(
            f"  [ORDER] {msg.symbol} {enum_value(msg.side)} | "
            f"OrderID: {msg.order_id} | Status: {enum_value(msg.status)} | "
            f"Khớp: {msg.filled_quantity}/{msg.quantity}"
        )
    elif isinstance(msg, OrderMatchMessage):
        # Khớp lệnh: cùng topic với order status, SDK tách thành OrderMatchMessage
        print(
            f"  [MATCH] {msg.symbol} {enum_value(msg.side)} | "
            f"OrderID: {msg.order_id} | Giá khớp: {msg.match_price} | KL khớp: {msg.match_qty}"
        )
    elif isinstance(msg, PortfolioMessage):
        # Chỉ có với tài khoản phái sinh. total_asset/cash_balance/stock_value đã deprecated
        # (server không gửi) nên không dùng; dùng các trường vị thế bên dưới.
        print(
            f"  [PORTFOLIO] {msg.symbol} | Long: {msg.long_qty} Short: {msg.short_qty} "
            f"Net: {msg.net} | Loại: {msg.position_type} | Floating P/L: {msg.floating_pl}"
        )
    else:
        print(f"  [TRADING] {msg}")


# --- Callback heartbeat ---
def on_heartbeat(msg: HeartbeatMessage):
    print(f"  [HEARTBEAT] {msg.status} - {msg.message}")


async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth, otp=OTP)

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
                await stream.streaming.wait(timeout_ms=300_000)  # 5 phút
            except KeyboardInterrupt:
                print("\nNgắt kết nối...")


asyncio.run(main())
