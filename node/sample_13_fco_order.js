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
 *   8. Truy vấn danh sách, lịch sử trạng thái & Hủy lệnh FCO
 *
 * CẢNH BÁO: sample này GỬI LỆNH ĐIỀU KIỆN THẬT lên tài khoản trong config.json. Chỉ chạy trên tài
 * khoản thử nghiệm hoặc khi bạn hiểu rõ rủi ro; sau khi chạy hãy hủy các lệnh còn treo.
 *
 * Lưu ý SDK 3.3.0:
 *   - Ngày FCO có dạng `yyyy/MM/dd HH:mm:ss`, khoảng [fromDate, toDate] tối đa 31 ngày.
 *   - Tài khoản yêu cầu mã OTP khi đặt/hủy FCO: đặt biến môi trường FCO_OTP (mã dùng một lần,
 *     không ghi vào file); SDK chỉ gửi trường `code` khi có giá trị.
 *   - getFcoStatusHistory(fcoId) trả về danh sách {state, time, code, detail} (rỗng nếu chưa có).
 */

import { Auth, Trading, OrderSide, FCOOperator } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const trading = new Trading(auth);

const symbol = 'SSI';

// Khoảng hiệu lực: từ bây giờ đến 7 ngày sau (≤ 31 ngày), định dạng yyyy/MM/dd HH:mm:ss
const pad = (n) => String(n).padStart(2, '0');
const fmt = (d) =>
  `${d.getFullYear()}/${pad(d.getMonth() + 1)}/${pad(d.getDate())} ` +
  `${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
const now = new Date();
const fromDate = fmt(now);
const toDate = fmt(new Date(now.getTime() + 7 * 24 * 3600 * 1000));

// Mã OTP một lần cho FCO (nếu tài khoản yêu cầu) — lấy từ môi trường, không hardcode
const FCO_CODE = process.env.FCO_OTP || undefined;

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
    FCO_CODE,
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
    FCO_CODE,
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
    FCO_CODE,
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
    FCO_CODE,
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
    FCO_CODE,
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
    FCO_CODE,
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
    FCO_CODE,
  );
  console.log('  Bull Bear Result:', bbRes);

  // --- 8. Truy vấn danh sách lệnh FCO ---
  console.log('\n--- 8. Danh sách lệnh FCO ---');
  const fcoList = await trading.trading.getFcoByAccountNo(ACCOUNT_NO, 1, 10);
  console.log(`  Tổng số lệnh FCO: ${fcoList.itemsCount}`);
  for (const item of fcoList.fcoList.slice(0, 5)) {
    console.log(`  FCO ID: ${item.fcoId} | Mã: ${item.symbol} | Loại: ${item.type} | Trạng thái: ${item.status}`);
  }

  // --- 9. Lịch sử trạng thái của lệnh GTD vừa tạo ---
  if (gtdRes.fcoId) {
    console.log(`\n--- 9. Lịch sử trạng thái FCO ID: ${gtdRes.fcoId} ---`);
    const history = await trading.trading.getFcoStatusHistory(gtdRes.fcoId);
    if (history.length === 0) console.log('  (chưa có lịch sử)');
    for (const h of history) {
      console.log(`  ${h.time} | ${h.state}${h.code ? ` (${h.code})` : ''} | ${h.detail}`);
    }
  }

  // --- 10. Hủy lệnh FCO vừa tạo nếu có ---
  if (gtdRes.fcoId) {
    console.log(`\n--- 10. Hủy lệnh FCO ID: ${gtdRes.fcoId} ---`);
    const cancelRes = await trading.trading.cancelFco(gtdRes.fcoId, FCO_CODE);
    console.log('  Hủy FCO Result:', cancelRes);
  }

  console.log('\n[Response] sample_13_fco_completed');
} catch (error) {
  // Chỉ in tên/mã/thông điệp lỗi — không in nguyên object lỗi (có thể chứa header/body request)
  console.error(`Lỗi khi thực thi Sample 13: ${error.name} [${error.code ?? 'n/a'}] ${error.message}`);
}
