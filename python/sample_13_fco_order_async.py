"""
Sample 13 Async — Đặt Lệnh Điều Kiện (FCO)
===========================================
Thể hiện đầy đủ các loại lệnh điều kiện (Fast Conditional Orders - FCO) phiên bản bất đồng bộ:
  1. GTD (Good-Till-Date / Lệnh chờ theo ngày)
  2. Stop (Lệnh dừng giá thị trường)
  3. Stop Limit (Lệnh dừng giá giới hạn)
  4. Trailing Stop (Lệnh dừng xu hướng)
  5. Trailing Stop Limit (Lệnh dừng xu hướng giới hạn)
  6. OCO (One-Cancels-the-Other / Lệnh Chốt lời & Cắt lỗ)
  7. Bull Bear (Lệnh Hai đầu)
  8. Truy vấn danh sách & Hủy lệnh FCO
"""

import asyncio
from ssi_sdk import AsyncAuth, AsyncTrading
from auth_helper import ensure_auth_async
from ssi_sdk.enums import FCOOperator, OrderSide
from config import config, ACCOUNT_NO, OTP

symbol = "SSI"
from_date = "2026/08/01 00:00:00"
to_date = "2026/08/30 23:59:59"


async def main():
    print("=== FASTCONNECT PYTHON SDK (ASYNC) — SAMPLE 13: LỆNH ĐIỀU KIỆN (FCO) ===\n")

    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth, otp=OTP)

        async with AsyncTrading(auth) as trading:
            # --- 1. Lệnh GTD (Good-Till-Date) ---
            print("--- 1. Đặt lệnh GTD ---")
            gtd_res = await trading.trading.place_fco_gtd(
                account_no=ACCOUNT_NO,
                symbol=symbol,
                side=OrderSide.BUY,
                quantity=100,
                price=26000,
                price_slip=0,
                from_date=from_date,
                to_date=to_date,
            )
            print(f"  GTD Result: {gtd_res}")

            # --- 2. Lệnh Stop (Stop Market) ---
            print("\n--- 2. Đặt lệnh Stop ---")
            stop_res = await trading.trading.place_fco_stop(
                account_no=ACCOUNT_NO,
                symbol=symbol,
                side=OrderSide.BUY,
                quantity=100,
                stop_price=27000,
                operator=FCOOperator.GREATER_OR_EQUAL,
                from_date=from_date,
                to_date=to_date,
            )
            print(f"  Stop Result: {stop_res}")

            # --- 3. Lệnh Stop Limit ---
            print("\n--- 3. Đặt lệnh Stop Limit ---")
            stop_limit_res = await trading.trading.place_fco_stop_limit(
                account_no=ACCOUNT_NO,
                symbol=symbol,
                side=OrderSide.BUY,
                quantity=100,
                price=27500,
                price_slip=0,
                stop_price=27000,
                operator=FCOOperator.GREATER_OR_EQUAL,
                from_date=from_date,
                to_date=to_date,
            )
            print(f"  Stop Limit Result: {stop_limit_res}")

            # --- 4. Lệnh Trailing Stop ---
            print("\n--- 4. Đặt lệnh Trailing Stop ---")
            trailing_res = await trading.trading.place_fco_trailing_stop(
                account_no=ACCOUNT_NO,
                symbol=symbol,
                side=OrderSide.SELL,
                quantity=100,
                active_price=28000,
                trailing_amount=1000,
                from_date=from_date,
                to_date=to_date,
            )
            print(f"  Trailing Stop Result: {trailing_res}")

            # --- 5. Lệnh Trailing Stop Limit ---
            print("\n--- 5. Đặt lệnh Trailing Stop Limit ---")
            trailing_limit_res = await trading.trading.place_fco_trailing_stop_limit(
                account_no=ACCOUNT_NO,
                symbol=symbol,
                side=OrderSide.SELL,
                quantity=100,
                active_price=28000,
                trailing_amount=1000,
                price_slip=500,
                from_date=from_date,
                to_date=to_date,
            )
            print(f"  Trailing Stop Limit Result: {trailing_limit_res}")

            # --- 6. Lệnh OCO (One-Cancels-the-Other) ---
            print("\n--- 6. Đặt lệnh OCO ---")
            oco_res = await trading.trading.place_fco_oco(
                account_no=ACCOUNT_NO,
                symbol=symbol,
                side=OrderSide.SELL,
                quantity=100,
                tp_active_price=30000,
                sl_active_price=24000,
                tp_price=30000,
                sl_price=24000,
                tp_slip=0,
                sl_slip=0,
                from_date=from_date,
                to_date=to_date,
            )
            print(f"  OCO Result: {oco_res}")

            # --- 7. Lệnh Bull Bear ---
            print("\n--- 7. Đặt lệnh Bull Bear ---")
            bb_res = await trading.trading.place_fco_bull_bear(
                account_no=ACCOUNT_NO,
                symbol=symbol,
                side=OrderSide.BUY,
                quantity=100,
                price=26000,
                price_slip=0,
                tp_active_price=30000,
                sl_active_price=24000,
                tp_price=30000,
                sl_price=24000,
                tp_slip=0,
                sl_slip=0,
                from_date=from_date,
                to_date=to_date,
            )
            print(f"  Bull Bear Result: {bb_res}")

            # --- 8. Truy vấn danh sách lệnh FCO ---
            print("\n--- 8. Danh sách lệnh FCO ---")
            fco_list = await trading.trading.get_fco_by_account_no(ACCOUNT_NO, page_index=1, page_size=10)
            print(f"  Tổng số lệnh FCO: {fco_list.items_count}")
            for item in fco_list.fco_list[:5]:
                print(f"  FCO ID: {item.fco_id} | Mã: {item.symbol} | Loại: {item.type} | Trạng thái: {item.status}")

            # --- 9. Hủy lệnh FCO vừa tạo nếu có ---
            if hasattr(gtd_res, "fco_id") and gtd_res.fco_id:
                print(f"\n--- 9. Hủy lệnh FCO ID: {gtd_res.fco_id} ---")
                cancel_res = await trading.trading.cancel_fco(gtd_res.fco_id)
                print(f"  Hủy FCO Result: {cancel_res}")

            print("\n[Response] sample_13_fco_completed")


if __name__ == "__main__":
    asyncio.run(main())
