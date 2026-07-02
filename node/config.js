/**
 * config — Cau hinh ket noi dung chung cho toan bo sample
 * =========================================================
 * Tat ca sample import config tu day thay vi nhung lai trong tung file.
 *
 * Day la thong tin moi truong UAT (sandbox), nhung truc tiep trong code
 * giong nhu sample python/go. Khi chuyen sang production, thay toan bo bang
 * thong tin that do SSI cap (nen dung vault/env var, khong commit key that).
 */

export const config = {
  clientId: '<CLIENT_ID>',
  apiKey: '<API_KEY>',
  apiSecret: '<API_SECRET>',
  // RSA private key (Base64 XML) — dung de ky lenh (sample trading 05-09, 11, 12).
  privateKey: '<PRIVATE_KEY_CONTENT>',
};

// Tai khoan giao dich dung cho cac sample trading & streaming.
export const ACCOUNT_NO = '<ACCOUNT_NO>';

// OTP chi can o lan authenticate dau tien (khi chua co token_cache.json hop le).
// Cac lan sau SDK tu refresh bang refresh token, khong can OTP.
export const OTP = '<OTP>';
