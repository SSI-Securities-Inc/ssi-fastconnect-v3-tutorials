/**
 * Sample 13 — Đặt Lệnh Điều Kiện (FCO)
 * =====================================
 * Thể hiện đầy đủ các loại lệnh điều kiện (Fast Conditional Orders - FCO):
 *   1. GTD (Good-Till-Date / Lệnh chờ theo ngày)
 *   2. Stop (Lệnh dừng giá thị trường)
 *   3. Stop Limit (Lệnh dừng giá giới hạn)
 *   4. Trailing Stop (Lệnh dừng xu hướng)
 *   5. Trailing Stop Limit (Lệnh dừng xu hướng giới hạn)
 *   6. OCO (One-Cancels-the-Other / Lệnh Chốt lời & Cắt lỗ)
 *   7. Bull Bear (Lệnh Hai đầu)
 *   8. Truy vấn danh sách & Hủy lệnh FCO
 */

import { Auth, Trading, OrderSide, FCOOperator } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const trading = new Trading(auth);

const symbol = 'SSI';
const fromDate = '2026/08/01 00:00:00';
const toDate = '2026/08/30 23:59:59';

console.log('=== FASTCONNECT NODE.JS SDK — SAMPLE 13: LỆNH ĐIỀU KIỆN (FCO) ===\n');

try {
  // --- 1. Lệnh GTD (Good-Till-Date) ---
  console.log('--- 1. Đặt lệnh GTD ---');
  const gtdRes = await trading.trading.placeFcoGtd(
    ACCOUNT_NO,
    symbol,
    OrderSide.BUY,
    100,
    26000,
    0,
    fromDate,
    toDate,
  );
  console.log('  GTD Result:', gtdRes);

  // --- 2. Lệnh Stop (Stop Market) ---
  console.log('\n--- 2. Đặt lệnh Stop ---');
  const stopRes = await trading.trading.placeFcoStop(
    ACCOUNT_NO,
    symbol,
    OrderSide.BUY,
    100,
    27000,
    FCOOperator.GREATER_OR_EQUAL,
    fromDate,
    toDate,
  );
  console.log('  Stop Result:', stopRes);

  // --- 3. Lệnh Stop Limit ---
  console.log('\n--- 3. Đặt lệnh Stop Limit ---');
  const stopLimitRes = await trading.trading.placeFcoStopLimit(
    ACCOUNT_NO,
    symbol,
    OrderSide.BUY,
    100,
    27500,
    0,
    27000,
    FCOOperator.GREATER_OR_EQUAL,
    fromDate,
    toDate,
  );
  console.log('  Stop Limit Result:', stopLimitRes);

  // --- 4. Lệnh Trailing Stop ---
  console.log('\n--- 4. Đặt lệnh Trailing Stop ---');
  const trailingRes = await trading.trading.placeFcoTrailingStop(
    ACCOUNT_NO,
    symbol,
    OrderSide.SELL,
    100,
    28000,
    1000,
    fromDate,
    toDate,
  );
  console.log('  Trailing Stop Result:', trailingRes);

  // --- 5. Lệnh Trailing Stop Limit ---
  console.log('\n--- 5. Đặt lệnh Trailing Stop Limit ---');
  const trailingLimitRes = await trading.trading.placeFcoTrailingStopLimit(
    ACCOUNT_NO,
    symbol,
    OrderSide.SELL,
    100,
    28000,
    1000,
    500,
    fromDate,
    toDate,
  );
  console.log('  Trailing Stop Limit Result:', trailingLimitRes);

  // --- 6. Lệnh OCO (One-Cancels-the-Other) ---
  console.log('\n--- 6. Đặt lệnh OCO ---');
  const ocoRes = await trading.trading.placeFcoOco(
    ACCOUNT_NO,
    symbol,
    OrderSide.SELL,
    100,
    30000, // tpActivePrice
    24000, // slActivePrice
    30000, // tpPrice
    24000, // slPrice
    0,     // tpSlip
    0,     // slSlip
    fromDate,
    toDate,
  );
  console.log('  OCO Result:', ocoRes);

  // --- 7. Lệnh Bull Bear ---
  console.log('\n--- 7. Đặt lệnh Bull Bear ---');
  const bbRes = await trading.trading.placeFcoBullBear(
    ACCOUNT_NO,
    symbol,
    OrderSide.BUY,
    100,
    26000, // price
    0,     // priceSlip
    30000, // tpActivePrice
    24000, // slActivePrice
    30000, // tpPrice
    24000, // slPrice
    0,     // tpSlip
    0,     // slSlip
    fromDate,
    toDate,
  );
  console.log('  Bull Bear Result:', bbRes);

  // --- 8. Truy vấn danh sách lệnh FCO ---
  console.log('\n--- 8. Danh sách lệnh FCO ---');
  const fcoList = await trading.trading.getFcoByAccountNo(ACCOUNT_NO, 1, 10);
  console.log(`  Tổng số lệnh FCO: ${fcoList.itemsCount}`);
  for (const item of fcoList.fcoList.slice(0, 5)) {
    console.log(`  FCO ID: ${item.fcoId} | Mã: ${item.symbol} | Loại: ${item.type} | Trạng thái: ${item.status}`);
  }

  // --- 9. Hủy lệnh FCO vừa tạo nếu có ---
  if (gtdRes.fcoId) {
    console.log(`\n--- 9. Hủy lệnh FCO ID: ${gtdRes.fcoId} ---`);
    const cancelRes = await trading.trading.cancelFco(gtdRes.fcoId);
    console.log('  Hủy FCO Result:', cancelRes);
  }

  console.log('\n[Response] sample_13_fco_completed');
} catch (error) {
  console.error('Lỗi khi thực thi Sample 13:', error);
}
