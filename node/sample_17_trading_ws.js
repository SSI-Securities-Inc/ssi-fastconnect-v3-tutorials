/**
 * Sample 17 — Trading WebSocket (THU NGHIEM / EXPERIMENTAL)
 * ===========================================================
 * Dat lenh, huy lenh va truy van tai khoan qua WebSocket `/ws/v3/trading` thay vi REST.
 *
 * !!! EXPERIMENTAL !!!  Phia server CHUA duoc cong bo la da phat hanh. API nay nam o
 * `@ssi.developer/ssi-sdk/experimental`, khong thuoc facade on dinh (Trading / Stream khong bao gio
 * ket noi toi no) va co the doi/bi go ma khong bao truoc. Chi chay tren moi truong ban duoc cho biet
 * co ho tro. Vi vay sample mac dinh THOAT NGAY, phai bat tuong minh:
 *
 *   TRADING_WS_EXPERIMENTAL=1   bat sample (chi ket noi, subscribe, truy van — khong gui lenh)
 *   TRADING_WS_PLACE=1          them: dat 1 lenh MUA LO tai GIA SAN roi huy ngay (LENH THAT)
 *
 * Luong:
 *   1. Ket noi tradingWs(auth) (can token trading -> OTP lan dau)
 *   2. onEvent: nhan su kien khong gan voi lenh nao (order / portfolio / pong)
 *   3. subscribeOrderEvents / subscribePortfolioEvents theo MOT tai khoan (khong wildcard / danh sach)
 *   4. Truy van: queryPosition, queryAccountBalance, queryMaxBuySell, queryPpmmrAccount, queryOrderBook
 *   5. (TRADING_WS_PLACE=1) placeOrder -> ack PD; ket qua that den qua su kien order; cancelOrder
 *   6. Rate limit: doc ws.client.rateLimit sau moi lenh; SDK tu cho khi `remaining` = 0
 *      (toi da maxRateLimitWaitMs); loi tu server la TradingWSError (status / errorCode / serverMessage)
 *   7. unsubscribeAllEvents, close
 *
 * Luu y: lenh doi trang thai duoc ky bang private key nhu REST; khong bao gio retry khi timeout —
 * neu mat ack, kiem tra su kien order / queryOrderBook truoc khi gui lai.
 */

import { Auth, Data, OrderSide, OrderType } from '@ssi.developer/ssi-sdk';
import { tradingWs, TradingWSError } from '@ssi.developer/ssi-sdk/experimental';
import { config, ACCOUNT_NO, OTP, maskAccount, assertConfigured } from './config.js';
import { ensureAuth } from './auth_helper.js';

if (process.env.TRADING_WS_EXPERIMENTAL !== '1') {
  console.log('Sample 17 la EXPERIMENTAL (server chua cong bo phat hanh trading WebSocket).');
  console.log('Dat TRADING_WS_EXPERIMENTAL=1 de chay. Dat them TRADING_WS_PLACE=1 de gui lenh that.');
  process.exit(0);
}

assertConfigured({ trading: true });

const SYMBOL = 'SSI';
const QUANTITY = 100;

// In ket qua gon, che so tai khoan; cat bot de khong tran man hinh
const show = (value, max = 500) => {
  const text = JSON.stringify(value) ?? 'undefined';
  const masked = text.split(ACCOUNT_NO).join(maskAccount(ACCOUNT_NO));
  return masked.length > max ? `${masked.slice(0, max)}...` : masked;
};
const showRate = (ws) => `rateLimit: ${show(ws.client.rateLimit, 200)}`;

function describeWsError(err) {
  if (err instanceof TradingWSError) {
    return `TradingWSError status=${err.status} code=${err.errorCode} msg=${err.serverMessage}`;
  }
  return `${err.name}: ${err.message}`;
}

const auth = new Auth(config);
await ensureAuth(auth, OTP);

console.log('Dang ket noi trading WebSocket (EXPERIMENTAL)...');
let ws;
try {
  ws = await tradingWs(auth); // doc Config.tradingWsUrl; timeout yeu cau mac dinh = Config.timeoutMs
} catch (err) {
  console.log(`Khong ket noi duoc: ${describeWsError(err)}`);
  console.log('Co the moi truong chua ho tro trading WebSocket — dung bang REST (sample_06...).');
  process.exit(1);
}
console.log(`Da ket noi. Tai khoan: ${maskAccount(ACCOUNT_NO)}`);

// --- Buoc 2: Su kien khong gan voi lenh nao ---
ws.client.onEvent = (frame) => {
  const kind = frame.topic ?? frame.method ?? frame.event ?? 'frame';
  console.log(`  [EVENT] ${kind} | ${show(frame, 300)}`);
};

try {
  // --- Buoc 3: Subscribe su kien theo tai khoan ---
  await ws.subscribeOrderEvents(ACCOUNT_NO);
  await ws.subscribePortfolioEvents(ACCOUNT_NO);
  console.log('Da subscribe order + portfolio events.');

  // --- Buoc 4: Truy van ---
  console.log('\n--- query.position ---');
  const pos = await ws.queryPosition(ACCOUNT_NO, true);
  console.log(`  status ${pos.status} | ${show(pos.result)} | ${showRate(ws)}`);

  console.log('\n--- query.accountBalance ---');
  const bal = await ws.queryAccountBalance(ACCOUNT_NO);
  console.log(`  status ${bal.status} | ${show(bal.result)}`);

  console.log('\n--- query.maxBuySell (gia tham chieu cua server) ---');
  const max = await ws.queryMaxBuySell(ACCOUNT_NO, SYMBOL);
  console.log(`  status ${max.status} | ${show(max.result)}`);

  console.log('\n--- query.ppmmrAccount ---');
  const ppmmr = await ws.queryPpmmrAccount(ACCOUNT_NO);
  console.log(`  status ${ppmmr.status} | ${show(ppmmr.result, 300)}`);

  console.log('\n--- query.orderBook (so lenh, trang 1) ---');
  const book = await ws.queryOrderBook(ACCOUNT_NO, { pageIndex: 1, pageSize: 10 });
  console.log(`  status ${book.status} | ${show(book.result, 400)}`);

  // --- Buoc 5: Dat / huy lenh (chi khi TRADING_WS_PLACE=1) ---
  if (process.env.TRADING_WS_PLACE === '1') {
    const data = new Data(auth);
    const info = (await data.marketData.getMasterData()).find((m) => m.symbol === SYMBOL);
    if (!info || !(info.floor > 0)) {
      console.log(`\nKhong co gia san cua ${SYMBOL} — bo qua buoc dat lenh.`);
    } else {
      const floorPrice = String(info.floor);
      const clientRequestId = `w${Date.now().toString(36)}`; // <= 20 ky tu
      console.log(`\n--- order.place: MUA ${SYMBOL} x ${QUANTITY} @ ${floorPrice} (LO) ---`);
      const ack = await ws.placeOrder(ACCOUNT_NO, SYMBOL, OrderSide.BUY, QUANTITY, floorPrice, OrderType.LO, clientRequestId);
      console.log(`  Ack: orderId ${ack.orderId || '-'} | status ${ack.status} | ${showRate(ws)}`);
      console.log('  (ket qua that cua lenh den qua su kien [EVENT] phia tren)');

      if (ack.orderId) {
        console.log('\n--- order.cancel ---');
        const cancelled = await ws.cancelOrder(ACCOUNT_NO, ack.orderId);
        console.log(`  Huy: status ${cancelled.status} | clientCancelId ${cancelled.clientCancelId || '-'}`);
      }
    }
  } else {
    console.log('\n(Bo qua dat/huy lenh. Dat TRADING_WS_PLACE=1 de gui 1 lenh LO tai gia san roi huy.)');
  }

  // Cho vai giay de nhan su kien con lai
  await new Promise((resolve) => setTimeout(resolve, 3000));
} catch (err) {
  // 429 / loi tu server la TradingWSError; mat ack (timeout) la WebSocketError — KHONG tu dong gui lai
  console.log(`\nLoi: ${describeWsError(err)}`);
  if (err instanceof TradingWSError && err.status === 429) {
    console.log('  Dang bi gioi han toc do — cho roi thu lai (xem sample_18).');
  }
  process.exitCode = 1;
} finally {
  await ws.unsubscribeAllEvents().catch(() => undefined);
  await ws.close();
  console.log('\nDa dong trading WebSocket.');
}
