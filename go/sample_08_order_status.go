/*
Sample 8 — Kiểm tra trạng thái lệnh
======================================
Theo dõi tiến trình khớp của một lệnh cụ thể.

Luồng:
 1. Client gọi GET order by orderId
 2. API trả về status, filledQuantity, fills
 3. Đối chiếu lượng còn lại và trạng thái hiện tại
 4. Nếu chưa hoàn tất thì tiếp tục polling chu kỳ ngắn
 5. Khi FILLED/CANCELLED/REJECTED thì đóng vòng theo dõi
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
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/trading"
)

// Trạng thái kết thúc — không cần polling thêm
var terminalStatuses = map[string]bool{
	string(trading.OrderStatusFilled):           true,
	string(trading.OrderStatusCancelled):        true,
	string(trading.OrderStatusRejected):         true,
	string(trading.OrderStatusExpired):          true,
	string(trading.OrderStatusPartialCancelled): true,
}

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

	t := ssi.NewTrading(auth)

	// --- Bước 1: Đặt một lệnh để theo dõi ---
	fmt.Println("Đặt lệnh Limit mua SSI @ 26000...")
	result, err := t.Trading.PlaceLimitOrder(
		accountNo, "SSI", trading.OrderSideBuy, 100, 26000,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Kết quả đặt lệnh: %v\n", result)

	// --- Bước 2-5: Polling trạng thái ---
	fmt.Println("\n--- Bắt đầu theo dõi trạng thái lệnh ---")
	maxPolls := 10
	pollInterval := 3 * time.Second

	for i := 1; i <= maxPolls; i++ {
		orders, err := t.Portfolio.GetTodayOrders(accountNo)
		if err != nil {
			log.Fatal(err)
		}

		if len(orders) == 0 {
			fmt.Printf("  Poll %d: Chưa có lệnh trong sổ.\n", i)
			time.Sleep(pollInterval)
			continue
		}

		latest := orders[len(orders)-1]
		remaining := latest.Quantity - latest.FilledQuantity - latest.CancelQuantity

		fmt.Printf("  Poll %d: OrderID=%s | Status=%s | Khớp=%d/%d | Còn lại=%d\n",
			i, latest.OrderID, latest.Status,
			latest.FilledQuantity, latest.Quantity, remaining)

		if terminalStatuses[string(latest.Status)] {
			fmt.Printf("\n→ Lệnh đã kết thúc với trạng thái: %s\n", latest.Status)
			if latest.FilledQuantity > 0 {
				fmt.Printf("  Đã khớp: %d cổ phiếu @ trung bình %.0f\n",
					latest.FilledQuantity, latest.AvgPrice)
			}
			return
		}

		time.Sleep(pollInterval)
	}

	fmt.Printf("\nHết %d lần poll, lệnh vẫn đang mở.\n", maxPolls)
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
