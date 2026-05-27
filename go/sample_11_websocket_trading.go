/*
Sample 11 — WebSocket trading real-time (trạng thái lệnh & danh mục)
======================================================================
Nhận cập nhật tức thời về lệnh khớp và danh mục tài khoản.

Luồng:
 1. Client mở kết nối WebSocket bằng token hợp lệ (cần OTP)
 2. Subscribe stream order_status và portfolio theo accountId
 3. Server push event khi trạng thái lệnh hoặc danh mục thay đổi
 4. Parse message theo loại (OrderStatus / Portfolio)
 5. Khi mất kết nối, chạy cơ chế reconnect exponential backoff
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/stream"
)

const accountNo = "<ACCOUNT_NO>" // VD: "1234561"

func main() {
	config := ssi.NewConfig("<CLIENT_ID>")
	config.APIKey = "<API_KEY>"
	config.APISecret = "<API_SECRET>"
	config.PrivateKey = "<PRIVATE_KEY_CONTENT>"
	config.LogLevel = "DEBUG"

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, "222222")

	s := ssi.NewStream(auth)
	defer s.Disconnect()

	// --- Bước 1: Mở kết nối WebSocket ---
	fmt.Println("Đang kết nối WebSocket...")
	if err := s.Connect(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Đã kết nối!")

	// --- Bước 2: Đăng ký callback sự kiện trading ---
	s.Streaming.SetOnTrading(func(msg interface{}) {
		switch m := msg.(type) {
		case *stream.OrderStatusMessage:
			fmt.Printf("  [ORDER] %s %s | OrderID: %s | Status: %s\n",
				m.Symbol, m.Side, m.OrderID, m.Status)
		case *stream.PortfolioMessage:
			fmt.Printf("  [PORTFOLIO] Account: %s | Tổng TS: %.0f\n",
				m.AccountNo, m.TotalAsset)
		default:
			fmt.Printf("  [TRADING] %v\n", msg)
		}
	})

	// --- Callback heartbeat ---
	s.Streaming.SetOnHeartbeat(func(msg *stream.HeartbeatMessage) {
		fmt.Printf("  [HEARTBEAT] %v\n", msg)
	})

	// --- Bước 3: Subscribe trạng thái lệnh real-time ---
	fmt.Println("Subscribing trạng thái lệnh...")
	s.Streaming.SubscribeOrderStatus("6666661", nil)

	// --- Bước 4: Subscribe danh mục tài khoản real-time ---
	// fmt.Println("Subscribing danh mục tài khoản...")
	// s.Streaming.SubscribePortfolio(accountNo, nil)

	// --- Bước 5: Lắng nghe liên tục ---
	fmt.Println("\nĐang lắng nghe... (Ctrl+C để dừng)")
	timeout := 5 * time.Minute
	s.Wait(&timeout)
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
