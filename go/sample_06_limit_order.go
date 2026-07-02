/*
Sample 6 — Đặt lệnh Limit (LO)
=================================
Đặt lệnh mua/bán tại mức giá chỉ định.

Luồng:
 1. Client tạo payload order (symbol, side, quantity, price, timeInForce)
 2. SDK tự gắn Idempotency-Key (clientRequestId) để chống submit trùng
 3. Gửi request tới Trading Orders API (có ký RSA)
 4. API trả về orderId và trạng thái ban đầu (PENDING)
 5. Lưu orderId để theo dõi khớp lệnh
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

const accountNo = "<ACCOUNT_NO>" // VD: "1234561"

func main() {
	config := ssi.NewConfig("<CLIENT_ID>")
	config.APIKey = "<API_KEY>"
	config.APISecret = "<API_SECRET>"
	config.PrivateKey = "<PRIVATE_KEY_CONTENT>"

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, "<OTP>")

	t := ssi.NewTrading(auth)

	// --- Bước 1: Kiểm tra sức mua trước ---
	price := 26000.0
	maxBS, err := t.Trading.GetMaxBuySell(accountNo, "SSI", &price)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Sức mua tối đa SSI @ 26,000: %d cổ phiếu\n", maxBS.MaxBuyQuantity)

	if maxBS.MaxBuyQuantity < 100 {
		fmt.Println("Không đủ sức mua, dừng lại.")
		return
	}

	// --- Bước 2: Đặt lệnh Limit mua ---
	fmt.Println("\n--- Đặt lệnh LIMIT mua SSI ---")
	buyResult, err := t.Trading.PlaceLimitOrder(
		accountNo, "SSI", trading.OrderSideBuy, 100, 28000,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Kết quả: %v\n", buyResult)

	// --- Bước 3: Đặt lệnh Limit bán ---
	fmt.Println("\n--- Đặt lệnh LIMIT bán SSI ---")
	sellResult, err := t.Trading.PlaceLimitOrder(
		accountNo, "SSI", trading.OrderSideSell, 100, 27000,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Kết quả: %v\n", sellResult)

	// --- Bước 4: Kiểm tra lệnh vừa đặt trong sổ lệnh ---
	fmt.Println("\n--- Sổ lệnh hôm nay ---")
	orders, err := t.Portfolio.GetTodayOrders(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	start := 0
	if len(orders) > 5 {
		start = len(orders) - 5
	}
	for _, o := range orders[start:] {
		fmt.Printf("  %s | %s %s | SL: %d @ %.0f | Trạng thái: %s\n",
			o.OrderID, o.Symbol, o.Side, o.Quantity, o.Price, o.Status)
	}

	// --- Response Summary ---
	fmt.Printf("\n[Response] max_buy_qty|buy_status|sell_status\n")
	fmt.Printf("%d|%s|%s\n", maxBS.MaxBuyQuantity, buyResult.Status, sellResult.Status)
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
