"""
Sample 18 — Xử lý lỗi đúng cách (SDK 3.3.0)
==============================================
Cách bắt và phản ứng với từng loại exception của SDK. Phần A và B chạy OFFLINE (không gọi mạng);
phần C gọi API đọc dữ liệu thật (không đặt lệnh).

Phân cấp exception (bắt CON trước, CHA sau):
  SSIError                      message, code, status_code, response_body, headers
   ├─ AuthenticationError
   │    ├─ SmartOTPPendingError     chưa bấm APPROVE trên app (HTTP 202 / 401114)
   │    ├─ SmartOTPRejectedError    bị từ chối / hết hạn / transaction lạ → xin OTP mới
   │    └─ ReauthenticationRequired không refresh được phiên → xác thực lại bằng OTP mới
   ├─ APIError                     code = mã lỗi server (vd "400210") hoặc status HTTP dạng chuỗi
   │    └─ DuplicateRequestError    HTTP 409, clientRequestId/batchRequestId đã dùng trong ngày
   ├─ RateLimitError               retry_after_ms (mili-giây), limit, remaining
   ├─ ValidationError              dữ liệu sai, bị chặn ngay phía client (chưa gửi request)
   └─ WebSocketError
        └─ TradingWSError          status, error_code, server_message (xem sample 17)

Quy tắc:
  - Mọi thời gian ở SDK 3.3.0 là mili-giây, tên có hậu tố *_ms. Tên cũ (timeout, retry_delay,
    poll_interval, interval...) báo TypeError.
  - SDK chỉ tự retry request GET. Lệnh đặt/hủy (POST/DELETE) KHÔNG tự retry: gặp lỗi mạng hãy tra sổ
    lệnh trước, đừng gửi lại mù quáng. DuplicateRequestError nghĩa là lệnh gốc có thể đang nằm trên sàn.
  - APIError.code có thể là chuỗi hoặc số: so sánh qua ServerErrorCode.from_value(code)
    (trả None với mã lạ, không raise).
  - KHÔNG in e.headers / e.response_body / config: có thể chứa thông tin nhạy cảm (token, tài khoản).
    Chỉ in message, code, status_code.
"""

import dataclasses
import time

from ssi_sdk import Auth, Data
from ssi_sdk.enums import ServerErrorCode
from ssi_sdk.exceptions import (
    APIError,
    AuthenticationError,
    DuplicateRequestError,
    RateLimitError,
    ReauthenticationRequired,
    SmartOTPPendingError,
    SmartOTPRejectedError,
    SSIError,
    ValidationError,
)
from auth_helper import ensure_auth
from config import config, OTP

MAX_ATTEMPTS = 3
MAX_WAIT_MS = 10_000


def explain(e: Exception) -> str:
    """Trả về hướng xử lý cho một exception. Thứ tự except: con trước, cha sau."""
    try:
        raise e
    except RateLimitError as err:
        return (f"Chạm rate limit (limit={err.limit}, remaining={err.remaining}). "
                f"Đợi {err.retry_after_ms} ms rồi gọi lại (chỉ với request đọc).")
    except DuplicateRequestError as err:  # phải đứng TRƯỚC APIError
        return f"409 trùng id ({err.message}). Lệnh gốc có thể đã nằm trên sàn: tra sổ lệnh, KHÔNG gửi lại."
    except ValidationError as err:
        return f"Dữ liệu không hợp lệ, chưa gửi request: {err.message}. Sửa input rồi chạy lại."
    except APIError as err:
        known = ServerErrorCode.from_value(err.code)
        name = known.name if known else "mã chưa có trong SDK"
        return f"Server từ chối: code={err.code} ({name}) status={err.status_code}: {err.message}"
    except SmartOTPPendingError:  # các lớp con của AuthenticationError đứng TRƯỚC lớp cha
        return "Chưa bấm APPROVE trên app SSI iBoard: bấm APPROVE rồi chờ tiếp, hoặc tăng số lần poll."
    except SmartOTPRejectedError as err:
        return f"Smart OTP bị từ chối/hết hạn ({err.message}): yêu cầu OTP mới."
    except ReauthenticationRequired:
        return "Không refresh được phiên: xác thực lại bằng OTP mới."
    except AuthenticationError as err:
        return f"Lỗi xác thực: {err.message}"
    except SSIError as err:
        return f"Lỗi SDK khác ({type(err).__name__}): {err.message}"


def call_with_retry(fn, *args, **kwargs):
    """Gọi request ĐỌC, tôn trọng retry_after_ms khi bị rate limit. KHÔNG dùng cho đặt/hủy lệnh."""
    for attempt in range(1, MAX_ATTEMPTS + 1):
        try:
            return fn(*args, **kwargs)
        except RateLimitError as e:
            if attempt == MAX_ATTEMPTS:
                raise
            wait_ms = min(e.retry_after_ms or config.retry_delay_ms, MAX_WAIT_MS)
            print(f"  Rate limit, đợi {wait_ms:.0f} ms (lần {attempt}/{MAX_ATTEMPTS})...")
            time.sleep(wait_ms / 1000)


# --- Phần A (offline): cấu hình theo *_ms ---
print("--- A. Cấu hình theo mili-giây ---")
print(f"timeout_ms={config.timeout_ms} retry_delay_ms={config.retry_delay_ms} "
      f"otp_poll_interval_ms={config.otp_poll_interval_ms} otp_poll_max_wait_ms={config.otp_poll_max_wait_ms}")
try:
    dataclasses.replace(config, timeout=5)  # tên cũ không còn
except TypeError as e:
    print(f"Tên cũ 'timeout' bị từ chối đúng như mong đợi: {e}")

# --- Phần B (offline): từng loại lỗi dùng exception giả ---
print("\n--- B. Hướng xử lý từng loại lỗi (exception giả, không gọi mạng) ---")
samples = [
    RateLimitError("Too many requests", retry_after_ms=1500, limit=10, remaining=0),
    ValidationError("price phải > 0"),
    DuplicateRequestError("clientRequestId đã dùng", code="409", status_code=409),
    APIError("Timeframe không hỗ trợ", code="400210", status_code=400),
    APIError("Mã lỗi lạ", code="999999", status_code=400),
    SmartOTPPendingError("chưa approve"),
    SmartOTPRejectedError("transaction hết hạn"),
    ReauthenticationRequired("refresh token bị thu hồi"),
]
for exc in samples:
    print(f"- {type(exc).__name__}: {explain(exc)}")

# --- Phần C (thật, chỉ đọc): xác thực + gọi API + bắt lỗi ---
print("\n--- C. Gọi API thật (chỉ đọc) ---")
with Auth(config) as auth:
    try:
        ensure_auth(auth, otp=OTP)
    except (SmartOTPPendingError, SmartOTPRejectedError, ReauthenticationRequired, AuthenticationError) as e:
        raise SystemExit(f"Xác thực thất bại: {explain(e)}")

    with Data(auth) as data:
        try:
            rows = call_with_retry(data.market_data.get_master_data)
            print(f"get_master_data: {len(rows)} mã")
        except SSIError as e:
            print(f"get_master_data lỗi: {explain(e)}")

        # Nến tuần/tháng không hỗ trợ: lỗi ValidationError ngay phía client
        try:
            data.market_data.get_ohlc_1week_historical("SSI", "2026/01/01 00:00:00", "2026/03/27 23:59:59")
        except SSIError as e:
            print(f"get_ohlc_1week_historical: {explain(e)}")
