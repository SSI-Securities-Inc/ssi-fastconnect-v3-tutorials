/**
 * Sample 2 — Lay danh sach chi so thi truong (Index)
 * =====================================================
 * Hien thi VN-Index, HNX-Index... tren dashboard.
 *
 * Luong:
 *   1. Client goi endpoint indexList (co the loc theo san)
 *   2. API tra ve danh sach chi so va du lieu gia hien tai
 *   3. Map du lieu sang model hien thi UI
 *   4. Lay summary chi tiet cho mot chi so cu the
 *   5. Cache ngan han de giam so lan goi API
 */

import { Auth, Data, Board } from '@ssi.developer/ssi-sdk';
import { config } from './config.js';
import { ensureAuth } from './auth_helper.js';

const auth = new Auth(config);
await ensureAuth(auth);

const data = new Data(auth);

// --- Buoc 1: Lay toan bo chi so ---
const allIndexes = await data.marketData.getIndexes();
console.log(`Tong so chi so: ${allIndexes.length}\n`);

for (const idx of allIndexes) {
  console.log(`  ${idx.index.padEnd(15)} | ${idx.indexName.padEnd(30)} | San: ${idx.board}`);
}

// --- Buoc 2: Loc chi so theo san HOSE ---
console.log('\n--- Chi so san HOSE ---');
const hoseIndexes = await data.marketData.getIndexesByBoard(Board.HOSE);
for (const idx of hoseIndexes) {
  console.log(`  ${idx.index.padEnd(15)} | ${idx.indexName}`);
}

// --- Buoc 3: Lay chi tiet summary cho mot chi so cu the ---
console.log('\n--- VN-Index Summary ---');
const summary = await data.marketData.getIndexSummary('VNINDEX');
if (summary) {
  console.log(`  Gia tri Index     : ${summary.indexValue}`);
  console.log(`  Thay doi          : ${summary.indexChange} (${summary.indexChangePercent}%)`);
  console.log(`  Tong KL khop      : ${summary.totalMatch}`);
  console.log(`  Tong GT khop      : ${summary.totalMatchValue}`);
  console.log(`  Tang / Giam / Dung: ${summary.totalAdvanceStock} / ${summary.totalDeclineStock} / ${summary.totalSteadyStock}`);
  console.log(`  Tran / San        : ${summary.totalCeilingStock} / ${summary.totalFloorStock}`);
} else {
  console.log('Khong lay duoc summary cho VN-Index.');
}

// --- Response Summary ---
console.log(`\n[Response] total_indexes|hose_indexes|first_index`);
console.log(`${allIndexes.length}|${hoseIndexes.length}|${allIndexes.length > 0 ? allIndexes[0].index : ''}`);
