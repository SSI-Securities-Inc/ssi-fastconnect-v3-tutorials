/*
Sample 9 — Hủy lệnh
=====================
Dừng phần khối lượng chưa khớp của lệnh đang mở.

Luồng:
 1. User bấm hủy trên lệnh đang PENDING/PARTIALLY_FILLED
 2. Client gửi DELETE order kèm thông tin account/symbol
 3. API xác thực quyền và trạng thái lệnh hiện tại
 4. Nếu hợp lệ, hệ thống cập nhật CANCELLED cho phần chưa khớp
 5. Đồng bộ lại sổ lệnh và số lượng còn treo
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

// Trạng thái có thể hủy
var cancellableStatuses = map[string]bool{
	string(trading.OrderStatusPendingApproval): true,
	string(trading.OrderStatusReady):           true,
	string(trading.OrderStatusSent):            true,
	string(trading.OrderStatusQueued):          true,
	string(trading.OrderStatusPartialFilled):   true,
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

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, otp)

	t := ssi.NewTrading(auth)

	// --- Bước 1: Lấy sổ lệnh, tìm lệnh đang mở ---
	fmt.Println("--- Sổ lệnh hôm nay ---")
	orders, err := t.Portfolio.GetTodayOrders(accountNo)
	if err != nil {
		log.Fatal(err)
	}

	var openOrders []struct {
		OrderID        string
		Symbol         string
		Side           string
		Quantity       int
		Price          float64
		FilledQuantity int
		Status         string
	}

	for _, o := range orders {
		if cancellableStatuses[string(o.Status)] {
			openOrders = append(openOrders, struct {
				OrderID        string
				Symbol         string
				Side           string
				Quantity       int
				Price          float64
				FilledQuantity int
				Status         string
			}{
				OrderID:        o.OrderID,
				Symbol:         o.Symbol,
				Side:           string(o.Side),
				Quantity:       o.Quantity,
				Price:          o.Price,
				FilledQuantity: o.FilledQuantity,
				Status:         string(o.Status),
			})
		}
	}

	fmt.Printf("Tổng lệnh: %d | Lệnh đang mở: %d\n\n", len(orders), len(openOrders))

	if len(openOrders) == 0 {
		fmt.Println("Không có lệnh nào đang mở để hủy.")

		// --- Response Summary ---
		fmt.Println("\n[Response] open_count|cancel_status")
		fmt.Println("0|N/A")
		return
	}

	for _, o := range openOrders {
		remaining := o.Quantity - o.FilledQuantity
		fmt.Printf("  OrderID: %s | %s %s | SL: %d @ %.0f | Khớp: %d | Còn: %d | Status: %s\n",
			o.OrderID, o.Symbol, o.Side, o.Quantity, o.Price,
			o.FilledQuantity, remaining, o.Status)
	}

	// --- Bước 2: Hủy lệnh đầu tiên trong danh sách ---
	target := openOrders[0]
	fmt.Printf("\n--- Hủy lệnh: %s ---\n", target.OrderID)

	result, err := t.Trading.CancelOrderByOrderID(accountNo, target.OrderID)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Kết quả hủy: %v\n", result)

	// --- Bước 3: Xác nhận trạng thái sau hủy ---
	fmt.Println("\n--- Kiểm tra sổ lệnh sau hủy ---")
	ordersAfter, err := t.Portfolio.GetTodayOrders(accountNo)
	if err != nil {
		log.Fatal(err)
	}
	for _, o := range ordersAfter {
		if o.OrderID == target.OrderID {
			fmt.Printf("  OrderID: %s | Status: %s | Khớp: %d | Đã hủy: %d\n",
				o.OrderID, o.Status, o.FilledQuantity, o.CancelQuantity)
			break
		}
	}

	// --- Response Summary ---
	fmt.Println("\n[Response] open_count|cancel_status")
	fmt.Printf("%d|%s\n", len(openOrders), result.Status)
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
