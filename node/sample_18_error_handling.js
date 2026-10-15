/**
 * Sample 18 — Xu ly loi, rate limit, timeout va khong lo secret
 * ===============================================================
 * Cach bat dung loai loi cua SDK 3.3.0. Phan lon sample chay OFFLINE (loi duoc tao tai cho hoac bi
 * SDK chan truoc khi gui request); chi buoc 5 goi mot API doc market data (khong can OTP).
 * KHONG gui lenh nao.
 *
 * Luong:
 *   1. classify(): thu tu instanceof dung (lop con phai dung truoc lop cha)
 *   2. Cac loi tao tai cho: RateLimitError.retryAfterMs, DuplicateRequestError, APIError.code vs
 *      ServerErrorCode, SmartOTPPending/Rejected, ReauthenticationRequired
 *   3. Loi bi SDK chan truoc khi gui: ValidationError (lo rong, OHLC 1w/1M), ConfigurationError
 *   4. Timeout / retry cau hinh bang `*Ms`
 *   5. Mau retry khi bi rate limit — CHI cho cac lenh DOC
 *   6. Che secret trong log: mask / redact / Token
 *
 * Luu y SDK 3.3.0:
 *   - Cay loi: SSIError -> { APIError -> DuplicateRequestError, AuthenticationError ->
 *     { SmartOTPPendingError, SmartOTPRejectedError, ReauthenticationRequired }, RateLimitError,
 *     WebSocketError -> TradingWSError, ValidationError, ConfigurationError }.
 *     AuthenticationError KHONG phai con cua APIError.
 *   - Moi thoi gian la mili-giay (`*Ms`): RateLimitError.retryAfterMs, timeoutMs, retryDelayMs,
 *     otpPollIntervalMs, otpPollMaxWaitMs.
 *   - `err.code` la ma cua server (so, vd 400210) hoac ma SDK (chuoi, vd 'RATE_LIMITED'); so sanh
 *     voi ServerErrorCode bang Number(err.code).
 *   - Lenh dat/huy KHONG bao gio duoc SDK retry. Khong tu bao boc lenh dat trong vong retry.
 */

import {
  Auth, Data, Timeframe,
  SSIError, APIError, AuthenticationError, RateLimitError, WebSocketError, ValidationError,
  ConfigurationError, SmartOTPPendingError, SmartOTPRejectedError, ReauthenticationRequired,
  DuplicateRequestError, ServerErrorCode, Trading, mask, redact,
} from '@ssi.developer/ssi-sdk';
import { TradingWSError } from '@ssi.developer/ssi-sdk/experimental';
import { randomUUID } from 'node:crypto';
import { config, assertConfigured } from './config.js';
import { ensureAuth } from './auth_helper.js';

assertConfigured();

// --- Buoc 1: Phan loai loi (lop con TRUOC lop cha) ---
function classify(err) {
  if (err instanceof RateLimitError) {
    return `RATE LIMIT -> cho ${err.retryAfterMs ?? 'khong ro'} ms roi thu lai (lenh doc); limit=${err.limit ?? '?'} remaining=${err.remaining ?? '?'}`;
  }
  if (err instanceof DuplicateRequestError) {
    return `TRUNG ID (HTTP ${err.statusCode}) -> lenh goc co the van song, kiem tra getTodayOrders; KHONG gui lai bang id khac`;
  }
  if (err instanceof SmartOTPPendingError) return 'SMART OTP CHUA DUYET -> cho nguoi dung bam APPROVE tren app, poll tiep';
  if (err instanceof SmartOTPRejectedError) return 'SMART OTP BI TU CHOI/HET HAN -> yeu cau OTP moi (auth.requestOtp)';
  if (err instanceof ReauthenticationRequired) return 'PHIEN HET HAN, KHONG REFRESH DUOC -> xac thuc lai voi OTP moi';
  if (err instanceof AuthenticationError) return `LOI XAC THUC (code ${err.code}) -> kiem tra credential / token`;
  if (err instanceof TradingWSError) return `TRADING WS status=${err.status} code=${err.errorCode}: ${err.serverMessage}`;
  if (err instanceof WebSocketError) return `WEBSOCKET: ${err.message} (mat ket noi/timeout; lenh khong duoc retry)`;
  if (err instanceof ValidationError) return `DU LIEU SAI (chua gui request): ${err.message}`;
  if (err instanceof ConfigurationError) return `CAU HINH SAI: ${err.message}`;
  if (err instanceof APIError) {
    const name = ServerErrorCode[Number(err.code)];
    return `API ERROR HTTP ${err.statusCode} | code=${err.code}${name ? ` (${name})` : ''}: ${err.message}`;
  }
  if (err instanceof SSIError) return `SSI ERROR: ${err.message}`;
  return `LOI KHAC (${err?.name}): ${err?.message}`;
}

// --- Buoc 2: Loi tao tai cho de xem cach xu ly ---
console.log('--- 1-2. Phan loai cac loai loi ---');
const fakeSecret = randomUUID(); // gia tri ngau nhien, chi de minh hoa che secret
const samples = [
  new RateLimitError('Too many requests', 1500, { limit: 10, remaining: 0 }),
  new DuplicateRequestError('clientRequestId already used', 400111, 409),
  new APIError('Timeframe not supported', ServerErrorCode.TIMEFRAME_NOT_SUPPORTED, 400),
  new APIError('OTP verification failed', ServerErrorCode.OTP_VERIFICATION_FAILED, 401),
  new SmartOTPPendingError('Approval pending', ServerErrorCode.PUSH_APPROVAL_PENDING, 202),
  new SmartOTPRejectedError('Approval rejected', ServerErrorCode.PUSH_APPROVAL_REJECTED, 401),
  new ReauthenticationRequired('Refresh token expired'),
  new TradingWSError(429, 'RATE_LIMITED', 'slow down'),
  new WebSocketError('timed out waiting for a response (not retried)'),
  new Error('loi khong phai cua SDK'),
];
for (const err of samples) console.log(`  ${err.constructor.name.padEnd(24)} -> ${classify(err)}`);

// responseBody cua loi duoc SDK che secret san
const withBody = new APIError('Bad request', 400, 400, { apiSecret: fakeSecret, note: 'ok' });
console.log(`\n  responseBody da che: ${JSON.stringify(withBody.responseBody)}`);

// --- Buoc 3: Loi bi chan truoc khi gui request ---
console.log('\n--- 3. Loi bi SDK chan tai may (khong request nao duoc gui) ---');
const auth = new Auth(config);
const trading = new Trading(auth);
const data = new Data(auth);
const offline = {
  'placeBatchOrders([]) (lo rong)': () => trading.trading.placeBatchOrders([]),
  'cancelBatchOrders([]) (lo rong)': () => trading.trading.cancelBatchOrders([]),
  'getOhlc1WeekHistorical (1w khong ho tro)': () =>
    data.marketData.getOhlc1WeekHistorical('SSI', '2026/01/01 00:00:00', '2026/01/31 23:59:59'),
  'new Auth({ khongTonTai: 1 }) (tuy chon la)': () => new Auth({ ...config, khongTonTai: 1 }),
};
for (const [label, fn] of Object.entries(offline)) {
  try {
    await fn();
    console.log(`  ${label}: (khong loi)`);
  } catch (err) {
    console.log(`  ${label}\n    -> ${classify(err)}`);
  }
}
console.log(`  (Khung nen ${Timeframe.WEEK_1} / ${Timeframe.MONTH_1} ton tai trong enum nhung OHLC tu choi chung)`);

// --- Buoc 4: Timeout / retry bang *Ms ---
console.log('\n--- 4. Cau hinh timeout / retry (mili-giay) ---');
const tuned = new Auth({
  ...config,
  timeoutMs: 15_000, // timeout moi request
  maxRetries: 2, // chi retry khi GET bi timeout
  retryDelayMs: 1_000, // back-off co so
  otpPollIntervalMs: 3_000, // chu ky poll Smart OTP
  otpPollMaxWaitMs: 30_000, // tong thoi gian cho Smart OTP
  rateLimitPerSecond: 5, // gioi han phia client
});
const c = tuned.config;
console.log(
  `  timeoutMs=${c.timeoutMs} maxRetries=${c.maxRetries} retryDelayMs=${c.retryDelayMs} ` +
  `otpPollIntervalMs=${c.otpPollIntervalMs} otpPollMaxWaitMs=${c.otpPollMaxWaitMs} rateLimitPerSecond=${c.rateLimitPerSecond}`,
);
console.log('  Lenh dat/huy khong bao gio duoc retry, bat ke cau hinh tren.');

// --- Buoc 5: Retry khi bi rate limit — CHI cho lenh doc ---
async function readWithRateLimitRetry(fn, { retries = 3, fallbackWaitMs = 1_000 } = {}) {
  for (let attempt = 0; ; attempt += 1) {
    try {
      return await fn();
    } catch (err) {
      if (!(err instanceof RateLimitError) || attempt >= retries) throw err;
      const waitMs = err.retryAfterMs ?? fallbackWaitMs * 2 ** attempt;
      console.log(`  Bi rate limit, cho ${waitMs} ms (lan ${attempt + 1}/${retries})...`);
      await new Promise((resolve) => setTimeout(resolve, waitMs));
    }
  }
}

console.log('\n--- 5. Goi API doc co retry rate limit (getMasterData, khong can OTP) ---');
try {
  await ensureAuth(auth);
  const master = await readWithRateLimitRetry(() => data.marketData.getMasterData());
  console.log(`  Nhan ${master.length} ma trong master data.`);
} catch (err) {
  console.log(`  ${classify(err)}`);
  if (err instanceof ReauthenticationRequired) console.log('  -> chay lai sample_01 de xac thuc.');
}

// --- Buoc 6: Khong lo secret trong log ---
console.log('\n--- 6. Che secret trong log ---');
console.log(`  mask(secret)   = ${mask(fakeSecret)}`);
console.log(`  redact(object) = ${JSON.stringify(redact({ apiSecret: fakeSecret, token: fakeSecret, symbol: 'SSI' }))}`);
const token = auth.getToken();
if (token) {
  // console.log / JSON.stringify che token; muon luu file phai dung toDict() va tu bao ve file do
  console.log(`  console.log(token) -> ${JSON.stringify(token)}`);
} else {
  console.log('  (chua co token de minh hoa)');
}
console.log('  Quy tac: khong in nguyen object loi/request/header; chi in name, code, message.');
