/**
 * Sample 11 — WebSocket trading real-time (trang thai lenh & danh muc)
 * ======================================================================
 * Nhan cap nhat tuc thoi ve lenh khop va danh muc tai khoan.
 *
 * Luong:
 *   1. Client mo ket noi WebSocket bang token hop le (can OTP)
 *   2. Subscribe stream order status va portfolio theo accountNo
 *   3. Server push event khi trang thai lenh hoac danh muc thay doi
 *   4. Nhan message qua callback co kieu: onOrder / onOrderMatch / onPortfolio
 *   5. Ping dinh ky giu ket noi; Ctrl+C de dung
 *
 * Luu y SDK 3.3.0:
 *   - Co the van dung onTrading (nhan moi loai), nhung callback co kieu gon hon.
 *   - Truong thoi gian sua lenh tren OrderStatusMessage la `modifiedTime` (`modifyTime` da deprecated).
 *   - PortfolioMessage.totalAsset/cashBalance/stockValue da deprecated (server khong gui) — dung
 *     symbol/longQty/shortQty/net/floatingPl.
 */

import { Auth, Stream } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP, maskAccount } from './config.js';
import { ensureAuth } from './auth_helper.js';

// --- Callback co kieu cho tung loai su kien ---
function onOrder(msg) {
  console.log(
    `  [ORDER] ${msg.symbol} ${msg.side} | OrderID: ${msg.orderId} | ` +
    `Status: ${msg.status} | Khop: ${msg.filledQuantity}/${msg.quantity} | Cap nhat: ${msg.modifiedTime}`,
  );
  if (msg.errorCode) console.log(`    Loi: ${msg.errorCode} ${msg.errorMessage}`);
}

function onOrderMatch(msg) {
  console.log(
    `  [MATCH] ${msg.symbol} ${msg.side} | OrderID: ${msg.orderId} | ` +
    `Gia khop: ${msg.matchPrice} | KL khop: ${msg.matchQty} | ${msg.matchTime}`,
  );
}

function onPortfolio(msg) {
  console.log(
    `  [PORTFOLIO] Account: ${maskAccount(msg.accountNo)} | ${msg.symbol} | ` +
    `Long: ${msg.longQty} | Short: ${msg.shortQty} | Net: ${msg.net} | Floating P/L: ${msg.floatingPl ?? 'N/A'}`,
  );
}

function onHeartbeat(msg) {
  console.log(`  [HEARTBEAT] ${msg.status} - ${msg.message}`);
}

const auth = new Auth(config);
await ensureAuth(auth, OTP);

const stream = new Stream(auth);

// --- Buoc 1: Dang ky callback ---
stream.streaming.onOrder = onOrder;
stream.streaming.onOrderMatch = onOrderMatch;
stream.streaming.onPortfolio = onPortfolio;
stream.streaming.onHeartbeat = onHeartbeat;

// --- Buoc 2: Mo ket noi WebSocket ---
console.log('Dang ket noi WebSocket...');
await stream.streaming.connect();
console.log('Da ket noi!\n');

// --- Buoc 3: Subscribe trang thai lenh real-time (gom ca lenh khop, FCO) ---
console.log('Subscribing trang thai lenh...');
await stream.streaming.subscribeOrderStatus(ACCOUNT_NO);

// --- Buoc 4: Subscribe danh muc tai khoan real-time ---
console.log('Subscribing danh muc tai khoan...');
await stream.streaming.subscribePortfolio(ACCOUNT_NO);

// Ping keepalive — SDK khong tu ping, bat vong lap moi 30 giay
await stream.streaming.ping(undefined, 30_000);

// Ctrl+C -> dong ket noi gon gang
process.on('SIGINT', async () => {
  console.log('\nNgat ket noi...');
  await stream.disconnect();
  process.exit(0);
});

// --- Buoc 5: Lang nghe lien tuc (tu dung sau 5 phut) ---
console.log('\nDang lang nghe... (Ctrl+C de dung)\n');
const TIMEOUT_MS = 300_000; // 5 phut
await stream.streaming.wait(TIMEOUT_MS);
await stream.disconnect();
console.log('Da dong ket noi.');
