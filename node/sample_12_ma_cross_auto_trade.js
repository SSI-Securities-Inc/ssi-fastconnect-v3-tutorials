/**
 * Sample 12 — MA Cross Signal Real-time (WebSocket bars)
 * =======================================================
 * Ket hop WebSocket de nhan tick real-time, tu aggregate thanh nen,
 * tinh MA5/MA10, dat lenh khi giao cat va theo doi lenh qua stream.
 *
 * Luong:
 *   1. Seed du lieu lich su 5m de khoi tao MA
 *   2. Mo WebSocket, subscribe trade stream (market data) va order stream (trading)
 *   3. Moi TradeMessage -> cap nhat nen hien tai (OHLCV)
 *   4. Khi nen dong (chuyen bucket thoi gian) -> tinh MA5/MA10 -> kiem tra giao cat
 *   5. Co tin hieu + khong co lenh dang cho -> kiem tra risk -> dat lenh MARKET
 *   6. OrderStatusMessage phan hoi trang thai lenh real-time, khong polling
 *   7. Ping/pong gui keepalive dinh ky de giu ket noi WebSocket
 */

import { Auth, Data, Stream, Trading, OrderSide, OrderStatus } from '@ssi.developer/ssi-sdk';
import { config, ACCOUNT_NO, OTP } from './config.js';
import { ensureAuth } from './auth_helper.js';

const SYMBOL = 'SSI';
const MA_FAST = 5;
const MA_SLOW = 10;
const QUANTITY = 100;
const BAR_SECONDS = 300; // Nen 5 phut

const vnd = (x) => Number(x).toLocaleString('en-US', { maximumFractionDigits: 0 });

const TERMINAL_STATUSES = new Set([
  OrderStatus.FILLED,
  OrderStatus.CANCELLED,
  OrderStatus.REJECTED,
  OrderStatus.EXPIRED,
  OrderStatus.PARTIAL_CANCELLED,
]);

// ---------------------------------------------------------------------------
// Bar builder — aggregate TradeMessage thanh nen OHLCV theo bucket thoi gian
// ---------------------------------------------------------------------------

class Bar {
  constructor(ts, open, high, low, close, volume) {
    this.ts = ts; // Unix timestamp mo nen (bucket)
    this.open = open;
    this.high = high;
    this.low = low;
    this.close = close;
    this.volume = volume;
  }
}

class BarBuilder {
  constructor(intervalSeconds) {
    this.interval = intervalSeconds;
    this._current = null;
    this.closed = []; // toi da 200 bars
  }

  /** Nap du lieu lich su de khoi tao MA. */
  seed(historicalBars) {
    for (const b of historicalBars) {
      this.closed.push(new Bar(0, b.openPrice, b.highPrice, b.lowPrice, b.closePrice, b.volume));
      if (this.closed.length > 200) this.closed.shift();
    }
    console.log(`  Seeded ${this.closed.length} historical bars`);
  }

  /**
   * Xu ly mot tick trade. Tra ve Bar vua dong neu chuyen bucket,
   * null neu nen hien tai chua dong.
   */
  onTrade(price, quantity) {
    const bucket = Math.floor(Date.now() / 1000 / this.interval) * this.interval;
    if (this._current === null || this._current.ts !== bucket) {
      const closed = this._current;
      this._current = new Bar(bucket, price, price, price, price, quantity);
      if (closed !== null) {
        this.closed.push(closed);
        if (this.closed.length > 200) this.closed.shift();
        return closed;
      }
      return null;
    }
    const b = this._current;
    b.high = Math.max(b.high, price);
    b.low = Math.min(b.low, price);
    b.close = price;
    b.volume += quantity;
    return null;
  }

  /** Lay toan bo bars (closed + current) duoi dang list. */
  snapshot() {
    const result = [...this.closed];
    if (this._current !== null) result.push(this._current);
    return result;
  }
}

// ---------------------------------------------------------------------------
// MA / Signal helpers
// ---------------------------------------------------------------------------

function calculateMa(bars, period) {
  if (bars.length < period) return null;
  const slice = bars.slice(-period);
  return slice.reduce((sum, b) => sum + b.close, 0) / period;
}

function detectCross(bars, fast, slow) {
  if (bars.length < slow + 1) return null;
  const mfNow = calculateMa(bars, fast);
  const msNow = calculateMa(bars, slow);
  const mfPrev = calculateMa(bars.slice(0, -1), fast);
  const msPrev = calculateMa(bars.slice(0, -1), slow);
  if ([mfNow, msNow, mfPrev, msPrev].includes(null)) return null;
  if (mfPrev <= msPrev && mfNow > msNow) return 'BUY';
  if (mfPrev >= msPrev && mfNow < msNow) return 'SELL';
  return null;
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

const auth = new Auth(config);
await ensureAuth(auth, OTP);

// ===== Buoc 1: Seed lich su 5m =====
const builder = new BarBuilder(BAR_SECONDS);
const data = new Data(auth);

console.log(`--- Load lich su OHLC 5m (${SYMBOL}) ---`);
const hist = await data.marketData.getOhlc5MinuteHistorical(
  SYMBOL, '2026/01/01 00:00:00', '2026/05/12 23:59:59', 1, MA_SLOW + 5,
);
builder.seed(hist);

// ===== State dung chung giua callbacks (Node single-threaded, khong can lock) =====
const state = {
  activeOrderId: null, // Lenh dang cho khop
  lastSignal: null,    // Tranh lap tin hieu cung chieu
};

const trading = new Trading(auth);

// ===== Buoc 2: Callbacks =====

/** Nhan TradeMessage -> cap nhat nen -> kiem tra tin hieu MA. */
async function onMarketData(msg) {
  if (msg.type !== 'trade') return;

  const closedBar = builder.onTrade(msg.price, msg.quantity);
  if (closedBar === null) return; // Nen chua dong, cho tick tiep theo

  // --- Nen vua dong ---
  const bars = builder.snapshot();
  const mf = calculateMa(bars, MA_FAST);
  const ms = calculateMa(bars, MA_SLOW);
  if (mf && ms) {
    console.log(`  [BAR] close=${vnd(closedBar.close)} vol=${vnd(closedBar.volume)} | MA${MA_FAST}=${mf.toFixed(2)} | MA${MA_SLOW}=${ms.toFixed(2)}`);
  } else {
    console.log(`  [BAR] close=${vnd(closedBar.close)} | Chua du du lieu MA`);
  }

  const signal = detectCross(bars, MA_FAST, MA_SLOW);
  if (signal === null) return;

  if (state.activeOrderId !== null) {
    console.log(`  [SIGNAL ${signal}] Dang co lenh cho, bo qua.`);
    return;
  }
  if (state.lastSignal === signal) return; // Khong lap tin hieu cung chieu
  state.lastSignal = signal;

  console.log(`\n  *** SIGNAL ${signal} ${SYMBOL} ***`);
  const side = signal === 'BUY' ? OrderSide.BUY : OrderSide.SELL;

  try {
    // --- Kiem tra risk ---
    const maxBs = await trading.trading.getMaxBuySellAtMarketPrice(ACCOUNT_NO, SYMBOL);
    const maxQty = signal === 'BUY' ? maxBs.maxBuyQuantity : maxBs.maxSellQuantity;
    if (maxQty < QUANTITY) {
      console.log(`  [RISK] Khong du ${QUANTITY} (co ${maxQty}). Bo qua.`);
      return;
    }

    // --- Dat lenh MARKET ---
    const result = await trading.trading.placeMarketOrder(ACCOUNT_NO, SYMBOL, side, QUANTITY);
    const orderId = result?.orderId || 'pending';
    console.log(`  [ORDER] Dat lenh thanh cong: orderId=${orderId}`);
    state.activeOrderId = orderId;
  } catch (err) {
    console.error(`  [ERROR] ${err.message}`);
    state.lastSignal = null; // Cho phep thu lai tin hieu nay
  }
}

/** Nhan OrderStatusMessage -> in trang thai, ghi P&L khi FILLED. */
function onTradingEvent(msg) {
  if (msg.type !== 'orderEvent') return;
  console.log(
    `  [ORDER UPDATE] ${msg.symbol} ${msg.side} | ID=${msg.orderId} | ` +
    `Status=${msg.status} | Khop=${msg.filledQuantity}/${msg.quantity}`,
  );
  if (!TERMINAL_STATUSES.has(msg.status)) return;

  state.activeOrderId = null;

  const filledQty = msg.filledQuantity || 0;
  // OrderStatusMessage co field `price` (gia lenh) — khong co avgPrice rieng.
  // Gia khop trung binh chinh xac nen lay tu getTodayOrders() -> Order.avgPrice neu can.
  const price = msg.price;
  if (msg.status === OrderStatus.FILLED && filledQty > 0 && price) {
    const cost = filledQty * price;
    console.log(`  [FILLED] ${state.lastSignal} ${msg.symbol}: ${filledQty} CP @ ${vnd(price)} | Tong: ${vnd(cost)} VND`);
  } else if (msg.status === OrderStatus.CANCELLED || msg.status === OrderStatus.REJECTED) {
    console.log(`  [CLOSED] Lenh ket thuc voi trang thai ${msg.status}`);
  }
}

function onHeartbeat(msg) {
  console.log(`  [HEARTBEAT] ${msg.status}`);
}

// ===== Buoc 3: Ket noi WebSocket =====
const stream = new Stream(auth);
stream.streaming.onData = onMarketData;
stream.streaming.onTrading = onTradingEvent;
stream.streaming.onHeartbeat = onHeartbeat;

console.log('\n--- Ket noi WebSocket ---');
await stream.streaming.connect();
console.log('Da ket noi!\n');

// Subscribe trade data de cap nhat nen
stream.streaming.subscribeSymbol([SYMBOL]);

// Subscribe order status de nhan ket qua lenh real-time
stream.streaming.subscribeOrderStatus(ACCOUNT_NO);

// ===== Buoc 4: Ping keepalive (SDK tu gui ping dinh ky, mac dinh 30s) =====
stream.streaming.ping();

// Ctrl+C -> dung chien luoc gon gang
process.on('SIGINT', () => {
  console.log('\nDung chien luoc.');
  stream.disconnect();
  process.exit(0);
});

console.log(`Dang lang nghe nen ${BAR_SECONDS}s cho ${SYMBOL}... (Ctrl+C de dung)\n`);
await stream.streaming.wait();
stream.disconnect();
