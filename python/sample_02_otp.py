"""
Sample 2 — Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
========================================================================
Hướng dẫn các luồng xác thực mã OTP (SMS/Email hoặc Smart OTP Push Notification).

Luồng:
  1. Gọi auth.request_otp() để yêu cầu gửi mã OTP hoặc bắn thông báo Smart OTP tới App
  2. Nếu tài khoản dùng Smart OTP Push Notification:
     - Rút trích transactionId từ kết quả trả về
     - Gọi auth.ensure_authenticated(transaction_id=..., poll_interval=5, poll_max_retries=6)
       SDK sẽ tự động Polling chờ người dùng bấm Approve trên ứng dụng di động SSI iBoard
  3. Nếu tài khoản dùng OTP 6 số (SMS/Email/Mã hiển thị trên App Smart OTP):
     - Nhập mã OTP trực tiếp và gọi auth.authenticate(otp=...) hoặc auth.ensure_authenticated(otp=...)
"""

from ssi_sdk import Auth, Config
import config as app_config

config = Config(
    client_id=app_config.CLIENT_ID,
    api_key=app_config.API_KEY,
    api_secret=app_config.API_SECRET,
    private_key=app_config.PRIVATE_KEY,
    log_level="INFO",
)

with Auth(config) as auth:
    print("=== Bước 1: Yêu cầu OTP (Request OTP) ===")
    otp_res = auth.request_otp()
    print("Request OTP Response:", otp_res)

    # Rút trích transactionId nếu có (cho tài khoản Smart OTP Push Notification)
    transaction_id = None
    if isinstance(otp_res, dict):
        data_map = otp_res.get("data", {})
        if isinstance(data_map, dict):
            transaction_id = data_map.get("transactionId") or data_map.get("transaction_id")
        if not transaction_id:
            transaction_id = otp_res.get("transactionId") or otp_res.get("transaction_id")

    if transaction_id:
        print(f"\n[Smart OTP] Đã nhận Transaction ID: {transaction_id}")
        print("Vui lòng mở ứng dụng SSI trên điện thoại và bấm APPROVE (Xác nhận)...")
        print("SDK đang Polling chờ bạn bấm phê duyệt...")
        
        try:
            access_token = auth.ensure_authenticated(
                transaction_id=transaction_id,
                poll_interval=5,
                poll_max_retries=6
            )
            print("\n[THÀNH CÔNG] Đã xác thực Smart OTP!")
            print("Access Token:", access_token[:40], "...")
        except Exception as e:
            print("\n[LỖI/TIMEOUT] Phê duyệt Smart OTP thất bại:", e)
    else:
        print("\n[OTP Thường / Smart OTP lấy trực tiếp trên App]")
        user_otp = input("Vui lòng nhập mã OTP 6 số: ").strip()
        if user_otp:
            token = auth.authenticate(otp=user_otp)
            print("\n[THÀNH CÔNG] Đã xác thực mã OTP!")
            print("Access Token:", token.access_token[:40], "...")
