"""
Sample 10 (Async) — WebSocket dữ liệu thị trường real-time
=============================================================
Nhận tick data (giá khớp, bảng giá, room nước ngoài) tức thời (phiên bản async).

Luồng:
  1. Client mở kết nối WebSocket bằng token hợp lệ
  2. Subscribe các stream dữ liệu theo symbol / index
  3. Server push event khi có giao dịch mới hoặc bảng giá thay đổi
  4. Parse message theo loại (Trade / Quote / ForeignRoom)
  5. Khi mất kết nối, chạy cơ chế reconnect exponential backoff
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncStream
from auth_helper import ensure_auth_async
from config import config
from ssi_sdk.models.streaming import (
    TradeMessage,
    QuoteMessage,
    ForeignRoomMessage,
    HeartbeatMessage,
)


# --- Callback xử lý dữ liệu thị trường ---
def on_market_data(msg):
    if isinstance(msg, TradeMessage):
        print(
            f"  [TRADE] {msg.symbol} | Giá: {msg.price} "
            f"| KL: {msg.quantity} | Side: {msg.side}"
        )
    elif isinstance(msg, QuoteMessage):
        print(
            f"  [QUOTE] {msg.symbol} | "
            f"Bid: {msg.bid_prices[:3]} | Ask: {msg.ask_prices[:3]}"
        )
    elif isinstance(msg, ForeignRoomMessage):
        print(
            f"  [ROOM]  {msg.symbol} | "
            f"Room còn: {msg.current_room}/{msg.total_room}"
        )
    else:
        print(f"  [DATA]  {msg}")


# --- Callback heartbeat ---
def on_heartbeat(msg: HeartbeatMessage):
    print(f"  [HEARTBEAT] {msg.status} - {msg.message}")


async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth)  # Không cần OTP cho market data

        async with AsyncStream(auth) as stream:
            # --- Bước 1: Mở kết nối WebSocket ---
            print("Đang kết nối WebSocket...")
            await stream.streaming.connect()
            print("Đã kết nối!\n")

            # --- Bước 2: Đăng ký callback ---
            stream.streaming.on_data = on_market_data
            stream.streaming.on_heartbeat = on_heartbeat

            # --- Bước 3: Subscribe dữ liệu theo symbol ---
            print("Subscribing dữ liệu symbol...")
            await stream.streaming.subscribe_symbol(["SSI", "HPG", "VNM"])

            # --- Bước 4: Subscribe dữ liệu theo index ---
            print("Subscribing dữ liệu index...")
            await stream.streaming.subscribe_index(["VNINDEX", "HNX-INDEX"])

            # --- Bước 5: Lắng nghe liên tục ---
            print("\nĐang lắng nghe... (Ctrl+C để dừng)\n")
            try:
                await stream.streaming.wait(timeout=300)  # 5 phút
            except KeyboardInterrupt:
                print("\nNgắt kết nối...")


asyncio.run(main())
