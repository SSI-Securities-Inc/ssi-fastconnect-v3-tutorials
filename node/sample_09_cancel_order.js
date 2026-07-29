/**
 * Sample 9 — Huy lenh
 * =====================
 * Dung phan khoi luong chua khop cua lenh dang mo.
 *
 * Luong:
 *   1. Lay so lenh, tim lenh dang PENDING/PARTIAL_FILLED
 *   2. Goi cancelOrderById kem accountNo/orderId
 *   3. API xac thuc quyen va trang thai lenh hien tai
 *   4. Neu hop le, he thong cap nhat CANCELLED cho phan chua khop
 *   5. Dong bo lai so lenh va so luong con treo
 */

import { Auth, Trading, OrderStatus } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

const vnd = (x) => Number(x).toLocaleString('en-US', { maximumFractionDigits: 0 });

// Trang thai co the huy
const CANCELLABLE_STATUSES = new Set([
  OrderStatus.PENDING_APPROVAL,
  OrderStatus.READY,
  OrderStatus.SENT,
  OrderStatus.QUEUED,
  OrderStatus.PARTIAL_FILLED,
]);

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const trading = new Trading(auth);

// --- Buoc 1: Lay so lenh, tim lenh dang mo ---
console.log('--- So lenh hom nay ---');
const orders = await trading.portfolio.getTodayOrders(ACCOUNT_NO);
const openOrders = orders.filter((o) => CANCELLABLE_STATUSES.has(o.status));
console.log(`Tong lenh: ${orders.length} | Lenh dang mo: ${openOrders.length}\n`);

if (openOrders.length === 0) {
  console.log('Khong co lenh nao dang mo de huy.');

  // --- Response Summary ---
  console.log('\n[Response] open_count|cancel_status');
  console.log('0|N/A');
} else {
  for (const order of openOrders) {
    const remaining = order.quantity - order.filledQuantity;
    console.log(
      `  OrderID: ${order.orderId} | ${order.symbol} ${order.side} ${order.orderType} | ` +
      `SL: ${order.quantity} @ ${order.price} | Khop: ${order.filledQuantity} | ` +
      `Con: ${remaining} | Status: ${order.status}`,
    );
  }

  // --- Buoc 2: Huy lenh dau tien trong danh sach ---
  const target = openOrders[0];
  console.log(`\n--- Huy lenh: ${target.orderId} ---`);
  const result = await trading.trading.cancelOrderById(ACCOUNT_NO, target.orderId);
  console.log('  Ket qua huy:', result);

  // --- Buoc 3: Xac nhan trang thai sau huy ---
  console.log('\n--- Kiem tra so lenh sau huy ---');
  const ordersAfter = await trading.portfolio.getTodayOrders(ACCOUNT_NO);
  for (const order of ordersAfter) {
    if (order.orderId === target.orderId) {
      console.log(
        `  OrderID: ${order.orderId} | Status: ${order.status} | ` +
        `Khop: ${order.filledQuantity} | Da huy: ${order.cancelQuantity}`,
      );
      break;
    }
  }

  // --- Response Summary ---
  console.log('\n[Response] open_count|cancel_status');
  console.log(`${openOrders.length}|${result.status}`);
}

// --- Buoc 4: Cap nhat lai so du ---
console.log('\n--- So du sau huy ---');
const balance = await trading.portfolio.getEquityBalance(ACCOUNT_NO);
console.log(`  Tien mat kha dung: ${vnd(balance.accountBalance).padStart(15)}`);
