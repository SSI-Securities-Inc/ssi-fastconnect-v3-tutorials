/*
Sample 13 — Đặt Lệnh Điều Kiện (FCO)
=====================================
Thể hiện đầy đủ các loại lệnh điều kiện (Fast Conditional Orders - FCO):
 1. GTD (Good-Till-Date / Lệnh chờ theo ngày)
 2. Stop (Lệnh dừng giá thị trường)
 3. Stop Limit (Lệnh dừng giá giới hạn)
 4. Trailing Stop (Lệnh dừng xu hướng)
 5. Trailing Stop Limit (Lệnh dừng xu hướng giới hạn)
 6. OCO (One-Cancels-the-Other / Lệnh Chốt lời & Cắt lỗ)
 7. Bull Bear (Lệnh Hai đầu)
 8. Truy vấn danh sách & Hủy lệnh FCO
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

	a := ssi.NewAuth(config)
	defer a.Close()

	ensureAuth(a, otp)

	t := ssi.NewTrading(a)

	symbol := "SSI"
	fromDate := "2026/08/01 00:00:00"
	toDate := "2026/08/30 23:59:59"

	fmt.Println("=== FASTCONNECT GO SDK — SAMPLE 13: LỆNH ĐIỀU KIỆN (FCO) ===\n")

	// --- 1. Lệnh GTD (Good-Till-Date) ---
	fmt.Println("--- 1. Đặt lệnh GTD ---")
	gtdRes, err := t.Trading.PlaceFcoGtd(
		accountNo, symbol, trading.OrderSideBuy, 100, 26000.0, 0, fromDate, toDate,
	)
	if err != nil {
		fmt.Printf("Lỗi đặt GTD: %v\n", err)
	} else {
		fmt.Printf("  GTD Result: FcoID=%s\n", gtdRes.FCOID)
	}

	// --- 2. Lệnh Stop (Stop Market) ---
	fmt.Println("\n--- 2. Đặt lệnh Stop ---")
	stopRes, err := t.Trading.PlaceFcoStop(
		accountNo, symbol, trading.OrderSideBuy, 100, 27000.0, trading.FCOOperatorGreaterOrEqual, fromDate, toDate,
	)
	if err != nil {
		fmt.Printf("Lỗi đặt Stop: %v\n", err)
	} else {
		fmt.Printf("  Stop Result: FcoID=%s\n", stopRes.FCOID)
	}

	// --- 3. Lệnh Stop Limit ---
	fmt.Println("\n--- 3. Đặt lệnh Stop Limit ---")
	stopLimitRes, err := t.Trading.PlaceFcoStopLimit(
		accountNo, symbol, trading.OrderSideBuy, 100, 27500.0, 0, 27000.0, trading.FCOOperatorGreaterOrEqual, fromDate, toDate,
	)
	if err != nil {
		fmt.Printf("Lỗi đặt Stop Limit: %v\n", err)
	} else {
		fmt.Printf("  Stop Limit Result: FcoID=%s\n", stopLimitRes.FCOID)
	}

	// --- 4. Lệnh Trailing Stop ---
	fmt.Println("\n--- 4. Đặt lệnh Trailing Stop ---")
	trailingRes, err := t.Trading.PlaceFcoTrailingStop(
		accountNo, symbol, trading.OrderSideSell, 100, 28000.0, 1000.0, fromDate, toDate,
	)
	if err != nil {
		fmt.Printf("Lỗi đặt Trailing Stop: %v\n", err)
	} else {
		fmt.Printf("  Trailing Stop Result: FcoID=%s\n", trailingRes.FCOID)
	}

	// --- 5. Lệnh Trailing Stop Limit ---
	fmt.Println("\n--- 5. Đặt lệnh Trailing Stop Limit ---")
	trailingLimitRes, err := t.Trading.PlaceFcoTrailingStopLimit(
		accountNo, symbol, trading.OrderSideSell, 100, 28000.0, 1000.0, 500.0, fromDate, toDate,
	)
	if err != nil {
		fmt.Printf("Lỗi đặt Trailing Stop Limit: %v\n", err)
	} else {
		fmt.Printf("  Trailing Stop Limit Result: FcoID=%s\n", trailingLimitRes.FCOID)
	}

	// --- 6. Lệnh OCO (One-Cancels-the-Other) ---
	fmt.Println("\n--- 6. Đặt lệnh OCO ---")
	ocoRes, err := t.Trading.PlaceFcoOco(
		accountNo, symbol, trading.OrderSideSell, 100, 30000.0, 24000.0, 30000.0, 24000.0, 0, 0, fromDate, toDate,
	)
	if err != nil {
		fmt.Printf("Lỗi đặt OCO: %v\n", err)
	} else {
		fmt.Printf("  OCO Result: FcoID=%s\n", ocoRes.FCOID)
	}

	// --- 7. Lệnh Bull Bear ---
	fmt.Println("\n--- 7. Đặt lệnh Bull Bear ---")
	bbRes, err := t.Trading.PlaceFcoBullBear(
		accountNo, symbol, trading.OrderSideBuy, 100, 26000.0, 0, 30000.0, 24000.0, 30000.0, 24000.0, 0, 0, fromDate, toDate,
	)
	if err != nil {
		fmt.Printf("Lỗi đặt Bull Bear: %v\n", err)
	} else {
		fmt.Printf("  Bull Bear Result: FcoID=%s\n", bbRes.FCOID)
	}

	// --- 8. Truy vấn danh sách lệnh FCO ---
	fmt.Println("\n--- 8. Danh sách lệnh FCO ---")
	fcoList, err := t.Trading.GetFcoByAccountNo(accountNo, 1, 10)
	if err != nil {
		fmt.Printf("Lỗi lấy danh sách FCO: %v\n", err)
	} else {
		fmt.Printf("  Tổng số lệnh FCO: %d\n", fcoList.ItemsCount)
		count := len(fcoList.FCOList)
		if count > 5 {
			count = 5
		}
		for _, item := range fcoList.FCOList[:count] {
			fmt.Printf("  FCO ID: %s | Mã: %s | Loại: %s | Trạng thái: %s\n",
				item.FCOID, item.Symbol, item.Type, item.Status)
		}
	}

	// --- 9. Hủy lệnh FCO vừa tạo nếu có ---
	if gtdRes != nil && gtdRes.FCOID != "" {
		fmt.Printf("\n--- 9. Hủy lệnh FCO ID: %s ---\n", gtdRes.FCOID)
		cancelRes, err := t.Trading.CancelFco(gtdRes.FCOID)
		if err != nil {
			fmt.Printf("Lỗi hủy FCO: %v\n", err)
		} else {
			fmt.Printf("  Hủy FCO Result: FcoID=%s\n", cancelRes.FCOID)
		}
	}

	fmt.Println("\n[Response] sample_13_fco_completed")
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
