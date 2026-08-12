"""
Sample 4 — Lấy danh sách cổ phiếu theo sàn
=============================================
Tạo watchlist/screener theo tiêu chí thị trường.

Luồng:
  1. Client gọi securitiesByBoard theo exchange, board, sector
  2. API trả về danh sách mã + thông tin giao dịch cơ bản
  3. Lọc/sắp xếp theo nhu cầu UI
  4. Lưu cursor để phân trang dữ liệu lớn
  5. Khi user chọn mã, chuyển sang luồng xem chi tiết/đặt lệnh
"""

from ssi_sdk import Auth, Data
from ssi_sdk.enums import Board
from auth_helper import ensure_auth
from config import config

with Auth(config) as auth:
    ensure_auth(auth)

    with Data(auth) as data:
        # --- Bước 1: Lấy danh sách cổ phiếu sàn HOSE ---
        print("--- Cổ phiếu sàn HOSE ---")
        hose_securities = data.market_data.get_securities_info_by_board(Board.HOSE)
        print(f"Tổng số mã: {len(hose_securities)}\n")

        for sec in hose_securities[:10]:
            print(
                f"  {sec.symbol:<10} | {sec.symbol_name_vi or sec.symbol_name_en or '':<30} "
                f"| Lot: {sec.lot_size}"
            )

        # --- Bước 2: Lấy danh sách cổ phiếu sàn HNX ---
        print("\n--- Cổ phiếu sàn HNX ---")
        hnx_securities = data.market_data.get_securities_info_by_board(Board.HNX)
        print(f"Tổng số mã: {len(hnx_securities)}")

        for sec in hnx_securities[:10]:
            print(
                f"  {sec.symbol:<10} | {sec.symbol_name_vi or sec.symbol_name_en or '':<30} "
                f"| Lot: {sec.lot_size}"
            )

        # --- Bước 3: Lấy theo chỉ số (index) ---
        print("\n--- Cổ phiếu thuộc VN30 ---")
        vn30_securities = data.market_data.get_securities_info_by_index("VN30")
        print(f"Tổng số mã: {len(vn30_securities)}")

        for sec in vn30_securities:
            print(f"  {sec.symbol:<10} | {sec.symbol_name_vi or sec.symbol_name_en or ''}")

        # --- Bước 4: Xem thông tin chi tiết một mã ---
        print("\n--- Chi tiết mã SSI ---")
        info = data.market_data.get_securities_info("SSI")
        print(f"  Mã          : {info.symbol}")
        print(f"  Tên (VI)    : {info.symbol_name_vi}")
        print(f"  Tên (EN)    : {info.symbol_name_en}")
        print(f"  Sàn         : {info.board}")
        print(f"  Lot size    : {info.lot_size}")
        print(f"  ICB Code    : {info.icb_code}")
        print(f"  ICB Name    : {info.icb_name}")
        print(f"  Listed Shares: {info.listed_shares}")
