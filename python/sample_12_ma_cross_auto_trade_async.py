"""
Sample 12 (Async) — MA Cross Signal Real-time (WebSocket bars)
==============================================================
Kết hợp WebSocket async để nhận tick real-time, tự aggregate thành nến,
tính MA5/MA10, đặt lệnh khi giao cắt và theo dõi lệnh qua stream.

Luồng:
  1. Seed dữ liệu lịch sử 5m để khởi tạo MA
  2. Mở WebSocket, subscribe trade stream (market data) và order stream (trading)
  3. Mỗi TradeMessage → cập nhật nến hiện tại (OHLCV)
  4. Khi nến đóng (chuyển bucket thời gian) → tính MA5/MA10 → kiểm tra giao cắt
  5. Có tín hiệu + không có lệnh đang chờ → kiểm tra risk → đặt lệnh MARKET
  6. OrderStatusMessage phản hồi trạng thái lệnh real-time, không polling
  7. SDK ping keepalive định kỳ để giữ kết nối WebSocket
"""

import asyncio
import time
from collections import deque
from dataclasses import dataclass

from ssi_sdk import AsyncAuth, AsyncData, AsyncStream, AsyncTrading
from ssi_sdk.enums import OrderSide, OrderStatus
from ssi_sdk.models.streaming import TradeMessage, OrderStatusMessage, HeartbeatMessage
from auth_helper import ensure_auth_async
from config import config, ACCOUNT_NO, OTP

SYMBOL = "SSI"
MA_FAST = 5
MA_SLOW = 10
QUANTITY = 100
BAR_SECONDS = 300   # Nến 5 phút

TERMINAL_STATUSES = {
    OrderStatus.FILLED,
    OrderStatus.CANCELLED,
    OrderStatus.REJECTED,
    OrderStatus.EXPIRED,
    OrderStatus.PARTIAL_CANCELLED,
}


# ---------------------------------------------------------------------------
# Bar builder — aggregate TradeMessage thành nến OHLCV theo bucket thời gian
# ---------------------------------------------------------------------------

@dataclass
class Bar:
    ts: int        # Unix timestamp mở nến (bucket)
    open: float
    high: float
    low: float
    close: float
    volume: int


class BarBuilder:
    """Nhận từng trade tick, tự tổng hợp thành nến fixed-interval."""

    def __init__(self, interval_seconds: int):
        self.interval = interval_seconds
        self._lock = asyncio.Lock()
        self._current: Bar | None = None
        self.closed: deque[Bar] = deque(maxlen=200)

    def seed(self, historical_bars) -> None:
        """Nạp dữ liệu lịch sử để khởi tạo MA (gọi trước event loop)."""
        for b in historical_bars:
            self.closed.append(
                Bar(0, b.open_price, b.high_price, b.low_price, b.close_price, b.volume)
            )
        print(f"  Seeded {len(self.closed)} historical bars")

    async def on_trade(self, price: float, quantity: int) -> Bar | None:
        """
        Xử lý một tick trade. Trả về Bar vừa đóng nếu chuyển bucket,
        None nếu nến hiện tại chưa đóng.
        """
        bucket = (int(time.time()) // self.interval) * self.interval
        async with self._lock:
            if self._current is None or self._current.ts != bucket:
                closed = self._current
                self._current = Bar(bucket, price, price, price, price, quantity)
                if closed is not None:
                    self.closed.append(closed)
                    return closed
                return None
            b = self._current
            b.high = max(b.high, price)
            b.low = min(b.low, price)
            b.close = price
            b.volume += quantity
            return None

    def snapshot(self) -> list[Bar]:
        """Lấy toàn bộ bars (closed + current) dưới dạng list."""
        result = list(self.closed)
        if self._current is not None:
            result.append(self._current)
        return result


# ---------------------------------------------------------------------------
# MA / Signal helpers
# ---------------------------------------------------------------------------

def calculate_ma(bars: list[Bar], period: int) -> float | None:
    if len(bars) < period:
        return None
    return sum(b.close for b in bars[-period:]) / period


def detect_cross(bars: list[Bar], fast: int, slow: int) -> str | None:
    if len(bars) < slow + 1:
        return None
    mf_now = calculate_ma(bars, fast)
    ms_now = calculate_ma(bars, slow)
    mf_prev = calculate_ma(bars[:-1], fast)
    ms_prev = calculate_ma(bars[:-1], slow)
    if None in (mf_now, ms_now, mf_prev, ms_prev):
        return None
    if mf_prev <= ms_prev and mf_now > ms_now:
        return "BUY"
    if mf_prev >= ms_prev and mf_now < ms_now:
        return "SELL"
    return None


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth, otp=OTP)

        # ===== Bước 1: Seed lịch sử 5m =====
        builder = BarBuilder(BAR_SECONDS)
        async with AsyncData(auth) as data:
            print(f"--- Load lịch sử OHLC 5m ({SYMBOL}) ---")
            hist = await data.market_data.get_ohlc_5minute_historical(
                symbol=SYMBOL,
                from_date="2026/01/01 00:00:00",
                to_date="2026/05/12 23:59:59",
                page=1,
                size=MA_SLOW + 5,
            )
            builder.seed(hist)

        # ===== State dùng chung giữa callbacks =====
        state = {
            "active_order_id": None,   # Lệnh đang chờ khớp
            "last_signal": None,       # Tránh lặp tín hiệu cùng chiều
        }
        state_lock = asyncio.Lock()

        async with AsyncTrading(auth) as trading:

            # ===== Bước 2: Callbacks =====

            async def on_market_data(msg):
                """Nhận TradeMessage → cập nhật nến → kiểm tra tín hiệu MA."""
                if not isinstance(msg, TradeMessage):
                    return

                closed_bar = await builder.on_trade(msg.price, msg.quantity)
                if closed_bar is None:
                    return  # Nến chưa đóng, chờ tick tiếp theo

                # --- Nến vừa đóng ---
                bars = builder.snapshot()
                mf = calculate_ma(bars, MA_FAST)
                ms = calculate_ma(bars, MA_SLOW)
                print(
                    f"  [BAR] close={closed_bar.close:,.0f} vol={closed_bar.volume:,} | "
                    f"MA{MA_FAST}={mf:,.2f} | MA{MA_SLOW}={ms:,.2f}"
                    if mf and ms else
                    f"  [BAR] close={closed_bar.close:,.0f} | Chưa đủ dữ liệu MA"
                )

                signal = detect_cross(bars, MA_FAST, MA_SLOW)
                if signal is None:
                    return

                async with state_lock:
                    if state["active_order_id"] is not None:
                        print(f"  [SIGNAL {signal}] Đang có lệnh chờ, bỏ qua.")
                        return
                    if state["last_signal"] == signal:
                        return  # Không lặp tín hiệu cùng chiều
                    state["last_signal"] = signal

                print(f"\n  *** SIGNAL {signal} {SYMBOL} ***")
                side = OrderSide.BUY if signal == "BUY" else OrderSide.SELL

                # --- Kiểm tra risk ---
                max_bs = await trading.trading.get_max_buy_sell_at_market_price(ACCOUNT_NO, SYMBOL)
                max_qty = max_bs.max_buy_quantity if signal == "BUY" else max_bs.max_sell_quantity
                if max_qty < QUANTITY:
                    print(f"  [RISK] Không đủ {QUANTITY} (có {max_qty}). Bỏ qua.")
                    return

                # --- Đặt lệnh MARKET ---
                result = await trading.trading.place_market_order(
                    account_no=ACCOUNT_NO,
                    symbol=SYMBOL,
                    side=side,
                    quantity=QUANTITY,
                )
                order_id = getattr(result, "order_id", None) or "pending"
                print(f"  [ORDER] Đặt lệnh thành công: orderId={order_id}")
                async with state_lock:
                    state["active_order_id"] = order_id

            async def on_trading_event(msg):
                """Nhận OrderStatusMessage → in trạng thái, ghi P&L khi FILLED."""
                if not isinstance(msg, OrderStatusMessage):
                    return
                print(
                    f"  [ORDER UPDATE] {msg.symbol} {msg.side} | "
                    f"ID={msg.order_id} | Status={msg.status} | "
                    f"Khớp={msg.filled_quantity}/{msg.quantity}"
                )
                if msg.status not in TERMINAL_STATUSES:
                    return

                async with state_lock:
                    state["active_order_id"] = None

                filled_qty = getattr(msg, "filled_quantity", 0) or 0
                avg_price = getattr(msg, "avg_price", None)
                if msg.status == OrderStatus.FILLED and filled_qty > 0 and avg_price:
                    cost = filled_qty * avg_price
                    print(
                        f"  [FILLED] {state['last_signal']} {msg.symbol}: "
                        f"{filled_qty} CP @ {avg_price:,.0f} | "
                        f"Tổng: {cost:,.0f} VND"
                    )
                elif msg.status in (OrderStatus.CANCELLED, OrderStatus.REJECTED):
                    print(f"  [CLOSED] Lệnh kết thúc với trạng thái {msg.status}")

            def on_heartbeat(msg: HeartbeatMessage):
                print(f"  [HEARTBEAT] {msg.status}")

            # ===== Bước 3: Kết nối WebSocket =====
            async with AsyncStream(auth) as stream:
                stream.streaming.on_data = on_market_data
                stream.streaming.on_trading = on_trading_event
                stream.streaming.on_heartbeat = on_heartbeat

                print("\n--- Kết nối WebSocket ---")
                await stream.streaming.connect()
                print("Đã kết nối!\n")

                # Subscribe trade data để cập nhật nến
                await stream.streaming.subscribe_symbol([SYMBOL])

                # Subscribe order status để nhận kết quả lệnh real-time
                await stream.streaming.subscribe_order_status(ACCOUNT_NO)

                # ===== Bước 4: Ping keepalive =====
                # SDK tự gửi ping định kỳ và xử lý pong để giữ kết nối
                await stream.streaming.ping(interval=30)

                print(
                    f"Đang lắng nghe nến {BAR_SECONDS}s cho {SYMBOL}... "
                    f"(Ctrl+C để dừng)\n"
                )
                try:
                    await stream.streaming.wait()
                except KeyboardInterrupt:
                    print("\nDừng chiến lược.")


asyncio.run(main())
