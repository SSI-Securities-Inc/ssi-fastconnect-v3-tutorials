"""
Sample 1 (Async) — Xác thực và lấy Access Token
==================================================
Đăng nhập và lấy token cho toàn bộ API call sau đó (phiên bản async).

Luồng:
  1. Tạo Config → AsyncAuth → authenticate(otp=...)
  2. Auth service trả về accessToken, refreshToken, expiresIn
  3. Lưu token vào session/runtime store
  4. Mọi request sau đó gắn Authorization: Bearer <accessToken>
  5. Nếu token hết hạn thì gọi refresh/re-login
"""

import asyncio

from ssi_sdk import AsyncAuth, AsyncTrading, Config
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
        # --- Bước 1-2: Xác thực, nhận accessToken + refreshToken ---
        await ensure_auth_async(auth)
        token = auth.token
        print("Access Token :", token.access_token[:40], "...")
        print("Token Type   :", token.token_type)
        print("Expires At   :", token.expires_at)
        print("Refresh Token:", token.refresh_token[:40] if token.refresh_token else "N/A", "...")

        # --- Bước 3: Token đã được SDK lưu tự động, mọi request kế tiếp ---
        #     sẽ gắn header Authorization: Bearer <accessToken>

        # --- Bước 4: Kiểm tra token hết hạn & refresh ---
        if auth.is_token_expired:
            print("\nToken hết hạn, đang refresh...")
            new_token = await auth.refresh()
            print("Token mới    :", new_token.access_token[:40], "...")

        # --- Xác nhận token hoạt động bằng cách gọi API ---
        async with AsyncTrading(auth) as trading:
            accounts = await trading.account.get_account_info()
            print(f"\nXác thực thành công! Tìm thấy {len(accounts)} tài khoản:")
            for acc in accounts:
                print(f"  - {acc.account_no} ({acc.account_type.value})")


asyncio.run(main())
