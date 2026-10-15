/*
Sample 16 — Streaming nâng cao (board, index, interval, sự kiện kết nối, quản lý subscription)
===============================================================================================
Chỉ NHẬN dữ liệu, không gửi lệnh. Sample chạy khoảng 40 giây rồi tự thoát.

Luồng:
 1. Đăng ký SetOnConnection: State (connected / reconnected / disconnected / failed),
    ReconnectCount, Err; cộng IsConnected / ReconnectCount / LastError trên Streaming
 2. Subscribe theo mã: trade, quote, nến 1 phút (interval chỉ có tick / 1m / 5m)
 3. Subscribe tick và nến của CHÍNH chỉ số (SubscribeIndexTrade), index summary, cờ phiên
    (market flag) và sự kiện lệnh / khớp lệnh (order, order match)
 4. ListSubscription: hỏi server đang subscribe những topic nào
 5. SubscribeBoard / SubscribeIndex: trade + quote + room của TOÀN BỘ mã trong sàn / rổ
    (topic dạng trade.HOSE@*) — lượng tin rất lớn nên sample chỉ đếm, không in từng tin
 6. UnsubscribeBoard / UnsubscribeIndex / UnsubscribeSymbolQuote / UnsubscribeIndexTrade,
    cuối cùng UnsubscribeAll dọn mọi topic còn lại
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync/atomic"

	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/market"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/stream"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/transport"
)

const (
	streamSymbol = "SSI"
	streamIndex  = "VNINDEX"
	basketIndex  = "VN30"
)

var (
	trades, quotes, rooms, indexTicks, candles atomic.Int64
)

// firstFew chỉ cho in 5 tin đầu của mỗi loại để màn hình không bị tràn.
func firstFew(c *atomic.Int64) bool { return c.Add(1) <= 5 }

func main() {
	config, app := loadConfig()
	mustHave("equity_account", app.EquityAccount)
	accountNo := app.EquityAccount

	a := ssi.NewAuth(config)
	defer a.Close()
	ensureAuth(a, app.OTP)

	s := ssi.NewStream(a)
	defer s.Disconnect()

	// --- Bước 1: Sự kiện kết nối (đăng ký TRƯỚC khi Connect để không bỏ lỡ "connected") ---
	s.Streaming.SetOnConnection(func(e transport.ConnectionEvent) {
		line := fmt.Sprintf("  [CONN] state=%s | reconnect=%d", e.State, e.ReconnectCount)
		if e.Err != nil {
			line += fmt.Sprintf(" | lỗi: %v", e.Err)
		}
		fmt.Println(line)
		if e.State == "failed" {
			fmt.Println("  Kết nối thất bại vĩnh viễn (ví dụ cần OTP mới) — SDK sẽ không tự kết nối lại.")
		}
	})

	// --- Callback dữ liệu: một handler chung, tách loại bằng type switch ---
	s.Streaming.SetOnData(func(msg interface{}) {
		switch m := msg.(type) {
		case *stream.TradeMessage:
			if firstFew(&trades) {
				fmt.Printf("  [TRADE] %s | Giá: %.0f | KL: %d\n", m.Symbol, m.Price, m.Quantity)
			}
		case *stream.QuoteMessage:
			if firstFew(&quotes) && len(m.BidPrices) > 0 && len(m.AskPrices) > 0 {
				fmt.Printf("  [QUOTE] %s | Bid1: %.0f | Ask1: %.0f\n", m.Symbol, m.BidPrices[0], m.AskPrices[0])
			}
		case *stream.ForeignRoomMessage:
			if firstFew(&rooms) {
				fmt.Printf("  [ROOM]  %s | Mua NN: %d | Bán NN: %d\n", m.Symbol, m.BuyQuantity, m.SellQuantity)
			}
		case *stream.IndexTickMessage:
			// Các trường là con trỏ: nil nghĩa là server không gửi khóa đó.
			if firstFew(&indexTicks) && m.Value != nil {
				fmt.Printf("  [INDEX] %s | Giá trị: %.2f\n", m.Symbol, *m.Value)
			}
		case *stream.IntervalMessage:
			if firstFew(&candles) {
				fmt.Printf("  [NẾN]   %s | %s | O:%.0f H:%.0f L:%.0f C:%.0f V:%d\n",
					m.Symbol, m.IntervalTime, m.Open, m.High, m.Low, m.Close, m.Volume)
			}
		}
	})
	s.Streaming.SetOnIndexSummary(func(m *stream.IndexSummaryMessage) {
		// Server không công bố danh sách trường của index summary nên SDK giữ nguyên trong Data.
		keys := make([]string, 0, len(m.Data))
		for k := range m.Data {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		fmt.Printf("  [SUMMARY] %s | các trường: %s\n", m.Index, strings.Join(keys, ","))
	})
	s.Streaming.SetOnMarketFlag(func(m *stream.MarketFlagMessage) {
		fmt.Printf("  [FLAG]  sàn %s | phiên %s | %s\n", m.Board, m.Flag, m.TradingTime)
	})
	s.Streaming.SetOnOrder(func(m *stream.OrderStatusMessage) {
		fmt.Printf("  [ORDER] %s %s | OrderID: %s | Status: %s | Khớp: %d/%d\n",
			m.Symbol, m.Side, m.OrderID, m.Status, m.FilledQuantity, m.Quantity)
	})
	s.Streaming.SetOnOrderMatch(func(m *stream.OrderMatchMessage) {
		fmt.Printf("  [MATCH] %s %s | OrderID: %s | KL: %d @ %.0f\n",
			m.Symbol, m.Side, m.OrderID, m.MatchQty, m.MatchPrice)
	})

	fmt.Println("Đang kết nối WebSocket...")
	if err := s.Connect(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("IsConnected=%v | ReconnectCount=%d | LastError=%v\n",
		s.Streaming.IsConnected(), s.Streaming.ReconnectCount(), s.Streaming.LastError())

	// --- Bước 2-3: Subscribe theo mã, theo chỉ số, cờ phiên, sự kiện lệnh ---
	symbols := []string{streamSymbol}
	indices := []string{streamIndex}
	check("SubscribeSymbolTrade", s.Streaming.SubscribeSymbolTrade(symbols, nil))
	check("SubscribeSymbolQuote", s.Streaming.SubscribeSymbolQuote(symbols, nil))
	check("SubscribeSymbolOhlcv 1m", s.Streaming.SubscribeSymbolOhlcv(symbols, stream.IntervalMinute1, nil))
	check("SubscribeIndexTrade tick", s.Streaming.SubscribeIndexTrade(indices, stream.IntervalTick, nil))
	check("SubscribeIndexTrade 1m", s.Streaming.SubscribeIndexTrade(indices, stream.IntervalMinute1, nil))
	check("SubscribeIndexSummary", s.Streaming.SubscribeIndexSummary(indices, nil))
	check("SubscribeMarketFlag", s.Streaming.SubscribeMarketFlag(nil))
	// Một lần subscribe order.<tài khoản> nhận cả trạng thái lệnh lẫn khớp lệnh.
	check("SubscribeOrderStatus", s.Streaming.SubscribeOrderStatus(accountNo, nil))
	fmt.Printf("Đã subscribe mã %s, chỉ số %s và sự kiện lệnh của %s.\n", streamSymbol, streamIndex, maskAccount(accountNo))

	// --- Bước 4: Server đang giữ những topic nào ---
	printSubscriptions(s, "Subscription hiện tại")

	fmt.Println("\nĐang lắng nghe 20 giây...")
	waitMs := 20 * 1000
	s.Wait(&waitMs)
	printCounters("Sau pha 1")

	// --- Bước 5: Board và rổ chỉ số — lượng tin lớn, chỉ đếm ---
	fmt.Printf("\nSubscribeBoard(HOSE) và SubscribeIndex(%s) trong 10 giây (chỉ đếm)...\n", basketIndex)
	check("SubscribeBoard", s.Streaming.SubscribeBoard([]market.Board{market.BoardHOSE}, nil))
	check("SubscribeIndex", s.Streaming.SubscribeIndex([]string{basketIndex}, nil))
	waitMs = 10 * 1000
	s.Wait(&waitMs)
	printCounters("Sau pha 2")

	// --- Bước 6: Unsubscribe từng phần ---
	fmt.Println("\nUnsubscribe board, rổ chỉ số, quote và tick chỉ số...")
	check("UnsubscribeBoard", s.Streaming.UnsubscribeBoard([]market.Board{market.BoardHOSE}))
	check("UnsubscribeIndex", s.Streaming.UnsubscribeIndex([]string{basketIndex}))
	check("UnsubscribeSymbolQuote", s.Streaming.UnsubscribeSymbolQuote(symbols))
	check("UnsubscribeIndexTrade tick", s.Streaming.UnsubscribeIndexTrade(indices, stream.IntervalTick))
	printSubscriptions(s, "Subscription còn lại")

	removed, err := s.Streaming.UnsubscribeAll()
	if err != nil {
		log.Printf("UnsubscribeAll lỗi: %v", err)
	}
	removedCount := 0
	for _, topics := range removed {
		removedCount += len(topics)
	}
	fmt.Printf("UnsubscribeAll đã gỡ %d topic.\n", removedCount)

	// --- Response Summary ---
	fmt.Println("\n[Response] trades|quotes|rooms|index_ticks|candles|topics_removed|reconnects")
	fmt.Printf("%d|%d|%d|%d|%d|%d|%d\n", trades.Load(), quotes.Load(), rooms.Load(), indexTicks.Load(),
		candles.Load(), removedCount, s.Streaming.ReconnectCount())
}

func check(what string, err error) {
	if err != nil {
		log.Fatalf("%s thất bại: %v", what, err)
	}
}

// printSubscriptions dùng ListSubscription (timeout tính bằng mili-giây, <= 0 là 5000).
func printSubscriptions(s *ssi.Stream, title string) {
	subs, err := s.Streaming.ListSubscription(5000)
	if err != nil {
		log.Printf("ListSubscription lỗi: %v", err)
		return
	}
	fmt.Printf("\n%s:\n", title)
	for _, channel := range []string{"data", "trading"} {
		fmt.Printf("  %-8s %d topic: %s\n", channel, len(subs[channel]), strings.Join(subs[channel], ", "))
	}
}

func printCounters(title string) {
	fmt.Printf("%s: trade=%d quote=%d room=%d index=%d nến=%d\n",
		title, trades.Load(), quotes.Load(), rooms.Load(), indexTicks.Load(), candles.Load())
}

type appConfig struct {
	ClientID          string `json:"client_id"`
	APIKey            string `json:"api_key"`
	APISecret         string `json:"api_secret"`
	PrivateKey        string `json:"private_key"`
	OTP               string `json:"otp"`
	EquityAccount     string `json:"equity_account"`
	DerivativeAccount string `json:"derivative_account"`
	LogLevel          string `json:"log_level"`
}

// loadConfig đọc config.json và dừng ngay khi thiếu trường bắt buộc (không có giá trị mặc định).
func loadConfig() (*ssi.Config, *appConfig) {
	for _, path := range []string{"../config.json", "config.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var app appConfig
		if err := json.Unmarshal(data, &app); err != nil {
			log.Fatalf("%s không phải JSON hợp lệ", path)
		}
		mustHave("client_id", app.ClientID)
		mustHave("api_key", app.APIKey)
		mustHave("api_secret", app.APISecret)

		config := ssi.NewConfig(app.ClientID)
		config.APIKey = app.APIKey
		config.APISecret = app.APISecret
		config.PrivateKey = app.PrivateKey
		if app.LogLevel != "" {
			config.LogLevel = app.LogLevel
		}
		return config, &app
	}
	log.Fatal("Không tìm thấy config.json (../config.json hoặc config.json). Xem ../config.example.json")
	return nil, nil
}

func mustHave(field, value string) {
	if value == "" || strings.HasPrefix(value, "YOUR_") {
		log.Fatalf("Thiếu cấu hình %q trong config.json", field)
	}
}

// maskAccount che số tài khoản khi in ra màn hình/log: giữ 2 ký tự đầu và 2 ký tự cuối.
func maskAccount(acc string) string {
	if len(acc) <= 4 {
		return "****"
	}
	return acc[:2] + strings.Repeat("*", len(acc)-4) + acc[len(acc)-2:]
}

// ── Token cache helper ──────────────────────────────────────────────────────

const sharedTokenFile = "../shared_token.json"
const tokenCacheFile = "token_cache.json"

func loadToken() *auth.Token {
	for _, file := range []string{sharedTokenFile, tokenCacheFile} {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		var token auth.Token
		if err := json.Unmarshal(data, &token); err != nil {
			continue
		}
		if token.AccessToken != "" {
			return &token
		}
	}
	return nil
}

func saveToken(token *auth.Token) {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		log.Printf("Lỗi khi serialize token: %v", err)
		return
	}
	if err := os.WriteFile(tokenCacheFile, data, 0600); err != nil {
		log.Printf("Lỗi khi lưu token: %v", err)
		return
	}
	fmt.Printf("Token đã lưu vào %s\n", tokenCacheFile)
}

// ensureAuth dùng token cache nếu còn hạn, refresh nếu được, còn không thì đăng nhập bằng OTP.
func ensureAuth(a *ssi.Auth, otp string) {
	if token := loadToken(); token != nil {
		a.TokenManager.SetToken(token)

		if !a.TokenManager.IsTokenExpired() {
			fmt.Println("Token còn hạn, dùng token từ file.")
			return
		}

		if !a.TokenManager.IsRefreshTokenExpired() {
			fmt.Println("Access token hết hạn, đang refresh...")
			newToken, err := a.Refresh()
			if err != nil {
				log.Printf("Refresh thất bại (%v), đang authenticate lại...", err)
			} else {
				saveToken(newToken)
				fmt.Println("Refresh token thành công.")
				return
			}
		}
	}

	mustHave("otp", otp)
	fmt.Println("Không có token hợp lệ, đang authenticate...")
	newToken, err := a.Authenticate(otp)
	if err != nil {
		log.Fatalf("Authenticate thất bại: %v", err)
	}
	saveToken(newToken)
	fmt.Println("Authenticate thành công.")
}
