/*
Sample 4 — Lấy danh sách cổ phiếu theo sàn
=============================================
Tạo watchlist/screener theo tiêu chí thị trường.

Luồng:
 1. Client gọi securitiesByBoard theo exchange, board, sector
 2. API trả về danh sách mã + thông tin giao dịch cơ bản
 3. Lọc/sắp xếp theo nhu cầu UI
 4. Lưu cursor để phân trang dữ liệu lớn
 5. Khi user chọn mã, chuyển sang luồng xem chi tiết/đặt lệnh
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
	config.LogLevel = "DEBUG"

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, "")

	data := ssi.NewData(auth)

	// --- Bước 1: Lấy danh sách cổ phiếu sàn HOSE ---
	fmt.Println("--- Cổ phiếu sàn HOSE ---")
	hoseSecurities, err := data.MarketData.GetSecuritiesInfoByBoard(market.BoardHOSE)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tổng số mã: %d\n\n", len(hoseSecurities))

	limit := 10
	if len(hoseSecurities) < limit {
		limit = len(hoseSecurities)
	}
	for _, sec := range hoseSecurities[:limit] {
		board := ""
		if sec.Board != nil {
			board = string(*sec.Board)
		}
		fmt.Printf("  %-10s | %s\n", sec.Symbol, board)
	}

	// --- Bước 2: Lấy danh sách cổ phiếu sàn HNX ---
	fmt.Println("\n--- Cổ phiếu sàn HNX ---")
	hnxSecurities, err := data.MarketData.GetSecuritiesInfoByBoard(market.BoardHNX)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tổng số mã: %d\n", len(hnxSecurities))

	limit = 10
	if len(hnxSecurities) < limit {
		limit = len(hnxSecurities)
	}
	for _, sec := range hnxSecurities[:limit] {
		board := ""
		if sec.Board != nil {
			board = string(*sec.Board)
		}
		fmt.Printf("  %-10s | %s\n", sec.Symbol, board)
	}

	// --- Bước 3: Lấy theo chỉ số (index) ---
	fmt.Println("\n--- Cổ phiếu thuộc VN30 ---")
	vn30Securities, err := data.MarketData.GetSecuritiesInfoByIndex("VN30")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tổng số mã: %d\n", len(vn30Securities))

	for _, sec := range vn30Securities {
		fmt.Printf("  %-10s | %s\n", sec.Symbol, sec.SymbolNameVi)
	}

	// --- Bước 4: Xem thông tin chi tiết một mã ---
	fmt.Println("\n--- Chi tiết mã SSI ---")
	info, err := data.MarketData.GetSecuritiesInfo("SSI")
	if err != nil {
		log.Fatal(err)
	}
	board := ""
	if info.Board != nil {
		board = string(*info.Board)
	}
	fmt.Printf("  Mã          : %s\n", info.Symbol)
	fmt.Printf("  Tên (VI)    : %s\n", info.SymbolNameVi)
	fmt.Printf("  Tên (EN)    : %s\n", info.SymbolNameEn)
	fmt.Printf("  Sàn         : %s\n", board)
	fmt.Printf("  Lot size    : %d\n", info.LotSize)
	fmt.Printf("  ICB Code    : %s\n", info.ICBCode)
	fmt.Printf("  Listed Shares: %d\n", info.ListedShares)
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
