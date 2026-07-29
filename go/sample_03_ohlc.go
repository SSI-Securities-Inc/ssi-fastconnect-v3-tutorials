/*
Sample 3 — Lấy dữ liệu K-line (OHLC)
=======================================
Cung cấp dữ liệu nến cho biểu đồ và phân tích kỹ thuật.

Luồng:
 1. Client gửi symbol, interval, startTime, endTime
 2. API trả về mảng OHLCV theo mốc thời gian
 3. Chuẩn hóa dữ liệu (time/open/high/low/close/volume)
 4. Truyền vào chart component hoặc indicator engine
 5. Nếu lịch sử dài thì lặp theo paging/window thời gian
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
)

const symbol = "SSI"

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

func loadConfig() (*ssi.Config, string, string) {
	config := ssi.NewConfig("<CLIENT_ID>")
	config.APIKey = "<API_KEY>"
	config.APISecret = "<API_SECRET>"
	config.PrivateKey = "<PRIVATE_KEY_CONTENT>"
	accountNo := "<ACCOUNT_NO>"
	otp := "<OTP>"

	for _, path := range []string{"../config.json", "config.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var appCfg appConfig
		if err := json.Unmarshal(data, &appCfg); err == nil {
			if appCfg.ClientID != "" {
				config.ClientID = appCfg.ClientID
			}
			if appCfg.APIKey != "" {
				config.APIKey = appCfg.APIKey
			}
			if appCfg.APISecret != "" {
				config.APISecret = appCfg.APISecret
			}
			if appCfg.PrivateKey != "" {
				config.PrivateKey = appCfg.PrivateKey
			}
			if appCfg.EquityAccount != "" {
				accountNo = appCfg.EquityAccount
			}
			if appCfg.OTP != "" {
				otp = appCfg.OTP
			}
			break
		}
	}
	return config, accountNo, otp
}

func main() {
	config, _, otp := loadConfig()

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, otp)

	data := ssi.NewData(auth)

	// --- Bước 1: Lấy OHLC 1 ngày lịch sử ---
	fmt.Printf("--- OHLC 1 ngày lịch sử (%s) ---\n", symbol)
	daily, err := data.MarketData.GetOHLC1DayHistorical(symbol, "2026/03/01 00:00:00", "2026/03/27 23:59:59", 1, 100)
	fmt.Printf("%v\n", daily)
	if err != nil {
		log.Fatal(err)
	}
	limit := 5
	if len(daily) < limit {
		limit = len(daily)
	}
	for _, candle := range daily[:limit] {
		fmt.Printf("  %s | O:%.0f H:%.0f L:%.0f C:%.0f V:%d\n",
			candle.TradingDate, candle.OpenPrice, candle.HighPrice,
			candle.LowPrice, candle.ClosePrice, candle.Volume)
	}

	// --- Bước 2: Lấy OHLC 1 giờ trong ngày ---
	fmt.Printf("\n--- OHLC 1 giờ gần nhất (%s) ---\n", symbol)
	hourly, err := data.MarketData.GetOHLC1Hour(symbol)
	if err != nil {
		log.Fatal(err)
	}
	limit = 5
	if len(hourly) < limit {
		limit = len(hourly)
	}
	for _, candle := range hourly[:limit] {
		fmt.Printf("  %s | O:%.0f H:%.0f L:%.0f C:%.0f V:%d\n",
			candle.TradingDate, candle.OpenPrice, candle.HighPrice,
			candle.LowPrice, candle.ClosePrice, candle.Volume)
	}

	// --- Bước 3: Lấy OHLC 15 phút trong ngày ---
	fmt.Printf("\n--- OHLC 15 phút gần nhất (%s) ---\n", symbol)
	m15, err := data.MarketData.GetOHLC15Minute(symbol)
	if err != nil {
		log.Fatal(err)
	}
	limit = 5
	if len(m15) < limit {
		limit = len(m15)
	}
	for _, candle := range m15[:limit] {
		fmt.Printf("  %s | O:%.0f H:%.0f L:%.0f C:%.0f V:%d\n",
			candle.TradingDate, candle.OpenPrice, candle.HighPrice,
			candle.LowPrice, candle.ClosePrice, candle.Volume)
	}

	// --- Bước 4: Phân trang cho dữ liệu lớn ---
	fmt.Printf("\n--- Paging OHLC 1 phút lịch sử (%s) ---\n", symbol)
	page := 1
	totalBars := 0
	for {
		bars, err := data.MarketData.GetOHLC1MinuteHistorical(symbol, "2026/03/25 00:00:00", "2026/03/25 23:59:59", page, 100)
		if err != nil {
			log.Fatal(err)
		}
		if len(bars) == 0 {
			break
		}
		totalBars += len(bars)
		fmt.Printf("  Trang %d: %d nến (tổng: %d)\n", page, len(bars), totalBars)
		page++
	}
	fmt.Printf("\nTổng cộng %d nến 1 phút được tải.\n", totalBars)

	// --- Response Summary ---
	fmt.Println("\n[Response] daily_bars|hourly_bars|paging_1min")
	fmt.Printf("%d|%d|%d\n", len(daily), len(hourly), totalBars)
	if len(daily) > 0 {
		c := daily[0]
		fmt.Println("[Response:first_daily] date|open|high|low|close|volume")
		fmt.Printf("%s|%g|%g|%g|%g|%d\n",
			c.TradingDate, c.OpenPrice, c.HighPrice, c.LowPrice, c.ClosePrice, c.Volume)
	}
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

func ensureAuth(a *ssi.Auth, otp string) {
	token := loadToken()

	if token != nil {
		// Luôn set token từ cache vào manager trước
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

	fmt.Println("Không tìm thấy token hợp lệ, đang authenticate...")
	newToken, err := a.Authenticate(otp)
	if err != nil {
		log.Fatalf("Authenticate thất bại: %v", err)
	}
	saveToken(newToken)
	fmt.Println("Authenticate thành công.")
}
