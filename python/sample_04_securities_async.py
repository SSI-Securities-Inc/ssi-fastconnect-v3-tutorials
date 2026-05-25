"""
Sample 4 (Async) — Lấy danh sách cổ phiếu theo sàn
=====================================================
Tạo watchlist/screener theo tiêu chí thị trường (phiên bản async).

Luồng:
  1. Client gọi securitiesByBoard theo exchange, board, sector
  2. API trả về danh sách mã + thông tin giao dịch cơ bản
  3. Lọc/sắp xếp theo nhu cầu UI
  4. Lưu cursor để phân trang dữ liệu lớn
  5. Khi user chọn mã, chuyển sang luồng xem chi tiết/đặt lệnh
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncData, Config
from ssi_sdk.enums import Board
from auth_helper import ensure_auth_async

config = Config(
    client_id="<CLIENT_ID>",
    api_key="<API_KEY>",
    api_secret="<API_SECRET>",
    private_key=("<PRIVATE_KEY_CONTENT>"),
    log_level="DEBUG",
)

async def main():
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth)

        async with AsyncData(auth) as data:
            # --- Bước 1-2: Lấy song song danh sách cổ phiếu HOSE + HNX ---
            hose_task = data.market_data.get_securities_info_by_board(Board.HOSE)
            hnx_task = data.market_data.get_securities_info_by_board(Board.HNX)
            hose_securities, hnx_securities = await asyncio.gather(hose_task, hnx_task)

            print("--- Cổ phiếu sàn HOSE ---")
            print(f"Tổng số mã: {len(hose_securities)}\n")
            for sec in hose_securities[:10]:
                print(
                    f"  {sec.symbol:<10} | {sec.symbol_name_vi or sec.symbol_name_en or '':<30} "
                    f"| Lot: {sec.lot_size}"
                )

            print("\n--- Cổ phiếu sàn HNX ---")
            print(f"Tổng số mã: {len(hnx_securities)}")
            for sec in hnx_securities[:10]:
                print(
                    f"  {sec.symbol:<10} | {sec.symbol_name_vi or sec.symbol_name_en or '':<30} "
                    f"| Lot: {sec.lot_size}"
                )

            # --- Bước 3: Lấy theo chỉ số (index) ---
            print("\n--- Cổ phiếu thuộc VN30 ---")
            vn30_securities = await data.market_data.get_securities_info_by_index("VN30")
            print(f"Tổng số mã: {len(vn30_securities)}")
            for sec in vn30_securities:
                print(f"  {sec.symbol:<10} | {sec.symbol_name_vi or sec.symbol_name_en or ''}")

            # --- Bước 4: Xem thông tin chi tiết một mã ---
            print("\n--- Chi tiết mã SSI ---")
            info = await data.market_data.get_securities_info("SSI")
            print(f"  Mã          : {info.symbol}")
            print(f"  Tên (VI)    : {info.symbol_name_vi}")
            print(f"  Tên (EN)    : {info.symbol_name_en}")
            print(f"  Sàn         : {info.board}")
            print(f"  Lot size    : {info.lot_size}")
            print(f"  ICB Code    : {info.icb_code}")
            print(f"  ICB Name    : {info.icb_name}")
            print(f"  Listed Shares: {info.listed_shares}")


asyncio.run(main())
