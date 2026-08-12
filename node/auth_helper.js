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
import readline from 'node:readline';
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

function askQuestion(query) {
  const rl = readline.createInterface({ input: process.stdin, output: process.stdout });
  return new Promise((resolve) => {
    rl.question(query, (ans) => {
      rl.close();
      resolve(ans);
    });
  });
}

/**
 * Dam bao `auth` co token hop le.
 * @param {import('@ssi.developer/ssi-sdk').Auth} auth
 * @param {string} [otp] OTP cho lan authenticate dau tien.
 */
export async function ensureAuth(auth, otp) {
  let token = loadToken();

  if (token != null) {
    auth.setToken(token);
    if (auth.tokenManager.isAuthenticated()) {
      console.log('Token con han, dung token tu file.');
      return;
    }

    console.log('Token da het han, dang refresh...');
    try {
      token = await auth.refresh();
      saveToken(token);
      console.log('Refresh token thanh cong.');
      return;
    } catch (err) {
      console.log(`Refresh token that bai (${err.message}), tien hanh xac thuc lai...`);
    }
  }

  console.log('Khong tim thay token hop le, dang thuc hien quy trinh xac thuc & OTP...');
  if (otp) {
    token = await auth.authenticate(otp);
  } else {
    console.log('=== Yeu cau OTP (Request OTP) ===');
    const otpRes = await auth.requestOtp();
    let transactionId = otpRes?.data?.transactionId || otpRes?.transactionId;

    if (transactionId) {
      console.log(`[Smart OTP] Transaction ID: ${transactionId}`);
      console.log('Vui long mo ung dung SSI tren dien thoai va bam APPROVE (Xac nhan)...');
      console.log('SDK dang Polling cho ban bam phe duyet...');
      await auth.ensureAuthenticated(undefined, transactionId, 5000, 6);
      token = auth.getToken();
    } else {
      const userOtp = await askQuestion('Vui long nhap ma OTP 6 so: ');
      if (userOtp.trim()) {
        token = await auth.authenticate(userOtp.trim());
      }
    }
  }

  if (auth.getToken()) {
    saveToken(auth.getToken());
    console.log('Authenticate thanh cong, token da luu.');
  }
}

