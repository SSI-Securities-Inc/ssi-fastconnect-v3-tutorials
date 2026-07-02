/**
 * Sample 1 — Xac thuc va lay Access Token
 * =========================================
 * Dang nhap va lay token cho toan bo API call sau do.
 *
 * Luong:
 *   1. Tao Config -> Auth -> authenticate(otp=...)
 *   2. Auth service tra ve accessToken, refreshToken, expiresAt
 *   3. Luu token vao token_cache.json de tai su dung
 *   4. Moi request sau do gan Authorization: Bearer <accessToken>
 *   5. Neu token het han thi goi refresh/re-login
 */

import { Auth, Trading } from '@ssi.developer/ssi-sdk';
import { config } from './config.js';
import { ensureAuth } from './auth_helper.js';

const auth = new Auth(config);

// --- Buoc 1-2: Xac thuc, nhan accessToken + refreshToken ---
await ensureAuth(auth);
const token = auth.getToken();
console.log('Access Token :', token.accessToken.slice(0, 40), '...');
console.log('Token Type   :', token.tokenType);
console.log('Expires At   :', token.expiresAt);
console.log('Refresh Token:', token.refreshToken ? `${token.refreshToken.slice(0, 40)} ...` : 'N/A');

// --- Buoc 3: Token da duoc SDK luu tu dong, moi request ke tiep ---
//     se gan header Authorization: Bearer <accessToken>

// --- Buoc 4: Kiem tra token het han & refresh ---
if (!auth.tokenManager.isAuthenticated()) {
  console.log('\nToken het han, dang refresh...');
  const newToken = await auth.refresh();
  console.log('Token moi    :', newToken.accessToken.slice(0, 40), '...');
}

// --- Xac nhan token hoat dong bang cach goi API ---
const trading = new Trading(auth);
const accounts = await trading.account.getAccountInfo();
console.log(`\nXac thuc thanh cong! Tim thay ${accounts.length} tai khoan:`);
for (const acc of accounts) {
  console.log(`  - ${acc.accountNo} (${acc.accountType})`);
}

// --- Response Summary ---
console.log(`\n[Response] token_type|expires_at|account_count`);
console.log(`${token.tokenType}|${token.expiresAt}|${accounts.length}`);
console.log(`[Response:account] account_no|account_type`);
for (const acc of accounts) {
  console.log(`${acc.accountNo}|${acc.accountType}`);
}
