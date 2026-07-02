/**
 * auth_helper — Token cache de tai su dung token giua cac lan chay
 * ==================================================================
 * Thay vi goi authenticate() moi lan chay script, module nay:
 *   1. Load token da luu tu file token_cache.json (neu co)
 *   2. Neu chua co -> authenticate lan dau (co the can OTP) va luu xuong file
 *   3. Neu token het han -> refresh va luu lai
 *   4. Neu token con han -> dung truc tiep, khong goi API
 *
 * Cach dung:
 *   import { Auth } from '@ssi.developer/ssi-sdk';
 *   import { config } from './config.js';
 *   import { ensureAuth } from './auth_helper.js';
 *
 *   const auth = new Auth(config);
 *   await ensureAuth(auth);          // market data — khong can OTP
 *   await ensureAuth(auth, OTP);     // trading/streaming — can OTP lan dau
 */

import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const SHARED_TOKEN_FILE = path.join(__dirname, '..', 'shared_token.json');
const TOKEN_FILE = path.join(__dirname, 'token_cache.json');

export function loadToken() {
  for (const file of [SHARED_TOKEN_FILE, TOKEN_FILE]) {
    if (!fs.existsSync(file)) continue;
    try {
      return JSON.parse(fs.readFileSync(file, 'utf-8'));
    } catch {
      continue;
    }
  }
  return null;
}

export function saveToken(token) {
  fs.writeFileSync(TOKEN_FILE, JSON.stringify(token, null, 2), 'utf-8');
  console.log(`Token da luu vao ${TOKEN_FILE}`);
}

/**
 * Dam bao `auth` co token hop le.
 * @param {import('@ssi.developer/ssi-sdk').Auth} auth
 * @param {string} [otp] OTP cho lan authenticate dau tien (sample trading/streaming).
 */
export async function ensureAuth(auth, otp) {
  let token = loadToken();

  if (token == null) {
    console.log('Khong tim thay file token, dang authenticate...');
    token = otp ? await auth.authenticate(otp) : await auth.authenticate();
    saveToken(token);
    console.log('Authenticate thanh cong, token da luu.');
    return;
  }

  auth.setToken(token);

  if (!auth.tokenManager.isAuthenticated()) {
    console.log('Token da het han, dang refresh...');
    token = await auth.refresh();
    saveToken(token);
    console.log('Refresh token thanh cong.');
  } else {
    console.log('Token con han, dung token tu file.');
  }
}
