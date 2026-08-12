"""
Sample 1 — Xác thực, Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
==============================================================================
Đăng nhập, yêu cầu/xác thực mã OTP (SMS/Email hoặc Smart OTP Push Notification)
và lấy token cho toàn bộ API call sau đó.

Luồng:
  1. Tạo Config → Auth
  2. Kiểm tra token cache / refresh token nếu có.
  3. Nếu chưa có token: Gọi auth.request_otp() để yêu cầu gửi mã OTP hoặc Smart OTP Push
     - Nếu nhận transactionId: Polling chờ người dùng phê duyệt trên app SSI iBoard
     - Nếu là OTP 6 số: Nhập mã OTP và gọi auth.authenticate(otp=...)
  4. Auth service trả về accessToken, refreshToken, expiresIn và lưu vào token_cache.json
  5. Mọi request sau đó gắn Authorization: Bearer <accessToken>
  6. Xác nhận token bằng cách gọi API lấy thông tin tài khoản.
"""

from ssi_sdk import Auth, Trading
from auth_helper import ensure_auth
from config import config

with Auth(config) as auth:
    # --- Bước 1-3: Xác thực + Yêu cầu & Xác thực OTP / Smart OTP ---
    ensure_auth(auth)
    token = auth.token
    print("\n--- Thông tin Token ---")
    print("Access Token :", token.access_token[:40], "...")
    print("Token Type   :", token.token_type)
    print("Expires At   :", token.expires_at)
    print("Refresh Token:", token.refresh_token[:40] if token.refresh_token else "N/A", "...")

    # --- Bước 4: Kiểm tra token hết hạn & refresh ---
    if auth.is_token_expired:
        print("\nToken hết hạn, đang refresh...")
        new_token = auth.refresh()
        print("Token mới    :", new_token.access_token[:40], "...")

    # --- Bước 5: Xác nhận token hoạt động bằng cách gọi API ---
    with Trading(auth) as trading:
        accounts = trading.account.get_account_info()
        print(f"\nXác thực thành công! Tìm thấy {len(accounts)} tài khoản:")
        for acc in accounts:
            print(f"  - {acc.account_no} ({acc.account_type.value})")

