/**
 * Sample 16 — Streaming nang cao: board / index / ket noi / danh sach subscription
 * ==================================================================================
 * Mo rong sample_10 (du lieu thi truong) va sample_11 (lenh & danh muc).
 *
 * Luong:
 *   1. Dang ky onConnection (connected / reconnected / disconnected / failed) va doc
 *      isConnected / reconnectCount / lastError
 *   2. subscribeTrades / subscribeQuotes (ten ngan cua subscribeSymbolTrade / subscribeSymbolQuote)
 *   3. subscribeIndexTicks: tick cua CHINH chi so (VNINDEX...) + subscribeIndexSummary: tong hop chi so
 *   4. subscribeBoard (cap ma cua ca san) va subscribeIndex (co phieu THANH PHAN chi so) — luu luong
 *      lon, nen chi DEM message theo loai thay vi in tung dong
 *   5. subscribeMarketFlag: trang thai phien (ATO / LO / ATC ...)
 *   6. subscribeOrders + subscribePortfolio: su kien lenh khop, danh muc (can OTP lan dau)
 *   7. listSubscription: hoi server dang subscribe nhung topic nao
 *   8. unsubscribe* roi unsubscribeAll, disconnect
 *
 * Thoi gian chay: STREAM_SECONDS (mac dinh 60 giay). Ctrl+C de dung som.
 *
 * Luu y SDK 3.3.0:
 *   - subscribeBoard / subscribeIndex tu gui topic `@*` (server moi tach theo ma khi co `@*`);
 *     khong truyen dau phay trong mot topic (vd 'VNM,FPT') — truyen mang ['VNM', 'FPT'].
 *   - Khung nen chi co tick / 1m / 5m (STREAM_INTERVALS). Khung khac bi tu choi tai may.
 *   - subscribeIndex = co phieu thanh phan; subscribeIndexTicks = tick cua chi so. Ma chi so viet
 *     lien (HNXINDEX, khong phai 'HNX-INDEX'); UPCOMINDEX duoc doi thanh HNXUPCOMINDEX.
 *   - SDK khong tu ping: ping(undefined, 30_000) bat vong ping nen. Moi ham subscribe/unsubscribe/
 *     disconnect tra ve Promise -> await.
 */

import { Auth, Stream, Board, Timeframe, STREAM_INTERVALS } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP, maskAccount, assertConfigured } from './config.js';
import { ensureAuth } from './auth_helper.js';

assertConfigured({ trading: true });

const SECONDS = Number(process.env.STREAM_SECONDS) > 0 ? Number(process.env.STREAM_SECONDS) : 60;
const counts = new Map();
const bump = (key) => counts.set(key, (counts.get(key) ?? 0) + 1);

console.log(`Khung nen hop le cho stream: ${[...STREAM_INTERVALS].join(', ')}`);

const auth = new Auth(config);
await ensureAuth(auth, OTP); // can token co quyen trading de nhan su kien lenh
const stream = new Stream(auth);
const s = stream.streaming;

// --- Buoc 1: Su kien ket noi ---
s.onConnection = (event) => {
  // event: { state: 'connected' | 'reconnected' | 'disconnected' | 'failed', reconnectCount, error? }
  console.log(
    `  [CONNECTION] ${event.state} | reconnectCount: ${event.reconnectCount}` +
    (event.error ? ` | loi: ${event.error.message}` : ''),
  );
};

// --- Callback du lieu: dem theo loai, chi in mau vai message dau ---
let printed = 0;
s.onData = (msg) => {
  // Phan biet: nen (co intervalTime), tick chi so (co value), trade, quote... theo truong du lieu
  let kind = msg.type ?? 'data';
  if (msg.intervalTime !== undefined) kind = 'candle';
  else if (msg.type === 'trade' && msg.value !== undefined) kind = 'index-tick';
  bump(kind);
  if (printed < 5 && ['trade', 'index-tick', 'candle', 'quote'].includes(kind)) {
    printed += 1;
    const sample =
      kind === 'index-tick' ? { value: msg.value, change: msg.change }
      : kind === 'candle' ? { open: msg.open, close: msg.close, volume: msg.volume }
      : kind === 'quote' ? { bid: msg.bidPrices?.[0], ask: msg.askPrices?.[0] }
      : { price: msg.price, quantity: msg.quantity };
    console.log(`  [${kind.toUpperCase()}] ${msg.symbol} ${JSON.stringify(sample)}`);
  }
};
s.onIndexSummary = (msg) => {
  bump('index-summary');
  console.log(`  [INDEX-SUMMARY] ${msg.index} | ${Object.keys(msg.data).length} truong`);
};
s.onMarketFlag = (msg) => console.log(`  [MARKET-FLAG] ${msg.board} | ${msg.flag} | ${msg.tradingTime}`);
s.onOrder = (msg) =>
  console.log(`  [ORDER] ${msg.symbol} ${msg.side} | ${msg.orderId} | ${msg.status} | khop ${msg.filledQuantity}/${msg.quantity}`);
s.onOrderMatch = (msg) =>
  console.log(`  [MATCH] ${msg.symbol} ${msg.side} | ${msg.orderId} | ${msg.matchQty} @ ${msg.matchPrice}`);
s.onPortfolio = (msg) =>
  console.log(`  [PORTFOLIO] ${maskAccount(msg.accountNo)} | ${msg.symbol} | net ${msg.net}`);
s.onHeartbeat = () => bump('heartbeat');

console.log('Dang ket noi WebSocket...');
await s.connect();
console.log(`isConnected: ${s.isConnected} | reconnectCount: ${s.reconnectCount} | lastError: ${s.lastError ?? 'khong'}`);

async function shutdown() {
  console.log('\nDong subscription va ngat ket noi...');
  const removed = await s.unsubscribeAll().catch((err) => {
    console.log(`  unsubscribeAll loi: ${err.message}`);
    return null;
  });
  if (removed) console.log(`  Da go: ${removed.data.length} topic data, ${removed.trading.length} topic trading`);
  await stream.disconnect();
}
process.on('SIGINT', async () => {
  await shutdown();
  process.exit(0);
});

// --- Buoc 2-5: Subscribe du lieu ---
await s.subscribeTrades(['SSI', 'VNM']);
await s.subscribeQuotes(['SSI']);
await s.subscribeIndexTicks(['VNINDEX', 'VN30']);
await s.subscribeIndexSummary(['VNINDEX']);
await s.subscribeMarketFlag();
// Nen 1 phut cua SSI (interval chi duoc tick / 1m / 5m)
await s.subscribeCandles(['SSI'], Timeframe.MINUTE_1);
// Ca san HNX + co phieu thanh phan VN30: luu luong lon -> chi dem
await s.subscribeBoard([Board.HNX]);
await s.subscribeIndex(['VN30']);

// --- Buoc 6: Su kien tai khoan ---
await s.subscribeOrders(ACCOUNT_NO);
await s.subscribePortfolio(ACCOUNT_NO);

await s.ping(undefined, 30_000);

// --- Buoc 7: Server dang giu nhung topic nao? ---
const subs = await s.listSubscription();
console.log(`\nlistSubscription: ${subs.data.length} topic data, ${subs.trading.length} topic trading`);
console.log(`  Data (toi da 10): ${subs.data.slice(0, 10).join(', ')}`);
console.log(`  Trading: ${subs.trading.length} topic (so tai khoan an)`);

// --- Lang nghe, in thong ke moi 10 giay ---
console.log(`\nDang lang nghe ${SECONDS} giay... (Ctrl+C de dung)\n`);
const stats = setInterval(() => {
  console.log(`  [THONG KE] ${[...counts].map(([k, v]) => `${k}=${v}`).join(' | ') || '(chua co du lieu)'}`);
}, 10_000);
await s.wait(SECONDS * 1000);
clearInterval(stats);

// --- Buoc 8: Go subscription, ngat ket noi ---
await s.unsubscribeBoard([Board.HNX]);
await s.unsubscribeIndex(['VN30']);
await s.unsubscribeCandles(['SSI'], Timeframe.MINUTE_1);
await shutdown();
console.log(`isConnected: ${s.isConnected}. Hoan tat.`);
