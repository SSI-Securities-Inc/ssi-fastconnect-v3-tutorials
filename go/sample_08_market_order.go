/*
Sample 7 — Đặt lệnh Market (MP)
==================================
Khớp lệnh nhanh theo giá thị trường hiện tại.

Luồng:
 1. Client tạo order MARKET (không gửi price)
 2. Gửi request tới Trading Orders API
 3. Hệ thống match theo thanh khoản thị trường tại thời điểm gửi
 4. API trả về trạng thái khớp (FILLED hoặc PARTIALLY_FILLED)
 5. Cập nhật ngay danh mục/số dư tạm tính
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/trading"
)

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
	config, accountNo, otp := loadConfig()

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, otp)

	t := ssi.NewTrading(auth)

	// --- Bước 1: Kiểm tra sức mua/bán ở giá thị trường ---
	maxBS, err := t.Trading.GetMaxBuySellAtMarketPrice(accountNo, "SSI")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Max mua (market): %d cổ phiếu\n", maxBS.MaxBuyQuantity)
	fmt.Printf("Max bán (market): %d cổ phiếu\n", maxBS.MaxSellQuantity)

	// --- Bước 2: Đặt lệnh Market mua ---
	fmt.Println("\n--- Đặt lệnh MARKET mua SSI ---")
	result, err := t.Trading.PlaceMarketOrder(
		accountNo, "SSI", trading.OrderSideBuy, 100,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Kết quả: %v\n", result)

	// --- Bước 3: Kiểm tra trạng thái lệnh ---
	fmt.Println("\n--- Sổ lệnh hôm nay ---")
	orders, err := t.Portfolio.GetTodayOrders(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	start := 0
	if len(orders) > 3 {
		start = len(orders) - 3
	}
	for _, o := range orders[start:] {
		fmt.Printf("  %s | %s %s | SL: %d | Trạng thái: %s\n",
			o.OrderID, o.Symbol, o.Side, o.Quantity, o.Status)
	}

	// --- Bước 4: Cập nhật lại số dư sau khi khớp ---
	fmt.Println("\n--- Số dư sau giao dịch ---")
	balance, err := t.Portfolio.GetEquityBalance(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Tiền mặt khả dụng: %.0f\n", balance.AccountBalance)

	// --- Bước 5: Cập nhật danh mục ---
	fmt.Println("\n--- Vị thế sau giao dịch ---")
	positions, err := t.Portfolio.GetEquityPositions(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	for _, pos := range positions {
		if pos.Symbol == "SSI" {
			fmt.Printf("  SSI | SL: %d | Giá vốn: %.0f\n", pos.Quantity, pos.CostPrice)
		}
	}

	// --- Response Summary ---
	fmt.Println("\n[Response] max_buy_mkt|max_sell_mkt|buy_status")
	fmt.Printf("%d|%d|%s\n", maxBS.MaxBuyQuantity, maxBS.MaxSellQuantity, result.Status)
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
