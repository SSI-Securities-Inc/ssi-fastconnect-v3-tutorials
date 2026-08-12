"""
auth_helper — Token cache để tái sử dụng token giữa các lần chạy
==================================================================
Thay vì gọi authenticate() mỗi lần chạy script, module này:
  1. Load token đã lưu từ file token_cache.json (nếu có)
  2. Nếu chưa có → authenticate lần đầu và lưu xuống file
  3. Nếu token hết hạn → refresh và lưu lại
  4. Nếu token còn hạn → dùng trực tiếp, không gọi API

Cách dùng (sync):
    from auth_helper import ensure_auth
    with Auth(config) as auth:
        ensure_auth(auth)

Cách dùng (async):
    from auth_helper import ensure_auth_async
    async with AsyncAuth(config) as auth:
        await ensure_auth_async(auth)
"""

import json
import os

from ssi_sdk.models import Token

TOKEN_FILE = os.path.join(os.path.dirname(__file__), "token_cache.json")


def load_token() -> Token | None:
    """Load token từ file, trả về None nếu file không tồn tại."""
    if not os.path.exists(TOKEN_FILE):
        return None
    with open(TOKEN_FILE, "r", encoding="utf-8") as f:
        _token = json.load(f)
    return Token.from_dict(_token)


def save_token(token: Token) -> None:
    """Lưu token xuống file."""
    with open(TOKEN_FILE, "w", encoding="utf-8") as f:
        json.dump(token.to_dict(), f, indent=2)
    print(f"Token đã lưu vào {TOKEN_FILE}")


def ensure_auth(auth, otp: str | None = None) -> None:
    """Đảm bảo auth có token hợp lệ (sync)."""
    cached_token = load_token()
    if cached_token is not None:
        auth.token_manager.set_token(cached_token)
        if not auth.is_token_expired:
            print("Token còn hạn, dùng token từ file.")
            return
        print("Token đã hết hạn, đang refresh...")
        try:
            cached_token = auth.refresh()
            save_token(cached_token)
            print("Refresh token thành công.")
            return
        except Exception as e:
            print(f"Refresh token thất bại ({e}), tiến hành xác thực lại...")

    print("Không tìm thấy token hợp lệ, đang thực hiện quy trình xác thực & OTP...")
    if otp:
        cached_token = auth.authenticate(otp=otp)
    else:
        print("=== Yêu cầu OTP (Request OTP) ===")
        otp_res = auth.request_otp()
        transaction_id = None
        if isinstance(otp_res, dict):
            data_map = otp_res.get("data", {})
            if isinstance(data_map, dict):
                transaction_id = data_map.get("transactionId") or data_map.get("transaction_id")
            if not transaction_id:
                transaction_id = otp_res.get("transactionId") or otp_res.get("transaction_id")

        if transaction_id:
            print(f"[Smart OTP] Transaction ID: {transaction_id}")
            print("Vui lòng mở app SSI iBoard trên điện thoại và bấm APPROVE (Xác nhận)...")
            print("SDK đang Polling chờ bạn bấm phê duyệt...")
            auth.ensure_authenticated(transaction_id=transaction_id, poll_interval=5, poll_max_retries=6)
        else:
            user_otp = input("Vui lòng nhập mã OTP 6 số: ").strip()
            auth.authenticate(otp=user_otp)

    if auth.token:
        save_token(auth.token)
        print("Authenticate thành công, token đã lưu.")


async def ensure_auth_async(auth, otp: str | None = None) -> None:
    """Đảm bảo auth có token hợp lệ (async)."""
    cached_token = load_token()
    if cached_token is not None:
        await auth.token_manager.set_token(cached_token)
        if not auth.is_token_expired:
            print("Token còn hạn, dùng token từ file.")
            return
        print("Token đã hết hạn, đang refresh...")
        try:
            cached_token = await auth.refresh()
            save_token(cached_token)
            print("Refresh token thành công.")
            return
        except Exception as e:
            print(f"Refresh token thất bại ({e}), tiến hành xác thực lại...")

    print("Không tìm thấy token hợp lệ, đang thực hiện quy trình xác thực & OTP...")
    if otp:
        cached_token = await auth.authenticate(otp=otp)
    else:
        print("=== Yêu cầu OTP (Request OTP) ===")
        otp_res = await auth.request_otp()
        transaction_id = None
        if isinstance(otp_res, dict):
            data_map = otp_res.get("data", {})
            if isinstance(data_map, dict):
                transaction_id = data_map.get("transactionId") or data_map.get("transaction_id")
            if not transaction_id:
                transaction_id = otp_res.get("transactionId") or otp_res.get("transaction_id")

        if transaction_id:
            print(f"[Smart OTP] Transaction ID: {transaction_id}")
            print("Vui lòng mở app SSI iBoard trên điện thoại và bấm APPROVE (Xác nhận)...")
            print("SDK đang Polling chờ bạn bấm phê duyệt...")
            await auth.ensure_authenticated(transaction_id=transaction_id, poll_interval=5, poll_max_retries=6)
        else:
            user_otp = input("Vui lòng nhập mã OTP 6 số: ").strip()
            await auth.authenticate(otp=user_otp)

    if auth.token:
        save_token(auth.token)
        print("Authenticate thành công, token đã lưu.")

