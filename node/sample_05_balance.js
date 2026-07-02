/**
 * Sample 5 — Lay so du tai khoan (Account Balance)
 * ===================================================
 * Kiem tra kha nang giao dich truoc khi dat lenh.
 *
 * Luong:
 *   1. Client goi endpoint balance theo accountNo
 *   2. API tra ve available, onHold, limits, settlement
 *   3. Tinh kha nang mua/ban thuc te theo nghiep vu
 *   4. Neu khong du dieu kien thi chan thao tac dat lenh
 *   5. Neu du dieu kien thi cho phep di tiep sang order flow
 */

import { Auth, Trading } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

const vnd = (x) => Number(x).toLocaleString('en-US', { maximumFractionDigits: 0 });

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const trading = new Trading(auth);

// --- Buoc 1: Lay danh sach tai khoan ---
const accounts = await trading.account.getAccountInfo();
console.log('Danh sach tai khoan:');
for (const acc of accounts) {
  console.log(`  - ${acc.accountNo} (${acc.accountType})`);
}

// --- Buoc 2: Lay so du tai khoan Equity ---
console.log(`\n--- So du tai khoan Equity: ${ACCOUNT_NO} ---`);
const balance = await trading.portfolio.getEquityBalance(ACCOUNT_NO);
console.log(`  Tien mat kha dung : ${vnd(balance.availableCash).padStart(15)}`);
console.log(`  Tong no           : ${vnd(balance.totalDebt).padStart(15)}`);
console.log(`  Mua T0/T1/T2      : ${vnd(balance.buyT0).padStart(12)} / ${vnd(balance.buyT1).padStart(12)} / ${vnd(balance.buyT2).padStart(12)}`);
console.log(`  Ban T0/T1/T2      : ${vnd(balance.sellT0).padStart(12)} / ${vnd(balance.sellT1).padStart(12)} / ${vnd(balance.sellT2).padStart(12)}`);

// --- Buoc 3: Kiem tra suc mua toi da cho mot ma ---
console.log('\n--- Suc mua/ban toi da: SSI ---');
const maxBs = await trading.trading.getMaxBuySell(ACCOUNT_NO, 'SSI', 26000);
console.log(`  Max mua : ${String(maxBs.maxBuyQuantity).padStart(10)} co phieu`);
console.log(`  Max ban : ${String(maxBs.maxSellQuantity).padStart(10)} co phieu`);
console.log(`  Suc mua : ${maxBs.purchasePower}`);

// --- Buoc 4: Logic kiem tra truoc khi dat lenh ---
const desiredQuantity = 100;
const desiredPrice = 26000;
const requiredAmount = desiredQuantity * desiredPrice;

if (balance.availableCash >= requiredAmount) {
  console.log(`\n[OK] Du dieu kien: can ${vnd(requiredAmount)}, co ${vnd(balance.availableCash)}`);
  console.log('  -> Cho phep dat lenh mua.');
} else {
  console.log(`\n[X] Khong du: can ${vnd(requiredAmount)}, chi co ${vnd(balance.availableCash)}`);
  console.log('  -> Chan dat lenh.');
}

// --- Buoc 5: Xem vi the hien co ---
console.log(`\n--- Vi the co phieu (${ACCOUNT_NO}) ---`);
const positions = await trading.portfolio.getEquityPositions(ACCOUNT_NO);
for (const pos of positions) {
  console.log(
    `  ${pos.symbol.padEnd(10)} | SL: ${String(pos.quantity).padStart(8)} | ` +
    `Ban duoc: ${String(pos.sellableQuantity).padStart(8)} | Gia von: ${vnd(pos.costPrice).padStart(10)}`,
  );
}

// --- Response Summary ---
console.log(`\n[Response] accounts|avail_cash|max_buy_qty|max_sell_qty|positions`);
console.log(`${accounts.length}|${balance.availableCash}|${maxBs.maxBuyQuantity}|${maxBs.maxSellQuantity}|${positions.length}`);
if (positions.length > 0) {
  const p = positions[0];
  console.log(`[Response:first_pos] symbol|quantity|sellable|cost_price`);
  console.log(`${p.symbol}|${p.quantity}|${p.sellableQuantity}|${p.costPrice}`);
}
