/**
 * Sample 15 — Vi the (positions) va suc mua/ban (PPMMR) theo T0 / T1 / T2
 * =========================================================================
 * Chi DOC du lieu, khong gui lenh. Can OTP lan dau (token duoc cache) vi la API tai khoan.
 *
 * Luong:
 *   1. getEquityPositions: moi ma co so luong ban cho ve T1/T2 (t1SellQuantity, t2SellQuantity) VA
 *      so luong mua cho ve T1/T2 (t1BuyQuantity, t2BuyQuantity — moi trong 3.3.0)
 *   2. getEquityPpmmr: suc mua / ban theo ngay (buyT0/T1/T2, sellT0/T1/T2), ty le ky quy, suc mua
 *   3. getDerivativePositions / getDerivativePpmmr (neu co tai khoan phai sinh)
 *   4. getMaxBuySellAtMarketPrice: so luong mua / ban toi da
 *
 * Luu y SDK 3.3.0:
 *   - Moi ham tren tra ve `undefined` khi tai khoan khong co phan Equity / Derivative tuong ung.
 *   - Positions Node chi tra danh sach theo ma (khong co dong tong "summary" rieng, khong co truong T0).
 *   - PPMMR: purchasingPower thay cho purchasePower (da deprecated); cac so tien la `number`
 *     phuc vu hien thi — khong dung de tinh toan tien te chinh xac.
 */

import { Auth, Trading } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, DERIVATIVE_ACCOUNT_NO, OTP, maskAccount, assertConfigured } from './config.js';
import { ensureAuth } from './auth_helper.js';

assertConfigured({ trading: true });

const vnd = (x) => Number(x).toLocaleString('en-US', { maximumFractionDigits: 0 });
const hasDerivative = !DERIVATIVE_ACCOUNT_NO.startsWith('<');

const auth = new Auth(config);
await ensureAuth(auth, OTP);
const trading = new Trading(auth);

// --- Buoc 1: Vi the co phieu, kem T1/T2 mua & ban ---
console.log(`--- 1. Vi the co phieu (${maskAccount(ACCOUNT_NO)}) ---`);
const positions = await trading.portfolio.getEquityPositions(ACCOUNT_NO);
if (!positions || positions.length === 0) {
  console.log('  (chua co vi the co phieu)');
} else {
  for (const p of positions) {
    console.log(
      `  ${p.symbol.padEnd(6)} | SL: ${p.quantity} | Co the ban: ${p.sellableQuantity} | Gia von: ${vnd(p.costPrice)}\n` +
      `         Ban cho ve  T1: ${p.t1SellQuantity}  T2: ${p.t2SellQuantity}` +
      ` | Mua cho ve  T1: ${p.t1BuyQuantity}  T2: ${p.t2BuyQuantity}\n` +
      `         Dang mua: ${p.buyingQuantity} | Da mua: ${p.boughtQuantity} | ` +
      `Dang ban: ${p.sellingQuantity} | Da ban: ${p.soldQuantity} | ` +
      `Phong toa: ${p.blockQuantity} | Cam co: ${p.mortgageQuantity} | Han che: ${p.restrictedQuantity}`,
    );
  }
  const totalT1Buy = positions.reduce((s, p) => s + p.t1BuyQuantity, 0);
  const totalT2Buy = positions.reduce((s, p) => s + p.t2BuyQuantity, 0);
  console.log(`  Tong mua cho ve: T1 ${totalT1Buy} | T2 ${totalT2Buy}`);
}

// --- Buoc 2: PPMMR co phieu ---
console.log('\n--- 2. PPMMR co phieu ---');
const ppmmr = await trading.portfolio.getEquityPpmmr(ACCOUNT_NO);
if (!ppmmr) {
  console.log('  (tai khoan khong co phan Equity)');
} else {
  console.log(`  Suc mua (purchasingPower): ${vnd(ppmmr.purchasingPower)} | EE: ${vnd(ppmmr.ee)} | EE tien mat: ${vnd(ppmmr.eeCash)}`);
  console.log(`  Ty le ky quy: ${ppmmr.marginRatio} (SSI: ${ppmmr.marginRatioSsi})`);
  console.log(`  Tien mat: ${vnd(ppmmr.accountBalance)} | Co the rut: ${vnd(ppmmr.withdrawable)} | No: ${vnd(ppmmr.totalDebt)}`);
  console.log(`  Mua    T0: ${vnd(ppmmr.buyT0)}  T1: ${vnd(ppmmr.buyT1)}  T2: ${vnd(ppmmr.buyT2)}  | Chua khop: ${vnd(ppmmr.buyUnmatched)}`);
  console.log(`  Ban    T0: ${vnd(ppmmr.sellT0)}  T1: ${vnd(ppmmr.sellT1)}  T2: ${vnd(ppmmr.sellT2)}  | Chua khop: ${vnd(ppmmr.sellUnmatched)}`);
  console.log(`  Canh bao: call margin ${vnd(ppmmr.callMargin)} | force sell ${vnd(ppmmr.callForceSell)}`);
}

// --- Buoc 3: Phai sinh (neu cau hinh tai khoan phai sinh) ---
if (hasDerivative) {
  console.log(`\n--- 3. Phai sinh (${maskAccount(DERIVATIVE_ACCOUNT_NO)}) ---`);
  const dpos = await trading.portfolio.getDerivativePositions(DERIVATIVE_ACCOUNT_NO);
  if (!dpos) {
    console.log('  (tai khoan khong co phan Derivative)');
  } else {
    console.log(`  Vi the mo: ${dpos.openPositions.length} | Vi the da dong: ${dpos.closedPositions.length}`);
    for (const p of dpos.openPositions) {
      console.log(`  ${p.symbol} | Long: ${p.long} Short: ${p.short} Net: ${p.net} | Floating P/L: ${vnd(p.floatingPl)}`);
    }
  }
  const dppmmr = await trading.portfolio.getDerivativePpmmr(DERIVATIVE_ACCOUNT_NO);
  if (dppmmr) {
    console.log(
      `  Tien: ${vnd(dppmmr.accountBalance)} | Co the rut: ${vnd(dppmmr.withdrawable)} | ` +
      `Tong P/L: ${vnd(dppmmr.totalPl)} | Tong tai san: ${vnd(dppmmr.totalAssetSsi)}`,
    );
  }
} else {
  console.log('\n--- 3. Phai sinh: bo qua (chua cau hinh derivative_account) ---');
}

// --- Buoc 4: Mua / ban toi da tai gia thi truong ---
console.log('\n--- 4. Mua/ban toi da SSI tai gia thi truong ---');
const maxBs = await trading.trading.getMaxBuySellAtMarketPrice(ACCOUNT_NO, 'SSI');
console.log(
  `  Mua toi da: ${maxBs.maxBuyQuantity} | Ban toi da: ${maxBs.maxSellQuantity} | ` +
  `Suc mua: ${maxBs.purchasingPower ?? 'N/A'} | Ty le ky quy: ${maxBs.marginRatio ?? 'N/A'}`,
);
