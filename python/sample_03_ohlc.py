"""
Sample 3 — Lấy dữ liệu K-line (OHLC)
=======================================
Cung cấp dữ liệu nến cho biểu đồ và phân tích kỹ thuật.

Luồng:
  1. Client gửi symbol, interval, startTime, endTime
  2. API trả về mảng OHLCV theo mốc thời gian
  3. Chuẩn hóa dữ liệu (time/open/high/low/close/volume)
  4. Truyền vào chart component hoặc indicator engine
  5. Nếu lịch sử dài thì lặp theo paging/window thời gian
"""

from ssi_sdk import Auth, Data
from auth_helper import ensure_auth
from config import config

SYMBOL = "SSI"

with Auth(config) as auth:
    ensure_auth(auth)

    with Data(auth) as data:
        # --- Bước 1: Lấy OHLC ngày gần nhất ---
        print(f"--- OHLC 1 ngày gần nhất ({SYMBOL}) ---")
        daily = data.market_data.get_ohlc_1day_historical(SYMBOL, from_date="2026/03/01 00:00:00", to_date="2026/03/27 23:59:59")
        print(daily)
        for bar in daily[:5]:
            print(
                f"  {bar.trading_date} | "
                f"O:{bar.open_price:>10} H:{bar.high_price:>10} "
                f"L:{bar.low_price:>10} C:{bar.close_price:>10} "
                f"V:{bar.volume:>12}"
            )

        # --- Bước 2: Lấy OHLC lịch sử theo khoảng thời gian ---
        print(f"\n--- OHLC 1 ngày lịch sử ({SYMBOL}) ---")
        hist = data.market_data.get_ohlc_1day_historical(
            symbol=SYMBOL,
            from_date="2026/01/01 00:00:00",
            to_date="2026/03/27 23:59:59",
            page=1,
            size=20,
        )
        for bar in hist:
            print(
                f"  {bar.trading_date} | "
                f"O:{bar.open_price:>10} H:{bar.high_price:>10} "
                f"L:{bar.low_price:>10} C:{bar.close_price:>10} "
                f"V:{bar.volume:>12}"
            )

        # --- Bước 3: Lấy OHLC theo timeframe khác (1h, 15m...) ---
        print(f"\n--- OHLC 1 giờ gần nhất ({SYMBOL}) ---")
        hourly = data.market_data.get_ohlc_1hour(SYMBOL)
        for bar in hourly[:5]:
            print(
                f"  {bar.trading_date} | "
                f"O:{bar.open_price:>10} H:{bar.high_price:>10} "
                f"L:{bar.low_price:>10} C:{bar.close_price:>10} "
                f"V:{bar.volume:>12}"
            )

        # --- Bước 4: Phân trang cho dữ liệu lớn ---
        print(f"\n--- Paging OHLC 1 phút lịch sử ({SYMBOL}) ---")
        page = 1
        total_bars = 0
        while True:
            bars = data.market_data.get_ohlc_1minute_historical(
                symbol=SYMBOL,
                from_date="2026/03/25 09:00:00",
                to_date="2026/03/25 15:00:00",
                page=page,
                size=100,
            )
            if not bars:
                break
            total_bars += len(bars)
            print(f"  Trang {page}: {len(bars)} nến (tổng: {total_bars})")
            page += 1

        print(f"\nTổng cộng {total_bars} nến 1 phút được tải.")
