/**
 * Sample 19 — Su kien doanh nghiep (Corporate actions)
 * =====================================================
 * Xem co tuc, co phieu thuong, quyen mua... cua cac ma. Chi DOC du lieu, khong dat lenh.
 *
 * Luong:
 *   1. getCorporateActions(): su kien cua HOM NAY
 *   2. getCorporateActionsHistorical(from, to): su kien trong mot khoang ngay o qua khu
 *      (server cho toi da 5 ngay moi lan goi)
 *   3. Loc theo ma (symbol) va ngon ngu ('vn' / 'en')
 *   4. Khoang ngay qua rong bi SDK chan tai client (ValidationError), khong goi server
 *
 * Luu y:
 *   - SDK tu lay het cac trang (pageSize 100), khong can tu phan trang.
 *   - Moi truong la chuoi dung nhu server gui. Ngay co dang YYYY/MM/DD, ratio KHONG duoc parse
 *     vi y nghia phu thuoc loai su kien. Truong server de trong la ''.
 *   - Ngay truyen vao phai dung dang YYYY/MM/DD (khong dung dau '-').
 *   - "Hom nay" lay theo ngay cua may ban.
 */

import { Auth, Data, ValidationError } from '@ssi.developer/ssi-sdk';
import { config } from './config.js';
import { ensureAuth } from './auth_helper.js';

const auth = new Auth(config);
await ensureAuth(auth);

const data = new Data(auth);

const pad2 = (n) => String(n).padStart(2, '0');
const fmtDate = (d) => `${d.getFullYear()}/${pad2(d.getMonth() + 1)}/${pad2(d.getDate())}`;
const daysAgo = (n) => {
  const d = new Date();
  d.setDate(d.getDate() - n);
  return d;
};

const show = (actions, limit = 10) => {
  for (const a of actions.slice(0, limit)) {
    console.log(
      `  ${a.symbol.padEnd(8)} | ${a.exchange.padEnd(6)} | GDKHQ: ${(a.exrightDate || '-').padEnd(10)} ` +
      `| Chot: ${(a.recordDate || '-').padEnd(10)} | Phat hanh: ${(a.issueDate || '-').padEnd(10)} ` +
      `| Ty le: ${(a.ratio || '-').padEnd(8)} | ${a.eventTitle}`,
    );
  }
  if (actions.length > limit) console.log(`  ... con ${actions.length - limit} su kien`);
};

// --- Buoc 1: Su kien cua hom nay ---
console.log('--- Su kien doanh nghiep hom nay ---');
const todayActions = await data.marketData.getCorporateActions();
console.log(`Tong so su kien: ${todayActions.length}`);
show(todayActions);

// --- Buoc 2: Su kien trong 5 ngay gan nhat (toi da 5 ngay cho moi lan goi) ---
const fromDate = fmtDate(daysAgo(4));
const toDate = fmtDate(daysAgo(0));
console.log(`\n--- Su kien tu ${fromDate} den ${toDate} ---`);
const rangeActions = await data.marketData.getCorporateActionsHistorical(fromDate, toDate);
console.log(`Tong so su kien: ${rangeActions.length}`);
show(rangeActions);

// --- Buoc 3: Loc theo ma, tra tieu de bang tieng Anh ---
console.log('\n--- Su kien cua ma SSI (tieng Anh) ---');
const ssiActions = await data.marketData.getCorporateActionsHistorical(fromDate, toDate, 'SSI', 'en');
console.log(`Tong so su kien: ${ssiActions.length}`);
show(ssiActions);

// --- Buoc 4: Khoang ngay qua rong bi chan tai SDK ---
console.log('\n--- Khoang 30 ngay (vuot gioi han 5 ngay) ---');
try {
  await data.marketData.getCorporateActionsHistorical(fmtDate(daysAgo(30)), toDate);
} catch (err) {
  if (!(err instanceof ValidationError)) throw err;
  console.log(`  SDK tu choi truoc khi goi server: ${err.message}`);
  console.log('  -> chia khoang ngay thanh cac doan <= 5 ngay roi goi tung doan.');
}

// --- Response Summary ---
console.log('\n[Response] today_count|range_count|ssi_range_count');
console.log(`${todayActions.length}|${rangeActions.length}|${ssiActions.length}`);
