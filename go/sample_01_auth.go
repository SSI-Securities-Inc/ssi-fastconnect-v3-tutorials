/*
Sample 1 — Xác thực, Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
==============================================================================
Đăng nhập, xử lý luồng OTP (Smart OTP Push Polling / OTP 6 số) và lấy token cho toàn bộ API call sau đó.

Luồng:
 1. Tạo Config -> Auth
 2. Kiểm tra token cache / refresh token nếu có.
 3. Nếu chưa có token: Gọi RequestOTP(), Polling Smart OTP hoặc nhập mã OTP 6 số.
 4. Auth service trả về accessToken, refreshToken, expiresIn và lưu vào token_cache.json
 5. Mọi request sau đó gắn Authorization: Bearer <accessToken>
 6. Xác nhận token bằng cách gọi API lấy thông tin tài khoản.
*/
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

	authObj := ssi.NewAuth(config)
	defer authObj.Close()

	// --- ensureAuth: load từ file nếu còn hạn, refresh hoặc request/verify OTP nếu cần ---
	ensureAuth(authObj, otp)

	accessToken := authObj.AccessToken()
	fmt.Println("\n--- Thông tin Token ---")
	if len(accessToken) > 40 {
		fmt.Printf("Access Token : %s...\n", accessToken[:40])
	} else {
		fmt.Printf("Access Token : %s\n", accessToken)
	}

	// --- Bước 5: Xác nhận token hoạt động bằng cách gọi API ---
	t := ssi.NewTrading(authObj)
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
	fmt.Printf("Bearer|%d|%d\n", authObj.TokenManager.Token().ExpiresAt, len(accounts))
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
		a.TokenManager.SetToken(token)

		if !a.TokenManager.IsTokenExpired() {
			fmt.Println("Token còn hạn, dùng token từ file.")
			return
		}

		if !a.TokenManager.IsRefreshTokenExpired() {
			fmt.Println("Access token hết hạn, đang refresh...")
			newToken, err := a.Refresh()
			if err != nil {
				log.Printf("Refresh thất bại (%v), tiến hành xác thực lại...", err)
			} else {
				saveToken(newToken)
				fmt.Println("Refresh token thành công.")
				return
			}
		}
	}

	fmt.Println("Không tìm thấy token hợp lệ, đang thực hiện quy trình xác thực & OTP...")
	if otp != "" {
		newToken, err := a.Authenticate(otp)
		if err != nil {
			log.Fatalf("Authenticate thất bại: %v", err)
		}
		saveToken(newToken)
		fmt.Println("Authenticate thành công.")
		return
	}

	fmt.Println("=== Yêu cầu OTP (Request OTP) ===")
	otpRes, err := a.RequestOTP()
	if err != nil {
		log.Fatalf("Lỗi Request OTP: %v", err)
	}

	var transactionID string
	if dataMap, ok := otpRes["data"].(map[string]interface{}); ok {
		if tid, ok := dataMap["transactionId"].(string); ok {
			transactionID = tid
		}
	}
	if transactionID == "" {
		if tid, ok := otpRes["transactionId"].(string); ok {
			transactionID = tid
		}
	}

	if transactionID != "" {
		fmt.Printf("[Smart OTP] Transaction ID: %s\n", transactionID)
		fmt.Println("Vui lòng mở ứng dụng SSI trên điện thoại và bấm APPROVE (Xác nhận)...")
		fmt.Println("SDK đang Polling chờ bạn bấm phê duyệt...")

		accessToken, err := a.EnsureAuthenticated("", transactionID)
		if err != nil {
			log.Fatalf("[LỖI/TIMEOUT] Phê duyệt Smart OTP thất bại: %v", err)
		}
		if a.TokenManager.Token() != nil {
			saveToken(a.TokenManager.Token())
		}
		fmt.Printf("Smart OTP xác thực thành công. Token: %s...\n", accessToken[:minInt(40, len(accessToken))])
	} else {
		fmt.Print("Vui lòng nhập mã OTP 6 số: ")
		reader := bufio.NewReader(os.Stdin)
		userOTP, _ := reader.ReadString('\n')
		userOTP = strings.TrimSpace(userOTP)

		if userOTP != "" {
			newToken, err := a.Authenticate(userOTP)
			if err != nil {
				log.Fatalf("Lỗi xác thực OTP: %v", err)
			}
			saveToken(newToken)
			fmt.Println("Authenticate OTP thành công.")
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

