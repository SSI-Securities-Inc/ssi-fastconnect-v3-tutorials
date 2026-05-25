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
    if cached_token is None:
        print("Không tìm thấy file token, đang authenticate...")
        cached_token = auth.token_manager.authenticate(otp=otp) if otp else auth.token_manager.authenticate()
        save_token(cached_token)
        print("Authenticate thành công, token đã lưu.")
    auth.token_manager.set_token(cached_token)
    if auth.token_manager.is_token_expired:
        print("Token đã hết hạn, đang refresh...")
        cached_token = auth.token_manager.refresh()
        save_token(cached_token)
        print("Refresh token thành công.")
    else:
        print("Token còn hạn, dùng token từ file.")


async def ensure_auth_async(auth, otp: str | None = None) -> None:
    """Đảm bảo auth có token hợp lệ (async)."""
    cached_token = load_token()
    if cached_token is None:
        print("Không tìm thấy file token, đang authenticate...")
        cached_token = await auth.token_manager.authenticate(otp=otp) if otp else await auth.token_manager.authenticate()
        save_token(cached_token)
        print("Authenticate thành công, token đã lưu.")
    await auth.token_manager.set_token(cached_token)
    if auth.token_manager.is_token_expired:
        print("Token đã hết hạn, đang refresh...")
        cached_token = await auth.token_manager.refresh()
        save_token(cached_token)
        print("Refresh token thành công.")
    else:
        print("Token còn hạn, dùng token từ file.")
