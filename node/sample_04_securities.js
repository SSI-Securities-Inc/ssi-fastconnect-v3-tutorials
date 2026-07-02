/**
 * Sample 4 — Lay danh sach co phieu theo san
 * =============================================
 * Tao watchlist/screener theo tieu chi thi truong.
 *
 * Luong:
 *   1. Client goi securitiesByBoard theo san / index
 *   2. API tra ve danh sach ma + thong tin giao dich co ban
 *   3. Loc/sap xep theo nhu cau UI
 *   4. Xem thong tin chi tiet mot ma
 *   5. Khi user chon ma, chuyen sang luong xem chi tiet/dat lenh
 */

import { Auth, Data, Board } from '@ssi.developer/ssi-sdk';
import { config } from './config.js';
import { ensureAuth } from './auth_helper.js';

const auth = new Auth(config);
await ensureAuth(auth);

const data = new Data(auth);

const fmtSec = (sec) =>
  `  ${sec.symbol.padEnd(10)} | ${(sec.symbolNameVi || sec.symbolNameEn || '').padEnd(30)} ` +
  `| Lot: ${sec.lotSize}`;

// --- Buoc 1: Lay danh sach co phieu san HOSE ---
console.log('--- Co phieu san HOSE ---');
const hoseSecurities = await data.marketData.getSecuritiesInfoByBoard(Board.HOSE);
console.log(`Tong so ma: ${hoseSecurities.length}\n`);
for (const sec of hoseSecurities.slice(0, 10)) {
  console.log(fmtSec(sec));
}

// --- Buoc 2: Lay danh sach co phieu san HNX ---
console.log('\n--- Co phieu san HNX ---');
const hnxSecurities = await data.marketData.getSecuritiesInfoByBoard(Board.HNX);
console.log(`Tong so ma: ${hnxSecurities.length}`);
for (const sec of hnxSecurities.slice(0, 10)) {
  console.log(fmtSec(sec));
}

// --- Buoc 3: Lay theo chi so (index) ---
console.log('\n--- Co phieu thuoc VN30 ---');
const vn30Securities = await data.marketData.getSecuritiesInfoByIndex('VN30');
console.log(`Tong so ma: ${vn30Securities.length}`);
for (const sec of vn30Securities) {
  console.log(`  ${sec.symbol.padEnd(10)} | ${sec.symbolNameVi || sec.symbolNameEn || ''}`);
}

// --- Buoc 4: Xem thong tin chi tiet mot ma ---
console.log('\n--- Chi tiet ma SSI ---');
const info = await data.marketData.getSecuritiesInfo('SSI');
if (info) {
  console.log(`  Ma           : ${info.symbol}`);
  console.log(`  Ten (VI)     : ${info.symbolNameVi}`);
  console.log(`  Ten (EN)     : ${info.symbolNameEn}`);
  console.log(`  San          : ${info.board}`);
  console.log(`  Lot size     : ${info.lotSize}`);
  console.log(`  ICB Code     : ${info.icbCode}`);
  console.log(`  ICB Name     : ${info.icbName}`);
  console.log(`  Listed Shares: ${info.listedShares}`);
} else {
  console.log('Khong tim thay thong tin ma SSI.');
}

// --- Response Summary ---
console.log(`\n[Response] hose_count|hnx_count|vn30_count|symbol|lot_size|listed_shares`);
console.log(`${hoseSecurities.length}|${hnxSecurities.length}|${vn30Securities.length}|${info?.symbol ?? ''}|${info?.lotSize ?? ''}|${info?.listedShares ?? ''}`);
