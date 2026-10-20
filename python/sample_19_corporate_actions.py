"""
Sample 19 — Sự kiện doanh nghiệp (Corporate actions)
=====================================================
Xem cổ tức, cổ phiếu thưởng, quyền mua... của các mã. Chỉ ĐỌC dữ liệu, không đặt lệnh.

Luồng:
  1. get_corporate_actions(): sự kiện của HÔM NAY
  2. get_corporate_actions_historical(from, to): sự kiện trong một khoảng ngày ở quá khứ
     (server cho tối đa 5 ngày mỗi lần gọi)
  3. Lọc theo mã (symbol) và ngôn ngữ ("vn" / "en")
  4. Khoảng ngày quá rộng bị SDK chặn tại client (ValidationError), không gọi server

Lưu ý:
  - SDK tự lấy hết các trang (pageSize 100), bạn không cần tự phân trang.
  - Mọi trường là chuỗi đúng như server gửi. Ngày có dạng YYYY/MM/DD, ratio KHÔNG được parse
    vì ý nghĩa phụ thuộc loại sự kiện. Trường server để trống (sự kiện mới công bố chưa có
    ngày GDKHQ/tỷ lệ) là "".
  - Ngày truyền vào phải đúng dạng YYYY/MM/DD (không dùng dấu '-').
  - "Hôm nay" lấy theo ngày của máy bạn.
"""

from datetime import datetime, timedelta

from ssi_sdk import Auth, Data, ValidationError
from auth_helper import ensure_auth
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


with Auth(config) as auth:
    ensure_auth(auth)

    with Data(auth) as data:
        # --- Bước 1: Sự kiện của hôm nay ---
        print("--- Sự kiện doanh nghiệp hôm nay ---")
        today_actions = data.market_data.get_corporate_actions()
        print(f"Tổng số sự kiện: {len(today_actions)}")
        show(today_actions)

        # --- Bước 2: Sự kiện trong 5 ngày gần nhất (tối đa 5 ngày cho mỗi lần gọi) ---
        today = datetime.now()
        from_date = (today - timedelta(days=4)).strftime(DATE_FMT)
        to_date = today.strftime(DATE_FMT)
        print(f"\n--- Sự kiện từ {from_date} đến {to_date} ---")
        range_actions = data.market_data.get_corporate_actions_historical(from_date, to_date)
        print(f"Tổng số sự kiện: {len(range_actions)}")
        show(range_actions)

        # --- Bước 3: Lọc theo mã, trả tiêu đề bằng tiếng Anh ---
        print("\n--- Sự kiện của mã SSI (tiếng Anh) ---")
        ssi_actions = data.market_data.get_corporate_actions_historical(
            from_date, to_date, symbol="SSI", language="en"
        )
        print(f"Tổng số sự kiện: {len(ssi_actions)}")
        show(ssi_actions)

        # --- Bước 4: Khoảng ngày quá rộng bị chặn tại SDK ---
        print("\n--- Khoảng 30 ngày (vượt giới hạn 5 ngày) ---")
        too_wide_from = (today - timedelta(days=30)).strftime(DATE_FMT)
        try:
            data.market_data.get_corporate_actions_historical(too_wide_from, to_date)
        except ValidationError as e:
            print(f"  SDK từ chối trước khi gọi server: {e}")
            print("  -> chia khoảng ngày thành các đoạn <= 5 ngày rồi gọi từng đoạn.")

        # --- Response Summary ---
        print("\n[Response] today_count|range_count|ssi_range_count")
        print(f"{len(today_actions)}|{len(range_actions)}|{len(ssi_actions)}")
