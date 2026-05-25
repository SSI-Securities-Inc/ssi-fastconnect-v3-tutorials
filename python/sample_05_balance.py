"""
Sample 5 — Lấy số dư tài khoản (Account Balance)
===================================================
Kiểm tra khả năng giao dịch trước khi đặt lệnh.

Luồng:
  1. Client gọi endpoint balances theo accountId
  2. API trả về available, onHold, limits, settlement
  3. Tính khả năng mua/bán thực tế theo nghiệp vụ
  4. Nếu không đủ điều kiện thì chặn thao tác đặt lệnh
  5. Nếu đủ điều kiện thì cho phép đi tiếp sang order flow
"""

from ssi_sdk import Auth, Trading, Config
from auth_helper import ensure_auth

config = Config(
    client_id="<CLIENT_ID>",
    api_key="<API_KEY>",
    api_secret="<API_SECRET>",
    private_key=("<PRIVATE_KEY_CONTENT>"),
    log_level="DEBUG",
)

ACCOUNT_NO = "<ACCOUNT_NO>"

with Auth(config) as auth:
    ensure_auth(auth, otp="203081")

    with Trading(auth) as trading:
        # --- Bước 1: Lấy danh sách tài khoản ---
        accounts = trading.account.get_account_info()
        print("Danh sách tài khoản:")
        for acc in accounts:
            print(f"  - {acc.account_no} ({acc.account_type.value})")

        # --- Bước 2: Lấy số dư tài khoản Equity ---
        print(f"\n--- Số dư tài khoản Equity: {ACCOUNT_NO} ---")
        balance = trading.portfolio.get_equity_balance(ACCOUNT_NO)
        print(f"  Tiền mặt khả dụng  : {balance.available_cash:>15,.0f}")
        print(f"  Tổng nợ            : {balance.total_debt:>15,.0f}")
        print(f"  Mua T0/T1/T2       : {balance.buy_t0:>12,.0f} / {balance.buy_t1:>12,.0f} / {balance.buy_t2:>12,.0f}")
        print(f"  Bán T0/T1/T2       : {balance.sell_t0:>12,.0f} / {balance.sell_t1:>12,.0f} / {balance.sell_t2:>12,.0f}")

        # --- Bước 3: Kiểm tra sức mua tối đa cho một mã ---
        print("\n--- Sức mua/bán tối đa: SSI ---")
        max_bs = trading.trading.get_max_buy_sell(ACCOUNT_NO, "SSI", 26000)
        print(f"  Max mua : {max_bs.max_buy_quantity:>10} cổ phiếu")
        print(f"  Max bán : {max_bs.max_sell_quantity:>10} cổ phiếu")
        print(f"  Sức mua : {max_bs.purchase_power:>15,.0f}")

        # --- Bước 4: Logic kiểm tra trước khi đặt lệnh ---
        desired_quantity = 100
        desired_price = 26000
        required_amount = desired_quantity * desired_price

        if balance.available_cash >= required_amount:
            print(f"\n✓ Đủ điều kiện: cần {required_amount:,.0f}, có {balance.available_cash:,.0f}")
            print("  → Cho phép đặt lệnh mua.")
        else:
            print(f"\n✗ Không đủ: cần {required_amount:,.0f}, chỉ có {balance.available_cash:,.0f}")
            print("  → Chặn đặt lệnh.")

        # --- Bước 5: Xem vị thế hiện có ---
        print(f"\n--- Vị thế cổ phiếu ({ACCOUNT_NO}) ---")
        positions = trading.portfolio.get_equity_positions(ACCOUNT_NO)
        for pos in positions:
            print(
                f"  {pos.symbol:<10} | SL: {pos.quantity:>8} | "
                f"Bán được: {pos.sellable_quantity:>8} | Giá vốn: {pos.cost_price:>10,.0f}"
            )
