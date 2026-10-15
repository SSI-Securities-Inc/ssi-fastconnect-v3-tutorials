/*
Sample 15 — Vị thế (positions), PPMMR (sức mua / ký quỹ) và sức mua/bán tối đa
================================================================================
Chỉ ĐỌC dữ liệu, không gửi lệnh.

Luồng:
 1. GetEquityPositions: vị thế cổ phiếu, gồm hàng chờ về T+1/T+2 cho cả chiều BÁN
    (T1SellQuantity/T2SellQuantity) và chiều MUA (T1BuyQuantity/T2BuyQuantity)
 2. GetEquityPPMMR: sức mua và ký quỹ, gồm tiền mua/bán chờ thanh toán T0/T1/T2
    (BuyT0/BuyT1/BuyT2, SellT0/SellT1/SellT2), MarginRatio, PurchasingPower
 3. GetMaxBuySell: khối lượng mua/bán tối đa — MarginRatio và PurchasingPower là con trỏ,
    nil ở tài khoản phái sinh
 4. (Tùy chọn) tài khoản phái sinh: GetDerivativePositions và GetDerivativePPMMR
    nếu config.json có derivative_account
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
)

func main() {
	config, app := loadConfig()
	mustHave("equity_account", app.EquityAccount)
	accountNo := app.EquityAccount

	a := ssi.NewAuth(config)
	defer a.Close()
	ensureAuth(a, app.OTP)

	t := ssi.NewTrading(a)
	fmt.Printf("Tài khoản cổ phiếu: %s\n", maskAccount(accountNo))

	// --- Bước 1: Vị thế cổ phiếu ---
	fmt.Println("\n--- Vị thế cổ phiếu ---")
	positions, err := t.Portfolio.GetEquityPositions(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	if len(positions) == 0 {
		fmt.Println("  Tài khoản chưa có vị thế.")
	}
	fmt.Printf("  %-8s %10s %10s | %9s %9s | %9s %9s | %10s\n",
		"Mã", "Tổng KL", "Bán được", "T1 bán", "T2 bán", "T1 mua", "T2 mua", "Giá vốn")
	for _, p := range positions {
		fmt.Printf("  %-8s %10d %10d | %9d %9d | %9d %9d | %10.0f\n",
			p.Symbol, p.Quantity, p.SellableQuantity,
			p.T1SellQuantity, p.T2SellQuantity, p.T1BuyQuantity, p.T2BuyQuantity, p.CostPrice)
	}

	// --- Bước 2: PPMMR cổ phiếu ---
	fmt.Println("\n--- PPMMR cổ phiếu ---")
	ppmmr, err := t.Portfolio.GetEquityPPMMR(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	var purchasingPower, marginRatio float64
	if ppmmr == nil {
		fmt.Println("  Server không trả PPMMR cổ phiếu cho tài khoản này.")
	} else {
		purchasingPower, marginRatio = ppmmr.PurchasingPower, ppmmr.MarginRatio
		fmt.Printf("  Sức mua (PurchasingPower): %.0f\n", ppmmr.PurchasingPower)
		fmt.Printf("  Tỷ lệ ký quỹ (MarginRatio): %.2f\n", ppmmr.MarginRatio)
		fmt.Printf("  Tiền mua chờ thanh toán : T0 %.0f | T1 %.0f | T2 %.0f\n", ppmmr.BuyT0, ppmmr.BuyT1, ppmmr.BuyT2)
		fmt.Printf("  Tiền bán chờ về         : T0 %.0f | T1 %.0f | T2 %.0f\n", ppmmr.SellT0, ppmmr.SellT1, ppmmr.SellT2)
	}

	// --- Bước 3: Sức mua/bán tối đa ---
	symbol := "SSI"
	if len(positions) > 0 && positions[0].Symbol != "" {
		symbol = positions[0].Symbol
	}
	fmt.Printf("\n--- Mua/bán tối đa %s (giá thị trường) ---\n", symbol)
	maxBS, err := t.Trading.GetMaxBuySellAtMarketPrice(accountNo, symbol)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Mua tối đa: %d | Bán tối đa: %d\n", maxBS.MaxBuyQuantity, maxBS.MaxSellQuantity)
	if maxBS.PurchasingPower != nil {
		fmt.Printf("  PurchasingPower: %d\n", *maxBS.PurchasingPower)
	}
	if maxBS.MarginRatio != nil {
		fmt.Printf("  MarginRatio: %s\n", *maxBS.MarginRatio)
	}

	// --- Bước 4: Tài khoản phái sinh (tùy chọn) ---
	derivOpen := -1
	if app.DerivativeAccount != "" && !strings.HasPrefix(app.DerivativeAccount, "YOUR_") {
		derivAccount := app.DerivativeAccount
		fmt.Printf("\n--- Phái sinh %s ---\n", maskAccount(derivAccount))
		all, err := t.Portfolio.GetDerivativePositions(derivAccount)
		if err != nil {
			log.Fatal(err)
		}
		if all == nil {
			fmt.Println("  Chưa có vị thế phái sinh.")
		} else {
			derivOpen = len(all.OpenPositions)
			for _, p := range all.OpenPositions {
				fmt.Printf("  %s | Long: %d | Short: %d | Net: %d | Lãi/lỗ tạm tính: %.0f\n",
					p.Symbol, p.Long, p.Short, p.Net, p.FloatingPL)
			}
			fmt.Printf("  Vị thế mở: %d | đã đóng: %d\n", len(all.OpenPositions), len(all.ClosedPositions))
		}
		dp, err := t.Portfolio.GetDerivativePPMMR(derivAccount)
		if err != nil {
			log.Fatal(err)
		}
		if dp != nil {
			fmt.Printf("  Số dư: %.0f | Floating P/L: %.0f | Trading P/L: %.0f | Rút được: %.0f\n",
				dp.AccountBalance, dp.FloatingPL, dp.TradingPL, dp.Withdrawable)
		}
	} else {
		fmt.Println("\n(Bỏ qua phần phái sinh: config.json chưa có derivative_account)")
	}

	// --- Response Summary ---
	fmt.Println("\n[Response] equity_positions|purchasing_power|margin_ratio|max_buy|max_sell|derivative_open")
	fmt.Printf("%d|%.0f|%.2f|%d|%d|%d\n",
		len(positions), purchasingPower, marginRatio, maxBS.MaxBuyQuantity, maxBS.MaxSellQuantity, derivOpen)
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
