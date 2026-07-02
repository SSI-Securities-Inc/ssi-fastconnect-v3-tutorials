/**
 * Sample 11 — WebSocket trading real-time (trang thai lenh & danh muc)
 * ======================================================================
 * Nhan cap nhat tuc thoi ve lenh khop va danh muc tai khoan.
 *
 * Luong:
 *   1. Client mo ket noi WebSocket bang token hop le (can OTP)
 *   2. Subscribe stream order status va portfolio theo accountNo
 *   3. Server push event khi trang thai lenh hoac danh muc thay doi
 *   4. Parse message theo loai (orderEvent / clientPortfolioEvent)
 *   5. Ping dinh ky giu ket noi; Ctrl+C de dung
 */

import { Auth, Stream } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

// --- Callback xu ly su kien trading ---
function onTradingEvent(msg) {
  if (msg.type === 'orderEvent') {
    console.log(
      `  [ORDER] ${msg.symbol} ${msg.side} | OrderID: ${msg.orderId} | ` +
      `Status: ${msg.status} | Khop: ${msg.filledQuantity}/${msg.quantity}`,
    );
  } else if (msg.type === 'clientPortfolioEvent') {
    console.log(
      `  [PORTFOLIO] Account: ${msg.accountNo} | ` +
      `Tong TS: ${msg.totalAsset} | Cash: ${msg.cashBalance}`,
    );
  } else {
    console.log(`  [TRADING] ${JSON.stringify(msg)}`);
  }
}

function onHeartbeat(msg) {
  console.log(`  [HEARTBEAT] ${msg.status} - ${msg.message}`);
}

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const stream = new Stream(auth);

// --- Buoc 1: Dang ky callback ---
stream.streaming.onTrading = onTradingEvent;
stream.streaming.onHeartbeat = onHeartbeat;

// --- Buoc 2: Mo ket noi WebSocket ---
console.log('Dang ket noi WebSocket...');
await stream.streaming.connect();
console.log('Da ket noi!\n');

// --- Buoc 3: Subscribe trang thai lenh real-time ---
console.log('Subscribing trang thai lenh...');
stream.streaming.subscribeOrderStatus(ACCOUNT_NO);

// --- Buoc 4: Subscribe danh muc tai khoan real-time (mo comment de bat) ---
// console.log('Subscribing danh muc tai khoan...');
// stream.streaming.subscribePortfolio(ACCOUNT_NO);

// Ping keepalive (mac dinh 30s)
stream.streaming.ping();

// Ctrl+C -> dong ket noi gon gang
process.on('SIGINT', () => {
  console.log('\nNgat ket noi...');
  stream.disconnect();
  process.exit(0);
});

// --- Buoc 5: Lang nghe lien tuc (tu dung sau 5 phut) ---
console.log('\nDang lang nghe... (Ctrl+C de dung)\n');
const TIMEOUT_MS = 300_000_000; // 5 phut
await Promise.race([
  stream.streaming.wait(),
  new Promise((resolve) => setTimeout(resolve, TIMEOUT_MS)),
]);
stream.disconnect();
console.log('Da dong ket noi.');
