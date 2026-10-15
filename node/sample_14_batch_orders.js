/**
 * Sample 14 — Dat / huy lenh theo lo (batch)
 * ============================================
 * Gui nhieu lenh trong MOT request da ky (placeBatchOrders) va huy nhieu lenh trong MOT request
 * (cancelBatchOrders).
 *
 * CANH BAO: sample nay GUI LENH THAT len tai khoan trong config.json. De an toan, no chi dat lenh
 * MUA Limit (LO) tai GIA SAN (xa thi truong, thuong nam cho khop) roi huy ngay. Chi chay tren tai
 * khoan thu nghiem hoac khi ban hieu ro rui ro.
 *
 * Luong:
 *   1. Doc gia san cua ma SSI tu master data (getMasterData) — khong hardcode gia
 *   2. Gioi han lo: Config.maxBatchOrders (mac dinh 20). Lo lon hon bi tu choi NGAY TAI MAY
 *      (ValidationError), khong co request nao duoc gui
 *   3. placeBatchOrders: moi lenh co clientRequestId (<= 20 ky tu) lam khoa chong trung
 *   4. Doc ket qua TUNG lenh (success / orderId / status / errorCode / errorMessage)
 *   5. cancelBatchOrders cac lenh vua dat (theo orderId)
 *   6. (tuy chon, BATCH_DUP_DEMO=1) gui lai cung clientRequestId de xem co che chong trung
 *
 * Luu y SDK 3.3.0:
 *   - Server chap nhan hoac tu choi CA LO: mot lenh sai thi ca lo bi tu choi (SDK kiem tra truoc).
 *   - Lo khong bao gio tu dong retry. DuplicateRequestError (HTTP 409) = id da dung trong ngay,
 *     lenh goc co the van dang song -> kiem tra getTodayOrders, KHONG gui lai bang id khac.
 *   - Gia truyen dang chuoi (PriceLike), khong dung so thuc.
 */

import {
  Auth, Data, Trading, OrderSide, OrderType, DuplicateRequestError, ValidationError,
} from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP, maskAccount, assertConfigured } from './config.js';
import { ensureAuth } from './auth_helper.js';

assertConfigured({ trading: true });

const SYMBOL = 'SSI';
const QUANTITY = 100;
const ORDER_COUNT = 2;

function printResults(title, response) {
  console.log(`  ${title} | batchRequestId: ${response.batchRequestId}`);
  for (const [i, r] of response.results.entries()) {
    console.log(
      `    #${i + 1} ${r.success ? 'OK ' : 'LOI'} | clientRequestId: ${r.clientRequestId} | ` +
      `orderId: ${r.orderId || '-'} | status: ${r.status}` +
      (r.success ? '' : ` | ${r.errorCode} ${r.errorMessage}`),
    );
  }
}

const auth = new Auth(config);
await ensureAuth(auth, OTP);
const data = new Data(auth);
const trading = new Trading(auth);
console.log(`Tai khoan: ${maskAccount(ACCOUNT_NO)}`);

// --- Buoc 1: Gia san cua SSI hom nay ---
console.log(`\n--- 1. Gia san ${SYMBOL} (master data) ---`);
const masterData = await data.marketData.getMasterData();
const info = masterData.find((m) => m.symbol === SYMBOL);
if (!info || !(info.floor > 0)) {
  console.log(`Khong co gia san cua ${SYMBOL} hom nay (ngoai gio / ngay nghi?) — dung lai, khong gui lenh.`);
  process.exit(0);
}
const floorPrice = String(info.floor); // gia truyen dang chuoi
console.log(`  ${SYMBOL}: san ${floorPrice} | tham chieu ${info.refPrice} | tran ${info.ceiling}`);

// --- Buoc 2: Gioi han lo ---
const maxBatch = auth.config.maxBatchOrders;
console.log(`\n--- 2. Gioi han lo: maxBatchOrders = ${maxBatch} ---`);
const tag = Date.now().toString(36); // clientRequestId <= 20 ky tu: 'b' + tag + '-' + stt
const makeItems = (count) =>
  Array.from({ length: count }, (_, i) => ({
    accountNo: ACCOUNT_NO,
    symbol: SYMBOL,
    side: OrderSide.BUY,
    quantity: QUANTITY,
    price: floorPrice,
    orderType: OrderType.LO,
    clientRequestId: `b${tag}-${i + 1}`,
  }));

try {
  await trading.trading.placeBatchOrders(makeItems(maxBatch + 1));
  console.log('  (bat thuong) lo vuot gioi han van duoc chap nhan — kiem tra getTodayOrders va huy tay');
  process.exit(1);
} catch (err) {
  if (!(err instanceof ValidationError)) throw err;
  console.log(`  Lo ${maxBatch + 1} lenh bi tu choi tai may (chua gui request): ${err.message}`);
}

// --- Buoc 3: Dat lo va doc ket qua tung lenh ---
console.log(`\n--- 3. Dat lo ${ORDER_COUNT} lenh MUA ${SYMBOL} x ${QUANTITY} @ ${floorPrice} (LO) ---`);
const items = makeItems(ORDER_COUNT);
let placed;
try {
  placed = await trading.trading.placeBatchOrders(items);
} catch (err) {
  if (err instanceof DuplicateRequestError) {
    console.log(`  Lo da duoc dung truoc do (409): ${err.message} — kiem tra getTodayOrders truoc khi lam gi them.`);
    process.exit(1);
  }
  throw err;
}
printResults('Ket qua dat lo', placed);

// --- Buoc 4: Huy lo cac lenh vua dat thanh cong ---
const accepted = placed.results.filter((r) => r.success && r.orderId);
if (accepted.length === 0) {
  console.log('\nKhong co lenh nao duoc nhan, khong can huy.');
} else {
  console.log(`\n--- 4. Huy lo ${accepted.length} lenh (theo orderId) ---`);
  const cancelled = await trading.trading.cancelBatchOrders(
    accepted.map((r) => ({ accountNo: ACCOUNT_NO, orderId: r.orderId })),
  );
  printResults('Ket qua huy lo', cancelled);
}

// --- Buoc 5 (tuy chon): Chong trung bang clientRequestId ---
if (process.env.BATCH_DUP_DEMO === '1') {
  console.log('\n--- 5. Gui lai CUNG clientRequestId (BATCH_DUP_DEMO=1) ---');
  try {
    const again = await trading.trading.placeBatchOrders(items);
    printResults('Ket qua gui lai (server co the tu choi tung lenh)', again);
    // Neu server van nhan, huy ngay de khong de lenh treo
    const extra = again.results.filter((r) => r.success && r.orderId);
    if (extra.length > 0) {
      await trading.trading.cancelBatchOrders(extra.map((r) => ({ accountNo: ACCOUNT_NO, orderId: r.orderId })));
      console.log(`  Da huy ${extra.length} lenh phat sinh.`);
    }
  } catch (err) {
    if (!(err instanceof DuplicateRequestError)) throw err;
    console.log(`  DuplicateRequestError (HTTP ${err.statusCode}, code ${err.code}): ${err.message}`);
  }
}

// --- Doi chieu so lenh ---
console.log('\n--- So lenh hom nay (cac lenh vua tao) ---');
const todays = await trading.portfolio.getTodayOrders(ACCOUNT_NO);
const ids = new Set(items.map((it) => it.clientRequestId));
for (const o of todays.filter((x) => ids.has(x.clientRequestId))) {
  console.log(`  OrderID: ${o.orderId} | ${o.symbol} ${o.side} ${o.quantity} @ ${o.price} | Status: ${o.status}`);
}
