/*
Sample 14 — Đặt / hủy lệnh theo lô (batch)
=============================================
Gửi nhiều lệnh trong MỘT request, nhận kết quả từng lệnh.

CẢNH BÁO: sample này GỬI LỆNH THẬT (mua SSI ở giá sàn, lệnh LO nằm chờ) rồi hủy ngay.
Chạy trên môi trường/tài khoản bạn được phép giao dịch. Sample hỏi xác nhận trước khi gửi.

Luồng:
 1. Lấy giá sàn từ master data và kiểm tra sức mua (GetMaxBuySell)
 2. PlaceBatchOrders: 1..Config.MaxBatchOrders lệnh, mỗi lệnh có ClientRequestID riêng
    (tối đa 20 ký tự). Cả lô được validate trước khi gửi: một lệnh sai thì cả lô bị từ chối.
 3. Đọc kết quả từng lệnh (Success / OrderID / Status / ErrorCode) — một lệnh có thể thất bại
    trong khi các lệnh khác vẫn nhận
 4. Idempotency: gửi lại cùng ClientRequestID => DuplicateRequestError (HTTP 409), SDK không
    tự đặt lại lệnh dưới id mới vì lệnh gốc có thể đang sống
 5. CancelBatchOrders: hủy theo ClientRequestID hoặc OrderID (đúng một trong hai cho mỗi lệnh)
 6. Vượt Config.MaxBatchOrders => bị từ chối ngay tại SDK, không gọi mạng
*/
package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ssisdk "github.com/SSI-Securities-Inc/ssi-sdk-go/v3"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/trading"
)

const (
	batchSymbol = "SSI"
	batchQty    = 100
	batchCount  = 3
)

func main() {
	config, app := loadConfig()
	mustHave("private_key", app.PrivateKey) // lệnh được ký RSA, thiếu key thì không gửi được
	mustHave("equity_account", app.EquityAccount)
	accountNo := app.EquityAccount

	a := ssi.NewAuth(config)
	defer a.Close()
	ensureAuth(a, app.OTP)

	d := ssi.NewData(a)
	t := ssi.NewTrading(a)

	// --- Bước 1: Giá sàn (giá hợp lệ theo bước giá) và sức mua ---
	floor := floorPrice(d, batchSymbol)
	qtyTotal := batchQty * batchCount
	maxBS, err := t.Trading.GetMaxBuySell(accountNo, batchSymbol, &floor)
	if err != nil {
		log.Fatalf("Không lấy được sức mua: %v", err)
	}
	fmt.Printf("Tài khoản %s | %s giá sàn %.0f | sức mua tối đa: %d cổ phiếu\n",
		maskAccount(accountNo), batchSymbol, floor, maxBS.MaxBuyQuantity)
	if maxBS.MaxBuyQuantity < qtyTotal {
		fmt.Println("Không đủ sức mua cho cả lô, dừng lại.")
		return
	}

	fmt.Printf("Giới hạn lô của SDK (Config.MaxBatchOrders): %d lệnh\n", config.MaxBatchOrders)
	n := batchCount
	if n > config.MaxBatchOrders {
		n = config.MaxBatchOrders
	}

	if !confirmRealOrders(fmt.Sprintf("sẽ gửi %d lệnh MUA THẬT %s x %d @ %.0f (giá sàn)", n, batchSymbol, batchQty, floor)) {
		return
	}

	// --- Bước 2: Tạo lô với ClientRequestID tự đặt (<= 20 ký tự) ---
	runTag := time.Now().Format("0102150405") // MMDDhhmmss
	items := make([]trading.BatchPlaceOrderItem, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, trading.BatchPlaceOrderItem{
			AccountNo:       accountNo,
			Symbol:          batchSymbol,
			Side:            trading.OrderSideBuy,
			Quantity:        batchQty,
			Price:           floor,
			OrderType:       trading.OrderTypeLO,
			ClientRequestID: fmt.Sprintf("B%s%02d", runTag, i+1),
		})
	}

	fmt.Printf("\n--- PlaceBatchOrders (%d lệnh) ---\n", len(items))
	res, err := t.Trading.PlaceBatchOrders(items)
	if err != nil {
		reportError("PlaceBatchOrders", err)
		return
	}
	fmt.Printf("  BatchRequestID: %s\n", res.BatchRequestID)

	// --- Bước 3: Kết quả từng lệnh (đúng thứ tự request) ---
	succeeded := 0
	for i, r := range res.Results {
		if r.Success {
			succeeded++
			fmt.Printf("  #%d OK   | ClientRequestID: %s | OrderID: %s | Status: %s\n",
				i+1, r.ClientRequestID, r.OrderID, r.Status)
		} else {
			fmt.Printf("  #%d FAIL | ClientRequestID: %s | Mã lỗi: %s | %s\n",
				i+1, r.ClientRequestID, r.ErrorCode, r.ErrorMessage)
		}
	}

	var cancels []trading.BatchCancelOrderItem
	for i, r := range res.Results {
		if !r.Success || i >= len(items) {
			continue
		}
		if i == 0 || r.OrderID == "" {
			cancels = append(cancels, trading.BatchCancelOrderItem{AccountNo: accountNo, ClientRequestID: items[i].ClientRequestID})
		} else {
			cancels = append(cancels, trading.BatchCancelOrderItem{AccountNo: accountNo, OrderID: r.OrderID})
		}
	}

	// --- Bước 4: Idempotency — gửi lại cùng ClientRequestID ---
	if len(res.Results) > 0 && res.Results[0].Success {
		fmt.Printf("\n--- Gửi lại ClientRequestID %s ---\n", items[0].ClientRequestID)
		dup, err := t.Trading.PlaceLimitOrder(accountNo, batchSymbol, trading.OrderSideBuy, batchQty, floor,
			trading.WithClientRequestID(items[0].ClientRequestID))
		var dupErr *ssisdk.DuplicateRequestError
		switch {
		case errors.As(err, &dupErr):
			fmt.Printf("  Bị từ chối đúng như dự kiến (409): %v\n", dupErr)
		case err != nil:
			reportError("PlaceLimitOrder (gửi lại)", err)
		default:
			fmt.Printf("  CẢNH BÁO: server nhận thêm lệnh %s, sẽ hủy luôn.\n", dup.OrderID)
			cancels = append(cancels, trading.BatchCancelOrderItem{AccountNo: accountNo, OrderID: dup.OrderID})
		}
	}

	// --- Bước 5: Hủy cả lô ---
	cancelled := 0
	if len(cancels) > 0 {
		fmt.Printf("\n--- CancelBatchOrders (%d lệnh) ---\n", len(cancels))
		cres, err := t.Trading.CancelBatchOrders(cancels)
		if err != nil {
			reportError("CancelBatchOrders", err)
			fmt.Println("  Kiểm tra sổ lệnh và hủy thủ công các lệnh còn mở.")
		} else {
			for i, r := range cres.Results {
				if r.Success {
					cancelled++
				}
				fmt.Printf("  #%d success=%v | Status: %s | %s %s\n", i+1, r.Success, r.Status, r.ErrorCode, r.ErrorMessage)
			}
		}
	}

	// --- Bước 6: Vượt giới hạn lô bị chặn tại SDK ---
	fmt.Printf("\n--- Lô %d lệnh (vượt giới hạn %d) ---\n", config.MaxBatchOrders+1, config.MaxBatchOrders)
	over := make([]trading.BatchCancelOrderItem, config.MaxBatchOrders+1)
	for i := range over {
		over[i] = trading.BatchCancelOrderItem{AccountNo: accountNo, OrderID: "0"}
	}
	var vErr *ssisdk.ValidationError
	if _, err := t.Trading.CancelBatchOrders(over); errors.As(err, &vErr) {
		fmt.Printf("  Bị từ chối tại SDK (không gọi mạng): %v\n", vErr)
	} else {
		fmt.Printf("  Kết quả bất ngờ: %v\n", err)
	}

	// --- Response Summary ---
	fmt.Println("\n[Response] requested|succeeded|cancelled")
	fmt.Printf("%d|%d|%d\n", len(items), succeeded, cancelled)
}

// floorPrice lấy giá sàn hôm nay từ master data (API này phân trang toàn bộ mã nên hơi nặng).
func floorPrice(d *ssi.Data, symbol string) float64 {
	rows, err := d.MarketData.GetMasterData()
	if err != nil {
		log.Fatalf("Không lấy được master data: %v", err)
	}
	for _, r := range rows {
		if r.Symbol == symbol && r.Floor > 0 {
			return r.Floor
		}
	}
	log.Fatalf("Master data không có giá sàn của %s", symbol)
	return 0
}

// reportError phân loại lỗi bằng errors.As; chỉ in thông điệp, không in payload hay token.
func reportError(op string, err error) {
	var dupErr *ssisdk.DuplicateRequestError
	var vErr *ssisdk.ValidationError
	var rlErr *ssisdk.RateLimitError
	switch {
	case errors.As(err, &dupErr):
		fmt.Printf("  %s: trùng ClientRequestID/BatchRequestID (409): %v\n", op, dupErr)
	case errors.As(err, &vErr):
		fmt.Printf("  %s: dữ liệu không hợp lệ (cả lô bị từ chối, chưa gửi lệnh nào): %v\n", op, vErr)
	case errors.As(err, &rlErr):
		wait := "không rõ"
		if rlErr.RetryAfterMs != nil {
			wait = fmt.Sprintf("%d ms", *rlErr.RetryAfterMs)
		}
		fmt.Printf("  %s: bị giới hạn tốc độ, thử lại sau %s\n", op, wait)
	default:
		fmt.Printf("  %s: lỗi: %v\n", op, err)
	}
}

// confirmRealOrders in cảnh báo và chỉ cho đi tiếp khi người dùng gõ đúng YES.
func confirmRealOrders(what string) bool {
	fmt.Printf("\nCẢNH BÁO: %s.\nGõ YES để tiếp tục: ", what)
	line, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	if strings.TrimSpace(line) != "YES" {
		fmt.Println("Đã hủy, không gửi lệnh nào.")
		return false
	}
	return true
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

// loadConfig đọc config.json và dừng ngay khi thiếu trường bắt buộc (không có giá trị mặc định).
func loadConfig() (*ssi.Config, *appConfig) {
	for _, path := range []string{"../config.json", "config.json"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var app appConfig
		if err := json.Unmarshal(data, &app); err != nil {
			log.Fatalf("%s không phải JSON hợp lệ", path)
		}
		mustHave("client_id", app.ClientID)
		mustHave("api_key", app.APIKey)
		mustHave("api_secret", app.APISecret)

		config := ssi.NewConfig(app.ClientID)
		config.APIKey = app.APIKey
		config.APISecret = app.APISecret
		config.PrivateKey = app.PrivateKey
		if app.LogLevel != "" {
			config.LogLevel = app.LogLevel
		}
		return config, &app
	}
	log.Fatal("Không tìm thấy config.json (../config.json hoặc config.json). Xem ../config.example.json")
	return nil, nil
}

func mustHave(field, value string) {
	if value == "" || strings.HasPrefix(value, "YOUR_") {
		log.Fatalf("Thiếu cấu hình %q trong config.json", field)
	}
}

// maskAccount che số tài khoản khi in ra màn hình/log: giữ 2 ký tự đầu và 2 ký tự cuối.
func maskAccount(acc string) string {
	if len(acc) <= 4 {
		return "****"
	}
	return acc[:2] + strings.Repeat("*", len(acc)-4) + acc[len(acc)-2:]
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

// ensureAuth dùng token cache nếu còn hạn, refresh nếu được, còn không thì đăng nhập bằng OTP.
func ensureAuth(a *ssi.Auth, otp string) {
	if token := loadToken(); token != nil {
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

	mustHave("otp", otp)
	fmt.Println("Không có token hợp lệ, đang authenticate...")
	newToken, err := a.Authenticate(otp)
	if err != nil {
		log.Fatalf("Authenticate thất bại: %v", err)
	}
	saveToken(newToken)
	fmt.Println("Authenticate thành công.")
}
