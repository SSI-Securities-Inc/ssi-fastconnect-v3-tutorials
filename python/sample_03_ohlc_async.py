"""
Sample 3 (Async) — Lấy dữ liệu K-line (OHLC)
================================================
Cung cấp dữ liệu nến cho biểu đồ và phân tích kỹ thuật (phiên bản async).

Luồng:
  1. Client gửi symbol, interval, startTime, endTime
  2. API trả về mảng OHLCV theo mốc thời gian
  3. Chuẩn hóa dữ liệu (time/open/high/low/close/volume)
  4. Truyền vào chart component hoặc indicator engine
  5. Nếu lịch sử dài thì lặp theo paging/window thời gian
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncData
from ssi_sdk.enums import Timeframe
from ssi_sdk.exceptions import ValidationError
from auth_helper import ensure_auth_async
from config import config

SYMBOL = "SSI"


async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth)

        async with AsyncData(auth) as data:
            # --- Bước 1: Lấy OHLC ngày gần nhất ---
            print(f"--- OHLC 1 ngày gần nhất ({SYMBOL}) ---")
            daily = await data.market_data.get_ohlc_1day_historical(
                SYMBOL, from_date="2026/03/01 00:00:00", to_date="2026/03/27 23:59:59"
            )
            for bar in daily[:5]:
                print(
                    f"  {bar.trading_date} | "
                    f"O:{bar.open_price:>10} H:{bar.high_price:>10} "
                    f"L:{bar.low_price:>10} C:{bar.close_price:>10} "
                    f"V:{bar.volume:>12}"
                )

            # --- Bước 2: Lấy OHLC lịch sử theo khoảng thời gian ---
            print(f"\n--- OHLC 1 ngày lịch sử ({SYMBOL}) ---")
            hist = await data.market_data.get_ohlc_1day_historical(
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
            hourly = await data.market_data.get_ohlc_1hour(SYMBOL)
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
                bars = await data.market_data.get_ohlc_1minute_historical(
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

            # --- Bước 5: Lấy TOÀN BỘ nến trong khoảng (SDK 3.3.0: get_ohlc_all / iter_ohlc) ---
            # Không cần tự viết vòng lặp paging như Bước 4. Timeframe hỗ trợ: 1d, 1m, 3m, 5m, 15m, 1h.
            print(f"\n--- get_ohlc_all / iter_ohlc ({SYMBOL}) ---")
            all_bars = await data.market_data.get_ohlc_all(
                SYMBOL, Timeframe.DAY_1,
                from_date="2026/01/01 00:00:00", to_date="2026/03/27 23:59:59",
                ascending=True,  # server trả newest first; ascending=True đảo lại cũ -> mới
            )
            print(f"  get_ohlc_all 1 ngày: {len(all_bars)} nến")
            # iter_ohlc (async generator): duyệt từng nến, tự phân trang, newest first
            count = 0
            async for bar in data.market_data.iter_ohlc(
                SYMBOL, Timeframe.MINUTE_5,
                from_date="2026/03/25 09:00:00", to_date="2026/03/25 15:00:00", size=100,
            ):
                count += 1
            print(f"  iter_ohlc 5 phút: {count} nến")

            # Lưu ý: nến tuần (1w) và tháng (1M) KHÔNG được server hỗ trợ (lỗi 400210).
            # Từ SDK 3.3.0 get_ohlc_1week_historical / get_ohlc_1month_historical raise ValidationError
            # ngay phía client, không gửi request. Hãy tự gộp nến ngày thành tuần/tháng nếu cần.
            try:
                await data.market_data.get_ohlc_1week_historical(
                    SYMBOL, from_date="2026/01/01 00:00:00", to_date="2026/03/27 23:59:59"
                )
            except ValidationError as e:
                print(f"  1 tuần không hỗ trợ: {e}")


asyncio.run(main())
