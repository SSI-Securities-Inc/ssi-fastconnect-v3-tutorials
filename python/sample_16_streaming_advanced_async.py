"""
Sample 16 (Async) — Streaming nâng cao (SDK 3.3.0)
=====================================================
Các tính năng streaming mới ngoài sample 10/11:

  - on_connection: theo dõi connected / reconnected / disconnected / failed
    cùng is_connected, reconnect_count, last_error
  - Callback theo loại: on_market_flag, on_index_summary, on_order, on_order_match, on_portfolio
    (on_data / on_trading vẫn nhận mọi message như cũ)
  - subscribe_board (cả sàn), subscribe_index (các mã thành phần), subscribe_index_ticks (tick
    của chính chỉ số), subscribe_index_summary (số liệu tổng hợp của chỉ số)
  - Tên dễ hiểu: subscribe_trades / subscribe_quotes / subscribe_orders / subscribe_session_flag
  - list_subscription, unsubscribe_*, unsubscribe_all

Lưu ý:
  - Callback chạy trong task của SDK: giữ NGẮN, chỉ đếm/ghi nhận rồi để luồng chính xử lý.
    Sample này chỉ đếm message theo loại và in tổng kết, không in từng message
    (subscribe_board có thể tạo hàng nghìn message/giây).
  - Interval nến chỉ hỗ trợ tick, 1m, 5m (xem subscribe_candles). 1w/1M không có stream.
  - on_order / on_order_match / on_portfolio cần OTP và tài khoản trong config.json.
    Không có tài khoản thì sample chỉ chạy phần dữ liệu thị trường.
  - on_portfolio chỉ có sự kiện với tài khoản PHÁI SINH.
"""

import asyncio
from collections import Counter

from ssi_sdk import AsyncAuth, AsyncStream
from ssi_sdk.enums import Board
from ssi_sdk.models.streaming import (
    IndexSummaryMessage,
    MarketFlagMessage,
    OrderMatchMessage,
    OrderStatusMessage,
    PortfolioMessage,
)
from auth_helper import ensure_auth_async
from helpers import enum_value, mask_account
from config import config, ACCOUNT_NO, OTP

LISTEN_MS = 30_000  # nghe 30 giây rồi dọn dẹp

counts: Counter = Counter()


# --- Callback: giữ ngắn, chỉ đếm ---
def on_connection(event):
    # state: connected | reconnected | disconnected | failed
    err = f" | lỗi: {type(event.error).__name__}" if event.error else ""
    print(f"  [CONN] {event.state} (reconnect_count={event.reconnect_count}){err}")


def on_data(msg):
    counts[type(msg).__name__] += 1


def on_market_flag(msg: MarketFlagMessage):
    print(f"  [FLAG] {msg.board} | {msg.flag} | {msg.trading_time}")


def on_index_summary(msg: IndexSummaryMessage):
    counts["IndexSummary"] += 1
    if counts["IndexSummary"] == 1:  # chỉ in lần đầu để xem cấu trúc
        print(f"  [INDEX-SUMMARY] {msg.index} | số trường: {len(msg.data)}")


def on_order(msg: OrderStatusMessage):
    print(
        f"  [ORDER] {msg.symbol} {enum_value(msg.side)} {enum_value(msg.order_type)} "
        f"| {enum_value(msg.status)} | khớp {msg.filled_quantity}/{msg.quantity} "
        f"| modified_time={msg.modified_time}"  # modify_time đã deprecated
    )


def on_order_match(msg: OrderMatchMessage):
    print(f"  [MATCH] {msg.symbol} {enum_value(msg.side)} | giá khớp {msg.match_price} | KL {msg.match_qty}")


def on_portfolio(msg: PortfolioMessage):
    print(f"  [PORTFOLIO] {msg.symbol} | long={msg.long_qty} short={msg.short_qty} net={msg.net}")


async def main():
    has_account = bool(ACCOUNT_NO) and not ACCOUNT_NO.startswith("<")

    async with AsyncAuth(config) as auth:
        # Dữ liệu thị trường không cần OTP; kênh trading (lệnh/danh mục) cần OTP
        if has_account:
            await ensure_auth_async(auth, otp=OTP)
        else:
            await ensure_auth_async(auth)

        async with AsyncStream(auth) as stream:
            ws = stream.streaming

            # --- Bước 1: Đăng ký callback TRƯỚC khi connect để không lỡ sự kiện "connected" ---
            ws.on_connection = on_connection
            ws.on_data = on_data
            ws.on_market_flag = on_market_flag
            ws.on_index_summary = on_index_summary
            ws.on_order = on_order
            ws.on_order_match = on_order_match
            ws.on_portfolio = on_portfolio

            # --- Bước 2: Kết nối và xem trạng thái ---
            await ws.connect()
            print(f"is_connected={ws.is_connected} reconnect_count={ws.reconnect_count} last_error={ws.last_error}")

            # --- Bước 3: Subscribe ---
            await ws.subscribe_trades(["SSI", "HPG"])          # alias của subscribe_symbol_trade
            await ws.subscribe_quotes(["SSI"])                 # alias của subscribe_symbol_quote
            await ws.subscribe_board([Board.HNX])              # toàn bộ trade/quote/room của sàn HNX (nhiều!)
            await ws.subscribe_index(["VN30"])                 # dữ liệu các MÃ THÀNH PHẦN của VN30
            await ws.subscribe_index_ticks(["VNINDEX"])        # tick của CHÍNH chỉ số (IndexTickMessage)
            await ws.subscribe_index_summary(["VN30"])         # IndexSummaryMessage
            await ws.subscribe_session_flag()                  # cờ phiên ATO/LO/ATC (alias subscribe_market_flag)
            await ws.subscribe_candles(["SSI"], "1m")          # nến 1 phút; chỉ tick/1m/5m

            if has_account:
                print(f"Subscribe lệnh/danh mục của tài khoản {mask_account(ACCOUNT_NO)}")
                await ws.subscribe_orders(ACCOUNT_NO)          # trạng thái lệnh + khớp lệnh
                await ws.subscribe_positions(ACCOUNT_NO)       # danh mục (chỉ phái sinh có event)

            # --- Bước 4: Lắng nghe ---
            print(f"\nĐang lắng nghe {LISTEN_MS // 1000} giây...\n")
            try:
                await ws.wait(timeout_ms=LISTEN_MS)
            except KeyboardInterrupt:
                print("\nNgắt theo yêu cầu...")

            print("\nSố message nhận được theo loại:")
            for name, n in counts.most_common():
                print(f"  {name:<22} {n}")

            # --- Bước 5: Xem server đang giữ những subscription nào ---
            subs = await ws.list_subscription()  # {"trading": [...], "data": [...]}
            print(f"\nlist_subscription: data={len(subs['data'])} topic, trading={len(subs['trading'])} topic")

            # --- Bước 6: Hủy từng phần rồi hủy tất cả ---
            await ws.unsubscribe_board([Board.HNX])  # board nặng nhất: hủy trước
            await ws.unsubscribe_index_ticks(["VNINDEX"])
            removed = await ws.unsubscribe_all()     # hỏi server rồi hủy đúng những gì còn lại
            print(f"unsubscribe_all đã gỡ: data={len(removed['data'])}, trading={len(removed['trading'])}")

            print(f"\nKết thúc: is_connected={ws.is_connected} reconnect_count={ws.reconnect_count} last_error={ws.last_error}")


if __name__ == "__main__":
    asyncio.run(main())
