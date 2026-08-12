/**
 * Sample 6 — Dat lenh Limit (LO)
 * =================================
 * Dat lenh mua/ban tai muc gia chi dinh.
 *
 * Luong:
 *   1. Client tao payload order (symbol, side, quantity, price)
 *   2. SDK tu gan clientRequestId (chong submit trung) va ky RSA
 *   3. Gui request toi Trading Orders API
 *   4. API tra ve orderId va trang thai ban dau (PENDING)
 *   5. Luu orderId de theo doi khop lenh
 */

import { Auth, Trading, OrderSide } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const trading = new Trading(auth);

// --- Buoc 1: Kiem tra suc mua truoc ---
const maxBs = await trading.trading.getMaxBuySell(ACCOUNT_NO, 'SSI', 26000);
console.log(`Suc mua toi da SSI @ 26,000: ${maxBs.maxBuyQuantity} co phieu`);

if (maxBs.maxBuyQuantity < 100) {
  console.log('Khong du suc mua, dung lai.');
} else {
  // --- Buoc 2: Dat lenh Limit mua ---
  console.log('\n--- Dat lenh LIMIT mua SSI ---');
  const buyResult = await trading.trading.placeLimitOrder(ACCOUNT_NO, 'SSI', OrderSide.BUY, 100, 26000);
  console.log('  Ket qua:', buyResult);

  // --- Buoc 3: Dat lenh Limit ban ---
  console.log('\n--- Dat lenh LIMIT ban SSI ---');
  const sellResult = await trading.trading.placeLimitOrder(ACCOUNT_NO, 'SSI', OrderSide.SELL, 100, 27000);
  console.log('  Ket qua:', sellResult);

  // --- Buoc 4: Kiem tra lenh vua dat trong so lenh ---
  console.log('\n--- So lenh hom nay ---');
  const orders = await trading.portfolio.getTodayOrders(ACCOUNT_NO);
  for (const order of orders.slice(-5)) {
    console.log(
      `  ${order.orderId} | ${order.symbol} ${order.side} ${order.orderType} | ` +
      `SL: ${order.quantity} @ ${order.price} | Trang thai: ${order.status}`,
    );
  }

  // --- Response Summary ---
  console.log(`\n[Response] max_buy_qty|buy_status|sell_status`);
  console.log(`${maxBs.maxBuyQuantity}|${buyResult.status}|${sellResult.status}`);
}
