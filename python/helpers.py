"""
helpers — Tiện ích nhỏ dùng chung cho các sample
==================================================
Từ SDK 3.3.0, các trường enum (Account.account_type, Order.side/order_type/status...)
có thể là enum, str (giá trị server mới SDK chưa biết) hoặc None. Hàm enum_value()
lấy chuỗi hiển thị an toàn cho cả 3 trường hợp.
"""


def enum_value(value) -> str:
    """Trả về .value nếu là enum, giữ nguyên nếu là str, '-' nếu None."""
    if value is None:
        return "-"
    return str(getattr(value, "value", value))


def mask_account(account_no: str) -> str:
    """Che số tài khoản khi in ra màn hình (chỉ giữ 3 ký tự cuối)."""
    if not account_no:
        return "-"
    return "*" * max(len(account_no) - 3, 0) + account_no[-3:]


def require_account(account_no: str) -> str:
    """Dừng ngay (fail fast) nếu config.json chưa có equity_account thật."""
    if not account_no or account_no.startswith("<"):
        raise SystemExit("Thiếu equity_account trong config.json — xem config.example.json.")
    return account_no


def confirm_real_orders(what: str) -> None:
    """Cảnh báo lệnh THẬT và yêu cầu gõ YES mới tiếp tục."""
    print("=" * 70)
    print(f"CẢNH BÁO: sample này sẽ gửi {what} THẬT lên tài khoản trong config.json.")
    print("Chỉ chạy khi bạn hiểu rõ và chấp nhận rủi ro. Lệnh dùng giá sàn (xa thị trường).")
    print("=" * 70)
    if input("Gõ YES để tiếp tục: ").strip() != "YES":
        raise SystemExit("Đã hủy, không có lệnh nào được gửi.")
