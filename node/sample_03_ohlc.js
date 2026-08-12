/**
 * Sample 3 — Lay du lieu K-line (OHLC)
 * =======================================
 * Cung cap du lieu nen cho bieu do va phan tich ky thuat.
 *
 * Luong:
 *   1. Client gui symbol, timeframe, from, to
 *   2. API tra ve mang OHLCV theo moc thoi gian
 *   3. Chuan hoa du lieu (date/open/high/low/close/volume)
 *   4. Ho tro nhieu timeframe (1m, 1h, 1d...) intraday & historical
 *   5. Neu lich su dai thi lap theo paging
 */

import { Auth, Data } from '@ssi.developer/ssi-sdk';
import { config } from './config.js';
import { ensureAuth } from './auth_helper.js';

const SYMBOL = 'SSI';

const auth = new Auth(config);
await ensureAuth(auth);

const data = new Data(auth);

const fmtBar = (bar) =>
  `  ${bar.tradingDate} | ` +
  `O:${String(bar.openPrice).padStart(10)} H:${String(bar.highPrice).padStart(10)} ` +
  `L:${String(bar.lowPrice).padStart(10)} C:${String(bar.closePrice).padStart(10)} ` +
  `V:${String(bar.volume).padStart(12)}`;

// --- Buoc 1: Lay OHLC ngay gan nhat ---
console.log(`--- OHLC 1 ngay gan nhat (${SYMBOL}) ---`);
const daily = await data.marketData.getOhlc1DayHistorical(
  SYMBOL, '2026/03/01 00:00:00', '2026/03/27 23:59:59',
);
for (const bar of daily.slice(0, 5)) {
  console.log(fmtBar(bar));
}

// --- Buoc 2: Lay OHLC lich su theo khoang thoi gian (co paging) ---
console.log(`\n--- OHLC 1 ngay lich su (${SYMBOL}) ---`);
const hist = await data.marketData.getOhlc1DayHistorical(
  SYMBOL, '2026/01/01 00:00:00', '2026/03/27 23:59:59', 1, 20,
);
for (const bar of hist) {
  console.log(fmtBar(bar));
}

// --- Buoc 3: Lay OHLC theo timeframe khac (1h) ---
console.log(`\n--- OHLC 1 gio gan nhat (${SYMBOL}) ---`);
const hourly = await data.marketData.getOhlc1Hour(SYMBOL);
for (const bar of hourly.slice(0, 5)) {
  console.log(fmtBar(bar));
}

// --- Buoc 4: Phan trang cho du lieu lon ---
console.log(`\n--- Paging OHLC 1 phut lich su (${SYMBOL}) ---`);
let page = 1;
let totalBars = 0;
while (true) {
  const bars = await data.marketData.getOhlc1MinuteHistorical(
    SYMBOL, '2026/03/25 09:00:00', '2026/03/25 15:00:00', page, 100,
  );
  if (bars.length === 0) break;
  totalBars += bars.length;
  console.log(`  Trang ${page}: ${bars.length} nen (tong: ${totalBars})`);
  page += 1;
}

console.log(`\nTong cong ${totalBars} nen 1 phut duoc tai.`);

// --- Response Summary ---
console.log(`\n[Response] daily_bars|hourly_bars|paging_1min`);
console.log(`${daily.length}|${hourly.length}|${totalBars}`);
if (daily.length > 0) {
  const c = daily[0];
  console.log(`[Response:first_daily] date|open|high|low|close|volume`);
  console.log(`${c.tradingDate}|${c.openPrice}|${c.highPrice}|${c.lowPrice}|${c.closePrice}|${c.volume}`);
}
