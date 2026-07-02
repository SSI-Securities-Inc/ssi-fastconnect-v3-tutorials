/**
 * Sample 8 — Kiem tra trang thai lenh
 * ======================================
 * Theo doi tien trinh khop cua mot lenh cu the.
 *
 * Luong:
 *   1. Dat mot lenh de theo doi
 *   2. Goi getTodayOrders, doi chieu status / filledQuantity
 *   3. Tinh luong con lai va trang thai hien tai
 *   4. Neu chua hoan tat thi tiep tuc polling chu ky ngan
 *   5. Khi FILLED/CANCELLED/REJECTED thi dong vong theo doi
 */

import { Auth, Trading, OrderSide, OrderStatus } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

const sleep = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
const vnd = (x) => Number(x).toLocaleString('en-US', { maximumFractionDigits: 0 });

// Trang thai ket thuc — khong can polling them
const TERMINAL_STATUSES = new Set([
  OrderStatus.FILLED,
  OrderStatus.CANCELLED,
  OrderStatus.REJECTED,
  OrderStatus.EXPIRED,
  OrderStatus.PARTIAL_CANCELLED,
]);

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const trading = new Trading(auth);

// --- Buoc 1: Dat mot lenh de theo doi ---
console.log('Dat lenh Limit mua SSI @ 26000...');
const result = await trading.trading.placeLimitOrder(ACCOUNT_NO, 'SSI', OrderSide.BUY, 100, 26000);
console.log('  Ket qua dat lenh:', result);

// --- Buoc 2-5: Polling trang thai ---
console.log('\n--- Bat dau theo doi trang thai lenh ---');
const maxPolls = 10;
const pollInterval = 3000; // ms
let finished = false;
let lastOrder = null;

for (let i = 1; i <= maxPolls; i += 1) {
  const orders = await trading.portfolio.getTodayOrders(ACCOUNT_NO);

  if (orders.length === 0) {
    console.log(`  Poll ${i}: Chua co lenh trong so.`);
    await sleep(pollInterval);
    continue;
  }

  const latest = orders[orders.length - 1];
  lastOrder = latest;
  const remaining = latest.quantity - latest.filledQuantity - latest.cancelQuantity;

  console.log(
    `  Poll ${i}: OrderID=${latest.orderId} | Status=${latest.status} | ` +
    `Khop=${latest.filledQuantity}/${latest.quantity} | Con lai=${remaining}`,
  );

  if (TERMINAL_STATUSES.has(latest.status)) {
    console.log(`\n-> Lenh da ket thuc voi trang thai: ${latest.status}`);
    if (latest.filledQuantity > 0) {
      console.log(`  Da khop: ${latest.filledQuantity} co phieu @ trung binh ${vnd(latest.avgPrice)}`);
    }
    finished = true;
    break;
  }

  await sleep(pollInterval);
}

if (!finished) {
  console.log(`\nHet ${maxPolls} lan poll, lenh van dang mo.`);
}

// --- Response Summary ---
if (lastOrder !== null) {
  console.log(`\n[Response] place_status|final_status|filled_qty|total_qty`);
  console.log(`${result.status}|${lastOrder.status}|${lastOrder.filledQuantity}|${lastOrder.quantity}`);
}
