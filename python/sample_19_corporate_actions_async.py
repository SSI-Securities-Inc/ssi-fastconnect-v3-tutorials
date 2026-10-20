"""
Sample 19 (Async) — Sự kiện doanh nghiệp (Corporate actions)
=============================================================
Xem cổ tức, cổ phiếu thưởng, quyền mua... của các mã. Chỉ ĐỌC dữ liệu, không đặt lệnh.

Luồng:
  1. get_corporate_actions(): sự kiện của HÔM NAY
  2. get_corporate_actions_historical(from, to): sự kiện trong một khoảng ngày ở quá khứ
     (server cho tối đa 5 ngày mỗi lần gọi)
  3. Lọc theo mã (symbol) và ngôn ngữ ("vn" / "en")
  4. Khoảng ngày quá rộng bị SDK chặn tại client (ValidationError), không gọi server

Bước 1-3 độc lập nên chạy song song bằng asyncio.gather.

Lưu ý: SDK tự lấy hết các trang. Mọi trường là chuỗi đúng như server gửi (ngày YYYY/MM/DD,
ratio không được parse, trường để trống là ""). Ngày truyền vào phải đúng dạng YYYY/MM/DD.
"""

import asyncio
from datetime import datetime, timedelta

from ssi_sdk import AsyncAuth, AsyncData, ValidationError
from auth_helper import ensure_auth_async
from config import config

DATE_FMT = "%Y/%m/%d"


def show(actions, limit=10):
    for a in actions[:limit]:
        print(
            f"  {a.symbol:<8} | {a.exchange:<6} | GDKHQ: {a.exright_date or '-':<10} "
            f"| Chốt: {a.record_date or '-':<10} | Phát hành: {a.issue_date or '-':<10} "
            f"| Tỷ lệ: {a.ratio or '-':<8} | {a.event_title}"
        )
    if len(actions) > limit:
        print(f"  ... còn {len(actions) - limit} sự kiện")


async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth)

        async with AsyncData(auth) as data:
            today = datetime.now()
            from_date = (today - timedelta(days=4)).strftime(DATE_FMT)
            to_date = today.strftime(DATE_FMT)

            # --- Bước 1-3: Gọi song song ---
            today_actions, range_actions, ssi_actions = await asyncio.gather(
                data.market_data.get_corporate_actions(),
                data.market_data.get_corporate_actions_historical(from_date, to_date),
                data.market_data.get_corporate_actions_historical(
                    from_date, to_date, symbol="SSI", language="en"
                ),
            )

            print("--- Sự kiện doanh nghiệp hôm nay ---")
            print(f"Tổng số sự kiện: {len(today_actions)}")
            show(today_actions)

            print(f"\n--- Sự kiện từ {from_date} đến {to_date} ---")
            print(f"Tổng số sự kiện: {len(range_actions)}")
            show(range_actions)

            print("\n--- Sự kiện của mã SSI (tiếng Anh) ---")
            print(f"Tổng số sự kiện: {len(ssi_actions)}")
            show(ssi_actions)

            # --- Bước 4: Khoảng ngày quá rộng bị chặn tại SDK ---
            print("\n--- Khoảng 30 ngày (vượt giới hạn 5 ngày) ---")
            too_wide_from = (today - timedelta(days=30)).strftime(DATE_FMT)
            try:
                await data.market_data.get_corporate_actions_historical(too_wide_from, to_date)
            except ValidationError as e:
                print(f"  SDK từ chối trước khi gọi server: {e}")
                print("  -> chia khoảng ngày thành các đoạn <= 5 ngày rồi gọi từng đoạn.")

            # --- Response Summary ---
            print("\n[Response] today_count|range_count|ssi_range_count")
            print(f"{len(today_actions)}|{len(range_actions)}|{len(ssi_actions)}")


asyncio.run(main())
