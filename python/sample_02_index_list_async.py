"""
Sample 2 (Async) — Lấy danh sách chỉ số thị trường (Index)
=============================================================
Hiển thị VN-Index, HNX-Index… trên dashboard (phiên bản async).

Luồng:
  1. Client gọi endpoint indexList với filter (exchange, limit, view)
  2. API trả về danh sách chỉ số và dữ liệu giá hiện tại
  3. Map dữ liệu sang model hiển thị UI
  4. Lưu cursor để tải trang kế tiếp khi cần
  5. Cache ngắn hạn để giảm số lần gọi API
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncData
from ssi_sdk.enums import Board
from auth_helper import ensure_auth_async
from config import config

async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth)

        async with AsyncData(auth) as data:
            # --- Bước 1: Lấy toàn bộ chỉ số ---
            all_indexes = await data.market_data.get_indexes()
            print(f"Tổng số chỉ số: {len(all_indexes)}\n")

            for idx in all_indexes:
                print(f"  {idx.index:<15} | {idx.index_name:<30} | Sàn: {idx.board}")

            # --- Bước 2: Lọc chỉ số theo sàn HOSE ---
            print("\n--- Chỉ số sàn HOSE ---")
            hose_indexes = await data.market_data.get_indexes_by_board(Board.HOSE)
            for idx in hose_indexes:
                print(f"  {idx.index:<15} | {idx.index_name}")

            # --- Bước 3: Lấy chi tiết summary cho một chỉ số cụ thể ---
            print("\n--- VN-Index Summary ---")
            summary = await data.market_data.get_index_summary("VNINDEX")
            if summary:
                print(f"  Giá trị Index    : {summary.index_value}")
                print(f"  Thay đổi         : {summary.index_change} ({summary.index_change_percent}%)")
                print(f"  Tổng KL khớp     : {summary.total_match}")
                print(f"  Tổng GT khớp     : {summary.total_match_value}")
                print(f"  Tăng / Giảm / Đứng: {summary.total_advance_stock} / {summary.total_decline_stock} / {summary.total_steady_stock}")
                print(f"  Trần / Sàn       : {summary.total_ceiling_stock} / {summary.total_floor_stock}")
            else:
                print("Không lấy được summary cho VN-Index.")


asyncio.run(main())
