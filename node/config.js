/**
 * config — Cấu hình kết nối dùng chung cho toàn bộ sample
 * =========================================================
 * Đọc từ tệp `config.json` tại gốc dự án (ssi-fastconnect-v3-tutorials/config.json,
 * mẫu: config.example.json). Tệp này đã được .gitignore — KHÔNG commit secret vào code.
 * Nếu thiếu file, các giá trị là placeholder `<...>`; sample mới gọi `assertConfigured()`
 * để dừng ngay (fail fast) thay vì gửi request với placeholder.
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const configPath = path.join(__dirname, '..', 'config.json');

const PLACEHOLDERS = new Set(['<CLIENT_ID>', '<API_KEY>', '<API_SECRET>', '<PRIVATE_KEY_CONTENT>', '<ACCOUNT_NO>', '<DERIVATIVE_ACCOUNT_NO>', '<OTP>']);

let clientConfig = {
  clientId: '<CLIENT_ID>',
  apiKey: '<API_KEY>',
  apiSecret: '<API_SECRET>',
  privateKey: '<PRIVATE_KEY_CONTENT>',
};
let accountNo = '<ACCOUNT_NO>';
let derivativeAccountNo = '<DERIVATIVE_ACCOUNT_NO>';
let otp = '<OTP>';

if (fs.existsSync(configPath)) {
  try {
    const raw = JSON.parse(fs.readFileSync(configPath, 'utf8'));
    clientConfig = {
      clientId: raw.client_id || clientConfig.clientId,
      apiKey: raw.api_key || clientConfig.apiKey,
      apiSecret: raw.api_secret || clientConfig.apiSecret,
      privateKey: raw.private_key || clientConfig.privateKey,
    };
    // log_level (DEBUG/INFO/...) la tuy chon; SDK 3.3.0 che secret trong log o moi muc.
    if (raw.log_level) clientConfig.logLevel = raw.log_level;
    accountNo = raw.equity_account || accountNo;
    derivativeAccountNo = raw.derivative_account || derivativeAccountNo;
    otp = raw.otp || otp;
  } catch (e) {
    console.warn('Lỗi khi đọc config.json:', e.message);
  }
}

export const config = clientConfig;
export const ACCOUNT_NO = accountNo;
export const DERIVATIVE_ACCOUNT_NO = derivativeAccountNo;
export const OTP = otp;

/** In tai khoan dang che (chi giu 4 so cuoi) de khong lo so tai khoan len log. */
export function maskAccount(value) {
  const text = String(value ?? '');
  return text.length <= 4 ? '***' : `***${text.slice(-4)}`;
}

/**
 * Dung ngay neu cac gia tri bat buoc van la placeholder (chua co config.json).
 * @param {{ trading?: boolean, derivative?: boolean }} [need]
 */
export function assertConfigured({ trading = false, derivative = false } = {}) {
  const missing = [];
  for (const [name, value] of Object.entries(clientConfig)) {
    if (name !== 'logLevel' && PLACEHOLDERS.has(String(value))) missing.push(name);
  }
  if (trading && PLACEHOLDERS.has(ACCOUNT_NO)) missing.push('equity_account');
  if (derivative && PLACEHOLDERS.has(DERIVATIVE_ACCOUNT_NO)) missing.push('derivative_account');
  if (missing.length > 0) {
    throw new Error(`Thiếu cấu hình trong config.json: ${missing.join(', ')} (xem config.example.json)`);
  }
}
