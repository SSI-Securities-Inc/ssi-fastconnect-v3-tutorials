/**
 * Sample 7 — Dat lenh Market (MTL)
 * ===================================
 * Khop lenh nhanh theo gia thi truong hien tai.
 *
 * Luong:
 *   1. Client tao order MARKET (khong gui price)
 *   2. Gui request toi Trading Orders API
 *   3. He thong match theo thanh khoan thi truong tai thoi diem gui
 *   4. API tra ve trang thai khop (FILLED hoac PARTIAL_FILLED)
 *   5. Cap nhat ngay danh muc/so du tam tinh
 */

import { Auth, Trading, OrderSide } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

const vnd = (x) => Number(x).toLocaleString('en-US', { maximumFractionDigits: 0 });

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const trading = new Trading(auth);

// --- Buoc 1: Kiem tra suc mua/ban o gia thi truong ---
const maxBs = await trading.trading.getMaxBuySellAtMarketPrice(ACCOUNT_NO, 'SSI');
console.log(`Max mua (market): ${maxBs.maxBuyQuantity} co phieu`);
console.log(`Max ban (market): ${maxBs.maxSellQuantity} co phieu`);

// --- Buoc 2: Dat lenh Market mua ---
console.log('\n--- Dat lenh MARKET mua SSI ---');
const result = await trading.trading.placeMarketOrder(ACCOUNT_NO, 'SSI', OrderSide.BUY, 100);
console.log('  Ket qua:', result);

// --- Buoc 3: Kiem tra trang thai lenh ---
console.log('\n--- So lenh hom nay ---');
const orders = await trading.portfolio.getTodayOrders(ACCOUNT_NO);
for (const order of orders.slice(-3)) {
  console.log(
    `  ${order.orderId} | ${order.symbol} ${order.side} ${order.orderType} | ` +
    `SL: ${order.quantity} | Khop: ${order.filledQuantity} | Trang thai: ${order.status}`,
  );
}

// --- Buoc 4: Cap nhat lai so du sau khi khop ---
console.log('\n--- So du sau giao dich ---');
const balance = await trading.portfolio.getEquityBalance(ACCOUNT_NO);
console.log(`  Tien mat kha dung: ${vnd(balance.availableCash).padStart(15)}`);

// --- Buoc 5: Cap nhat danh muc ---
console.log('\n--- Vi the sau giao dich ---');
const positions = await trading.portfolio.getEquityPositions(ACCOUNT_NO);
for (const pos of positions) {
  if (pos.symbol === 'SSI') {
    console.log(`  SSI | SL: ${pos.quantity} | Gia von: ${vnd(pos.costPrice)}`);
  }
}

// --- Response Summary ---
console.log(`\n[Response] max_buy_mkt|max_sell_mkt|buy_status`);
console.log(`${maxBs.maxBuyQuantity}|${maxBs.maxSellQuantity}|${result.status}`);
