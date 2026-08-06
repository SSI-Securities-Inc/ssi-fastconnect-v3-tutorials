/*
Sample 5 — Lấy số dư tài khoản (Account Balance)
===================================================
Kiểm tra khả năng giao dịch trước khi đặt lệnh.

Luồng:
 1. Client gọi endpoint balances theo accountId
 2. API trả về available, onHold, limits, settlement
 3. Tính khả năng mua/bán thực tế theo nghiệp vụ
 4. Nếu không đủ điều kiện thì chặn thao tác đặt lệnh
 5. Nếu đủ điều kiện thì cho phép đi tiếp sang order flow
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

	// --- Bước 1: Lấy danh sách tài khoản ---
	accounts, err := t.Account.GetAccountInfo()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Danh sách tài khoản:")
	for _, acc := range accounts {
		fmt.Printf("  - %s (%s)\n", acc.AccountNo, acc.AccountType)
	}

	// --- Bước 2: Lấy số dư tài khoản Equity ---
	fmt.Printf("\n--- Số dư tài khoản Equity: %s ---\n", accountNo)
	balance, err := t.Portfolio.GetEquityBalance(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Tiền mặt khả dụng : %.0f\n", balance.AccountBalance)

	// --- Bước 3: Kiểm tra sức mua tối đa cho một mã ---
	fmt.Println("\n--- Sức mua/bán tối đa: SSI ---")
	price := 26000.0
	maxBS, err := t.Trading.GetMaxBuySell(accountNo, "SSI", &price)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Max mua  : %d cổ phiếu\n", maxBS.MaxBuyQuantity)
	fmt.Printf("  Max bán  : %d cổ phiếu\n", maxBS.MaxSellQuantity)

	// --- Bước 4: Logic kiểm tra trước khi đặt lệnh ---
	desiredQuantity := 100
	desiredPrice := 26000.0
	requiredAmount := float64(desiredQuantity) * desiredPrice

	if balance.AccountBalance >= requiredAmount {
		fmt.Printf("\n✓ Đủ điều kiện: cần %.0f, có %.0f\n", requiredAmount, balance.AccountBalance)
		fmt.Println("  → Cho phép đặt lệnh mua.")
	} else {
		fmt.Printf("\n✗ Không đủ: cần %.0f, chỉ có %.0f\n", requiredAmount, balance.AccountBalance)
		fmt.Println("  → Chặn đặt lệnh.")
	}

	// --- Bước 5: Xem vị thế hiện có ---
	fmt.Printf("\n--- Vị thế cổ phiếu (%s) ---\n", accountNo)
	positions, err := t.Portfolio.GetEquityPositions(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	for _, pos := range positions {
		fmt.Printf("  %-10s | SL: %d | Bán được: %d | Giá vốn: %.0f\n",
			pos.Symbol, pos.Quantity, pos.SellableQuantity, pos.CostPrice)
	}

	// --- Response Summary ---
	fmt.Println("\n[Response] accounts|avail_cash|max_buy_qty|max_sell_qty|positions")
	fmt.Printf("%d|%.0f|%d|%d|%d\n", len(accounts), balance.AccountBalance, maxBS.MaxBuyQuantity, maxBS.MaxSellQuantity, len(positions))
	if len(positions) > 0 {
		p := positions[0]
		fmt.Println("[Response:first_pos] symbol|quantity|sellable|cost_price")
		fmt.Printf("%s|%d|%d|%.0f\n", p.Symbol, p.Quantity, p.SellableQuantity, p.CostPrice)
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
