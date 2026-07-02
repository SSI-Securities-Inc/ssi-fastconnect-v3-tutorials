/*
Sample 2 — Lấy danh sách chỉ số thị trường (Index)
=====================================================
Hiển thị VN-Index, HNX-Index… trên dashboard.

Luồng:
 1. Client gọi endpoint indexList với filter (exchange, limit, view)
 2. API trả về danh sách chỉ số và dữ liệu giá hiện tại
 3. Map dữ liệu sang model hiển thị UI
 4. Lưu cursor để tải trang kế tiếp khi cần
 5. Cache ngắn hạn để giảm số lần gọi API
*/
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/market"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
)

func main() {
	config := ssi.NewConfig("<CLIENT_ID>")
	config.APIKey = "<API_KEY>"
	config.APISecret = "<API_SECRET>"
	config.PrivateKey = "<PRIVATE_KEY_CONTENT>"

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, "")

	data := ssi.NewData(auth)

	// --- Bước 1: Lấy toàn bộ chỉ số ---
	allIndexes, err := data.MarketData.GetIndexes()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v\n", allIndexes)
	fmt.Printf("Tổng số chỉ số: %d\n\n", len(allIndexes))

	for _, idx := range allIndexes {
		fmt.Printf("  %-15s | %-30s\n", idx.Index, idx.IndexName)
	}

	// --- Bước 2: Lọc chỉ số theo sàn HOSE ---
	fmt.Println("\n--- Chỉ số sàn HOSE ---")
	hoseIndexes, err := data.MarketData.GetIndexesByBoard(market.BoardHOSE)
	if err != nil {
		log.Fatal(err)
	}
	for _, idx := range hoseIndexes {
		fmt.Printf("  %-15s | %s\n", idx.Index, idx.IndexName)
	}

	// --- Bước 3: Lấy chi tiết summary cho một chỉ số cụ thể ---
	fmt.Println("\n--- VN-Index Summary ---")
	summary, err := data.MarketData.GetIndexSummary("VNINDEX")
	if err != nil {
		fmt.Printf("  (Không có dữ liệu summary — có thể ngoài giờ giao dịch)\n")
	} else {
		fmt.Printf("  Giá trị Index    : %f\n", summary.IndexValue)
		fmt.Printf("  Thay đổi         : %+.2f\n", summary.IndexChange)
	}

	// --- Response Summary ---
	fmt.Println("\n[Response] total_indexes|hose_indexes|first_index")
	first := ""
	if len(allIndexes) > 0 {
		first = allIndexes[0].Index
	}
	fmt.Printf("%d|%d|%s\n", len(allIndexes), len(hoseIndexes), first)
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
