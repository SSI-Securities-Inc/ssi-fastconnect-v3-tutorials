/*
Sample 19 — Sự kiện doanh nghiệp (Corporate actions)
=====================================================
Xem cổ tức, cổ phiếu thưởng, quyền mua... của các mã. Chỉ ĐỌC dữ liệu, không đặt lệnh.

Luồng:
 1. GetCorporateActions(): sự kiện của HÔM NAY
 2. GetCorporateActionsHistorical(from, to): sự kiện trong một khoảng ngày ở quá khứ
    (server cho tối đa 5 ngày mỗi lần gọi)
 3. Lọc theo mã (WithCorporateActionsSymbol) và ngôn ngữ (WithCorporateActionsLanguage "vn"/"en")
 4. Khoảng ngày quá rộng bị SDK chặn tại client (ValidationError), không gọi server

Lưu ý:
  - SDK tự lấy hết các trang (pageSize 100), bạn không cần tự phân trang.
  - Mọi trường là chuỗi đúng như server gửi. Ngày có dạng YYYY/MM/DD, Ratio KHÔNG được parse
    vì ý nghĩa phụ thuộc loại sự kiện. Trường server để trống là "".
  - Ngày truyền vào phải đúng dạng YYYY/MM/DD (không dùng dấu '-').
*/
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ssisdk "github.com/SSI-Securities-Inc/ssi-sdk-go/v3"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/market"
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

const dateFmt = "2006/01/02"

func dash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

func show(actions []market.CorporateAction, limit int) {
	for i, a := range actions {
		if i >= limit {
			fmt.Printf("  ... còn %d sự kiện\n", len(actions)-limit)
			break
		}
		fmt.Printf("  %-8s | %-6s | GDKHQ: %-10s | Chốt: %-10s | Phát hành: %-10s | Tỷ lệ: %-8s | %s\n",
			a.Symbol, a.Exchange, dash(a.ExrightDate), dash(a.RecordDate), dash(a.IssueDate), dash(a.Ratio), a.EventTitle)
	}
}

func main() {
	config, _, otp := loadConfig()

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, otp)

	data := ssi.NewData(auth)

	// --- Bước 1: Sự kiện của hôm nay ---
	fmt.Println("--- Sự kiện doanh nghiệp hôm nay ---")
	todayActions, err := data.MarketData.GetCorporateActions()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tổng số sự kiện: %d\n", len(todayActions))
	show(todayActions, 10)

	// --- Bước 2: Sự kiện trong 5 ngày gần nhất (tối đa 5 ngày cho mỗi lần gọi) ---
	now := time.Now()
	fromDate := now.AddDate(0, 0, -4).Format(dateFmt)
	toDate := now.Format(dateFmt)
	fmt.Printf("\n--- Sự kiện từ %s đến %s ---\n", fromDate, toDate)
	rangeActions, err := data.MarketData.GetCorporateActionsHistorical(fromDate, toDate)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tổng số sự kiện: %d\n", len(rangeActions))
	show(rangeActions, 10)

	// --- Bước 3: Lọc theo mã, trả tiêu đề bằng tiếng Anh ---
	fmt.Println("\n--- Sự kiện của mã SSI (tiếng Anh) ---")
	ssiActions, err := data.MarketData.GetCorporateActionsHistorical(fromDate, toDate,
		market.WithCorporateActionsSymbol("SSI"),
		market.WithCorporateActionsLanguage("en"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Tổng số sự kiện: %d\n", len(ssiActions))
	show(ssiActions, 10)

	// --- Bước 4: Khoảng ngày quá rộng bị chặn tại SDK ---
	fmt.Println("\n--- Khoảng 30 ngày (vượt giới hạn 5 ngày) ---")
	_, err = data.MarketData.GetCorporateActionsHistorical(now.AddDate(0, 0, -30).Format(dateFmt), toDate)
	var vErr *ssisdk.ValidationError
	if errors.As(err, &vErr) {
		fmt.Printf("  SDK từ chối trước khi gọi server: %v\n", err)
		fmt.Println("  -> chia khoảng ngày thành các đoạn <= 5 ngày rồi gọi từng đoạn.")
	} else if err != nil {
		log.Fatal(err)
	}

	// --- Response Summary ---
	fmt.Println("\n[Response] today_count|range_count|ssi_range_count")
	fmt.Printf("%d|%d|%d\n", len(todayActions), len(rangeActions), len(ssiActions))
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
