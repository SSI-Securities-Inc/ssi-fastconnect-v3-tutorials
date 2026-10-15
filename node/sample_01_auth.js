/**
 * Sample 1 — Xác thực, Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
 * ==============================================================================
 * Đăng nhập, xử lý luồng OTP (Smart OTP Push Polling / OTP 6 số) và lấy token cho toàn bộ API call sau đó.
 *
 * Luồng:
 *   1. Tạo Config -> Auth
 *   2. Kiểm tra token cache / refresh token nếu có
 *   3. Nếu chưa có token: Gọi requestOtp(), Polling Smart OTP hoặc nhập mã OTP 6 số
 *   4. Auth service trả về accessToken, refreshToken, expiresAt và lưu token_cache.json
 *   5. Mọi request sau đó gắn Authorization: Bearer <accessToken>
 *   6. Kiểm tra token hoạt động bằng cách gọi API lấy thông tin tài khoản
 */

import { Auth, Trading } from '@ssi.developer/ssi-sdk';
import { config, maskAccount } from './config.js';
import { ensureAuth } from './auth_helper.js';

const auth = new Auth(config);

// --- Bước 1-3: Xác thực + Yêu cầu & Xác thực OTP / Smart OTP ---
await ensureAuth(auth);
const token = auth.getToken();
console.log('\n--- Thông tin Token ---');
// Khong in noi dung token ra log — chi in thong tin khong nhay cam.
console.log('Access Token :', token.accessToken ? `(da an, ${token.accessToken.length} ky tu)` : 'N/A');
console.log('Token Type   :', token.tokenType);
console.log('Expires At   :', token.expiresAt);
console.log('Refresh Token:', token.refreshToken ? `(da an, ${token.refreshToken.length} ky tu)` : 'N/A');
console.log('Refresh Exp  :', token.refreshExpiresAt);

// --- Bước 4: Kiểm tra token hết hạn & refresh ---
if (!auth.tokenManager.isAuthenticated()) {
  console.log('\nToken hết hạn, đang refresh...');
  const newToken = await auth.refresh();
  console.log('Token mới    :', `(da an, ${newToken.accessToken.length} ky tu)`);
}

// --- Bước 5: Xác nhận token hoạt động bằng cách gọi API ---
const trading = new Trading(auth);
const accounts = await trading.account.getAccountInfo();
console.log(`\nXác thực thành công! Tìm thấy ${accounts.length} tài khoản:`);
for (const acc of accounts) {
  console.log(`  - ${maskAccount(acc.accountNo)} (${acc.accountType})`);
}

// --- Response Summary ---
console.log(`\n[Response] token_type|expires_at|account_count`);
console.log(`${token.tokenType}|${token.expiresAt}|${accounts.length}`);
console.log(`[Response:account] account_no|account_type`);
for (const acc of accounts) {
  console.log(`${maskAccount(acc.accountNo)}|${acc.accountType}`);
}

