/*
Sample 10 — WebSocket dữ liệu thị trường real-time
====================================================
Nhận tick data (giá khớp, bảng giá, room nước ngoài) tức thời.

Luồng:
 1. Client mở kết nối WebSocket bằng token hợp lệ
 2. Subscribe các stream dữ liệu theo symbol / index
 3. Server push event khi có giao dịch mới hoặc bảng giá thay đổi
 4. Parse message theo loại (Trade / Quote / ForeignRoom)
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

func main() {
	config := ssi.NewConfig("<CLIENT_ID>")
	config.APIKey = "<API_KEY>"
	config.APISecret = "<API_SECRET>"
	config.PrivateKey = "<PRIVATE_KEY_CONTENT>"
	config.LogLevel = "DEBUG"

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, "") // Không cần OTP cho market data

	s := ssi.NewStream(auth)
	defer s.Disconnect()

	// --- Bước 1: Mở kết nối WebSocket ---
	fmt.Println("Đang kết nối WebSocket...")
	if err := s.Connect(); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Đã kết nối!")

	// --- Bước 2: Đăng ký callback dữ liệu thị trường ---
	s.Streaming.SetOnData(func(msg interface{}) {
		switch m := msg.(type) {
		case *stream.TradeMessage:
			fmt.Printf("  [TRADE] %s | Giá: %.0f | KL: %d\n",
				m.Symbol, m.Price, m.Quantity)
		case *stream.QuoteMessage:
			fmt.Printf("  [QUOTE] %s | Bid: %v | Ask: %v\n",
				m.Symbol, m.BidPrices, m.AskPrices)
		case *stream.ForeignRoomMessage:
			fmt.Printf("  [ROOM]  %s | BuyQty: %d | SellQty: %d\n",
				m.Symbol, m.BuyQuantity, m.SellQuantity)
		default:
			fmt.Printf("  [DATA]  %v\n", msg)
		}
	})

	// --- Callback heartbeat ---
	s.Streaming.SetOnHeartbeat(func(msg *stream.HeartbeatMessage) {
		fmt.Printf("  [HEARTBEAT] %v\n", msg)
	})

	// --- Bước 3: Subscribe dữ liệu theo symbol ---
	fmt.Println("Subscribing dữ liệu symbol...")
	s.Streaming.SubscribeSymbol([]string{"SSI", "HPG", "VNM"}, nil)

	// --- Bước 4: Subscribe dữ liệu theo index ---
	fmt.Println("Subscribing dữ liệu index...")
	s.Streaming.SubscribeIndex([]string{"VNINDEX", "HNX-INDEX"}, nil)

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
