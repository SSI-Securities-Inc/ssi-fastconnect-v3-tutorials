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

func main() {
	config := ssi.NewConfig("<CLIENT_ID>")
	config.APIKey = "<API_KEY>"
	config.APISecret = "<API_SECRET>"
	config.PrivateKey = "<PRIVATE_KEY_CONTENT>"
	config.LogLevel = "DEBUG"

	auth := ssi.NewAuth(config)
	defer auth.Close()

	// --- ensureAuth: load từ file nếu còn hạn, refresh hoặc authenticate nếu cần ---
	ensureAuth(auth, "")

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
	fmt.Printf("%v\n", accounts)
}

// ── Token cache helper ──────────────────────────────────────────────────────

const tokenCacheFile = "token_cache.json"

func loadToken() *auth.Token {
	data, err := os.ReadFile(tokenCacheFile)
	if err != nil {
		return nil
	}
	var token auth.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil
	}
	if token.AccessToken == "" {
		return nil
	}
	return &token
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
