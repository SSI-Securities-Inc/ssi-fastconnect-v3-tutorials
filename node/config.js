/**
 * config — Cấu hình kết nối dùng chung cho toàn bộ sample
 * =========================================================
 * Đọc tự động từ tệp `config.json` tại gốc dự án (ssi-fastconnect-v3-tutorials/config.json).
 * Nếu không tìm thấy file, sử dụng giá trị mặc định / UAT sandbox.
 */

import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const configPath = path.join(__dirname, '..', 'config.json');

let clientConfig = {
  clientId: '<CLIENT_ID>',
  apiKey: '<API_KEY>',
  apiSecret: '<API_SECRET>',
  privateKey: '<PRIVATE_KEY_CONTENT>',
};
let accountNo = '<ACCOUNT_NO>';
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
    accountNo = raw.equity_account || accountNo;
    otp = raw.otp || otp;
  } catch (e) {
    console.warn('Lỗi khi đọc config.json:', e.message);
  }
}

export const config = clientConfig;
export const ACCOUNT_NO = accountNo;
export const OTP = otp;
