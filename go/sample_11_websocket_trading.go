/*
Sample 11 — WebSocket trading real-time (trạng thái lệnh & danh mục)
======================================================================
Nhận cập nhật tức thời về lệnh khớp và danh mục tài khoản.

Luồng:
 1. Client mở kết nối WebSocket bằng token hợp lệ (cần OTP)
 2. Subscribe stream order_status và portfolio theo accountId
 3. Server push event khi trạng thái lệnh hoặc danh mục thay đổi
 4. Parse message theo loại (OrderStatus / OrderMatch / Portfolio) bằng callback typed
    SetOnOrder / SetOnOrderMatch / SetOnPortfolio (portfolio chỉ có ở tài khoản phái sinh)
 5. Khi mất kết nối, chạy cơ chế reconnect exponential backoff
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
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/stream"
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
	config.LogLevel = "DEBUG"

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, otp)

	s := ssi.NewStream(auth)
	defer s.Disconnect()

	// --- Bước 1: Mở kết nối WebSocket ---
	fmt.Println("Đang kết nối WebSocket...")
	if err := s.Connect(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Đã kết nối!")

	// --- Bước 2: Đăng ký callback sự kiện trading (theo từng loại message) ---
	s.Streaming.SetOnOrder(func(m *stream.OrderStatusMessage) {
		fmt.Printf("  [ORDER] %s %s | OrderID: %s | Status: %s | Khớp: %d/%d\n",
			m.Symbol, m.Side, m.OrderID, m.Status, m.FilledQuantity, m.Quantity)
	})
	s.Streaming.SetOnOrderMatch(func(m *stream.OrderMatchMessage) {
		fmt.Printf("  [MATCH] %s %s | OrderID: %s | KL: %d @ %.0f\n",
			m.Symbol, m.Side, m.OrderID, m.MatchQty, m.MatchPrice)
	})
	// Portfolio chỉ có ở tài khoản phái sinh. TotalAsset/CashBalance/StockValue luôn nil
	// vì server không gửi — đọc vị thế (Long/Short/Net) thay vì tổng tài sản.
	s.Streaming.SetOnPortfolio(func(m *stream.PortfolioMessage) {
		fmt.Printf("  [PORTFOLIO] %s | Long: %d | Short: %d | Net: %d\n",
			m.Symbol, m.LongQty, m.ShortQty, m.Net)
	})

	// --- Callback heartbeat ---
	s.Streaming.SetOnHeartbeat(func(msg *stream.HeartbeatMessage) {
		fmt.Printf("  [HEARTBEAT] %v\n", msg)
	})

	// --- Bước 3: Subscribe trạng thái lệnh real-time ---
	fmt.Println("Subscribing trạng thái lệnh...")
	if err := s.Streaming.SubscribeOrderStatus(accountNo, nil); err != nil {
		log.Fatal(err)
	}

	// --- Bước 4: Subscribe vị thế real-time (chỉ tài khoản phái sinh) ---
	// fmt.Println("Subscribing vị thế...")
	// s.Streaming.SubscribePortfolio(derivativeAccountNo, nil)

	// --- Bước 5: Lắng nghe liên tục ---
	fmt.Println("\nĐang lắng nghe... (Ctrl+C để dừng)")
	timeoutMs := 5 * 60 * 1000 // 5 phút, đơn vị mili-giây
	s.Wait(&timeoutMs)
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

// cleanOTP bỏ qua OTP rỗng hoặc còn là placeholder (<OTP>, YOUR_OTP_CODE) để rơi sang luồng Smart OTP / nhập tay.
func cleanOTP(otp string) string {
	v := strings.TrimSpace(otp)
	if v == "" || (strings.HasPrefix(v, "<") && strings.HasSuffix(v, ">")) || strings.HasPrefix(strings.ToUpper(v), "YOUR_") {
		return ""
	}
	return v
}

func ensureAuth(a *ssi.Auth, otpInput string) {
	otp := cleanOTP(otpInput)
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
