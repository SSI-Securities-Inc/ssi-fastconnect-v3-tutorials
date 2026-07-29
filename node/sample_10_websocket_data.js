/**
 * Sample 10 — WebSocket du lieu thi truong real-time
 * =====================================================
 * Nhan tick data (gia khop, bang gia, room nuoc ngoai) tuc thoi.
 *
 * Luong:
 *   1. Client mo ket noi WebSocket bang token hop le (khong can OTP)
 *   2. Subscribe cac stream du lieu theo symbol / index
 *   3. Server push event khi co giao dich moi hoac bang gia thay doi
 *   4. Parse message theo loai (trade / quote / room)
 *   5. Ping dinh ky giu ket noi; Ctrl+C de dung
 */

import { Auth, Stream } from '@ssi.developer/ssi-sdk';
import { config } from './config.js';
import { ensureAuth } from './auth_helper.js';

// --- Callback xu ly du lieu thi truong ---
function onMarketData(msg) {
  if (msg.type === 'trade') {
    console.log(`  [TRADE] ${msg.symbol} | Gia: ${msg.price} | KL: ${msg.quantity} | Side: ${msg.side}`);
  } else if (msg.type === 'quote') {
    console.log(
      `  [QUOTE] ${msg.symbol} | ` +
      `Bid: ${JSON.stringify(msg.bidPrices.slice(0, 3))} | Ask: ${JSON.stringify(msg.askPrices.slice(0, 3))}`,
    );
  } else if (msg.type === 'room') {
    console.log(`  [ROOM]  ${msg.symbol} | Room con: ${msg.currentRoom}/${msg.totalRoom}`);
  } else {
    console.log(`  [DATA]  ${JSON.stringify(msg)}`);
  }
}

function onHeartbeat(msg) {
  console.log(`  [HEARTBEAT] ${msg.status} - ${msg.message}`);
}

const auth = new Auth(config);
await ensureAuth(auth); // Khong can OTP cho market data

const stream = new Stream(auth);

// --- Buoc 1: Dang ky callback truoc khi co du lieu ---
stream.streaming.onData = onMarketData;
stream.streaming.onHeartbeat = onHeartbeat;

// --- Buoc 2: Mo ket noi WebSocket ---
console.log('Dang ket noi WebSocket...');
await stream.streaming.connect();
console.log('Da ket noi!\n');

// --- Buoc 3: Subscribe du lieu theo symbol ---
console.log('Subscribing du lieu symbol...');
stream.streaming.subscribeSymbol(["41i1g8000"]);

// --- Buoc 4: Subscribe du lieu theo index ---
console.log('Subscribing du lieu index...');
stream.streaming.subscribeIndex(['VNINDEX', 'HNX-INDEX']);

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
const TIMEOUT_MS = 300_000; // 5 phut
await Promise.race([
  stream.streaming.wait(),
  new Promise((resolve) => setTimeout(resolve, TIMEOUT_MS)),
]);
stream.disconnect();
console.log('Da dong ket noi.');
