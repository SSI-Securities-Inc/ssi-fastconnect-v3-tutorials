"""
Sample 15 (Async) — Vị thế (T1/T2 mua & bán), PPMMR (sức mua / margin) và Max mua/bán
================================================================================
Tổng hợp các API danh mục mới/đổi ở SDK 3.3.0. Chỉ ĐỌC dữ liệu, không đặt lệnh.

Luồng:
  1. Vị thế cổ phiếu: số lượng, bán được, KL chờ về T1/T2 (CẢ mua lẫn bán)
  2. PPMMR tài khoản cổ phiếu: tiền mua/bán T0/T1/T2, tỷ lệ margin, sức mua
  3. Tài khoản phái sinh (nếu có derivative_account trong config.json): vị thế + PPMMR
  4. Max mua / bán cho một mã (giá giới hạn và giá thị trường)

Phân biệt dễ nhầm:
  - EquityPosition.t1/t2_buy_quantity, t1/t2_sell_quantity là KHỐI LƯỢNG (cổ phiếu) chờ về.
  - EquityPPMMR.buy_t0/t1/t2, sell_t0/t1/t2 là SỐ TIỀN (VND) mua/bán chờ thanh toán.
  - REST chỉ trả danh sách vị thế theo mã; dòng tổng hợp (query_summary) chỉ có ở trading
    WebSocket thử nghiệm (xem sample 17).
  - MaxBuySellResponse.purchasing_power thay cho purchase_power (deprecated); với tài khoản
    phái sinh purchasing_power và margin_ratio là None (khác 0).
"""

import asyncio
from decimal import Decimal

from ssi_sdk import AsyncAuth, AsyncData, AsyncTrading
from auth_helper import ensure_auth_async
from helpers import mask_account, require_account
from config import config, ACCOUNT_NO, DERIVATIVE_ACCOUNT_NO, OTP

SYMBOL = "SSI"


def money(value) -> str:
    return "-" if value is None else f"{value:,.0f}"


async def main():
    require_account(ACCOUNT_NO)

    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth, otp=OTP)

        async with AsyncTrading(auth) as trading, AsyncData(auth) as data:
            # --- Bước 1: Vị thế cổ phiếu ---
            print(f"--- Vị thế cổ phiếu ({mask_account(ACCOUNT_NO)}) ---")
            positions = await trading.portfolio.get_equity_positions(ACCOUNT_NO)
            print(f"{'Mã':<8}{'SL':>9}{'Bán được':>10}{'T1 mua':>9}{'T2 mua':>9}{'T1 bán':>9}{'T2 bán':>9}{'Giá vốn':>12}")
            for pos in positions:
                print(
                    f"{pos.symbol:<8}{pos.quantity:>9}{pos.sellable_quantity:>10}"
                    f"{pos.t1_buy_quantity:>9}{pos.t2_buy_quantity:>9}"
                    f"{pos.t1_sell_quantity:>9}{pos.t2_sell_quantity:>9}{pos.cost_price:>12,.0f}"
                )
            if not positions:
                print("  (không có vị thế)")

            # --- Bước 2: PPMMR tài khoản cổ phiếu ---
            print("\n--- PPMMR cổ phiếu ---")
            eq = await trading.portfolio.get_equity_ppmmr(ACCOUNT_NO)
            print(f"  Sức mua (purchasing_power) : {money(eq.purchasing_power)}")
            print(f"  Tỷ lệ margin               : {eq.margin_ratio}")
            print(f"  Tổng tài sản / Tổng nợ     : {money(eq.total_assets)} / {money(eq.total_debt)}")
            print(f"  Tiền mua T0/T1/T2          : {money(eq.buy_t0)} / {money(eq.buy_t1)} / {money(eq.buy_t2)}")
            print(f"  Tiền bán T0/T1/T2          : {money(eq.sell_t0)} / {money(eq.sell_t1)} / {money(eq.sell_t2)}")

            # --- Bước 3: Tài khoản phái sinh (tùy chọn) ---
            if DERIVATIVE_ACCOUNT_NO and not DERIVATIVE_ACCOUNT_NO.startswith("<"):
                print(f"\n--- Phái sinh ({mask_account(DERIVATIVE_ACCOUNT_NO)}) ---")
                all_pos = await trading.portfolio.get_derivative_positions(DERIVATIVE_ACCOUNT_NO)
                for group in all_pos:
                    for p in group.open_positions:
                        print(f"  MỞ  {p.symbol:<12} long={p.long} short={p.short} net={p.net} floating P/L={money(p.floating_pl)}")
                    for p in group.closed_positions:
                        print(f"  ĐÓNG {p.symbol:<11} long={p.long} short={p.short} net={p.net} trading P/L={money(p.trading_pl)}")
                dv = await trading.portfolio.get_derivative_ppmmr(DERIVATIVE_ACCOUNT_NO)
                print(f"  Số dư: {money(dv.account_balance)} | Tổng P/L: {money(dv.total_pl)} | Rút được: {money(dv.withdrawable)}")
                print(f"  Ký quỹ yêu cầu SSI: {money(dv.margin_req_ssi)} | Tỷ lệ TK SSI: {dv.account_ratio_ssi}")
            else:
                print("\n(Bỏ qua phái sinh: chưa có derivative_account trong config.json)")

            # --- Bước 4: Max mua / bán ---
            print(f"\n--- Max mua/bán {SYMBOL} ---")
            master_data = await data.market_data.get_master_data()
            master = next((m for m in master_data if m.symbol == SYMBOL), None)
            if master is not None and master.ref_price > 0:
                price = Decimal(str(master.ref_price))
                mx = await trading.trading.get_max_buy_sell(ACCOUNT_NO, SYMBOL, price)
                print(f"  @ giá tham chiếu {price}: mua {mx.max_buy_quantity} | bán {mx.max_sell_quantity}"
                      f" | sức mua {mx.purchasing_power} | margin {mx.margin_ratio}")
            mx_mkt = await trading.trading.get_max_buy_sell_at_market_price(ACCOUNT_NO, SYMBOL)
            print(f"  @ giá thị trường: mua {mx_mkt.max_buy_quantity} | bán {mx_mkt.max_sell_quantity}")


if __name__ == "__main__":
    asyncio.run(main())
