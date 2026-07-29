/*
Sample 1 — Xác thực và lấy Access Token
=========================================
Đăng nhập và lấy token cho toàn bộ API call sau đó.

Luồng:
 1. Client gửi username/password/appId tới Auth API
 2. Auth service trả về accessToken, refreshToken, expiresIn
 3. Lưu token vào session/runtime store
 4. Mọi request sau đó gắn Authorization: Bearer <accessToken>
 5. Nếu token hết hạn thì gọi refresh/re-login
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
	config, _, otp := loadConfig()

	auth := ssi.NewAuth(config)
	defer auth.Close()

	// --- ensureAuth: load từ file nếu còn hạn, refresh hoặc authenticate nếu cần ---
	ensureAuth(auth, otp)

	accessToken := auth.AccessToken()
	if len(accessToken) > 40 {
		fmt.Printf("Access Token : %s...\n", accessToken[:40])
	} else {
		fmt.Printf("Access Token : %s\n", accessToken)
	}

	// --- Bước 3: Token đã được SDK lưu tự động, mọi request kế tiếp ---
	//     sẽ gắn header Authorization: Bearer <accessToken>

	// --- Xác nhận token hoạt động bằng cách gọi API ---
	t := ssi.NewTrading(auth)
	accounts, err := t.Account.GetAccountInfo()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nXác thực thành công! Tìm thấy %d tài khoản:\n", len(accounts))
	for _, acc := range accounts {
		fmt.Printf("  - %s (%s)\n", acc.AccountNo, acc.AccountType)
	}
	// --- Response Summary ---
	fmt.Println("\n[Response] token_type|expires_at|account_count")
	fmt.Printf("Bearer|%d|%d\n", auth.TokenManager.Token().ExpiresAt, len(accounts))
	fmt.Println("[Response:account] account_no|account_type")
	for _, acc := range accounts {
		fmt.Printf("%s|%s\n", acc.AccountNo, acc.AccountType)
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
