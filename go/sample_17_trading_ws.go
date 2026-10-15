/*
Sample 17 — Trading WebSocket (THỬ NGHIỆM / EXPERIMENTAL)
==========================================================
!!! Package experimental: API có thể thay đổi hoặc bị bỏ mà không báo trước giữa các bản SDK.
!!! Không dùng cho production. Luồng đặt lệnh chính thức vẫn là REST (sample 06, 07, 14).

CẢNH BÁO: sample này GỬI LỆNH THẬT qua WebSocket (mua SSI ở giá sàn, lệnh LO nằm chờ) rồi
hủy ngay. Chạy trên môi trường/tài khoản bạn được phép giao dịch. Sample hỏi xác nhận trước khi gửi.
Cần private_key trong config.json: mỗi lệnh ghi (mutating) được ký RSA.

Luồng:
 1. experimental.TradingWS: mở kết nối (token lấy từ Auth, tự lấy mới mỗi lần Connect);
    luôn Close khi xong
 2. SetOnEvent + SubscribeOrderEvents: nhận orderEvent / orderMatchEvent dạng frame
    {"channel","topic","data"}
 3. PlaceOrder: lệnh LO mua ở giá sàn (xa thị trường, nằm chờ) với ClientRequestID tự đặt
 4. QueryPosition, QueryOrderBook: truy vấn qua WebSocket (không ký, chỉ đọc)
 5. CancelOrder theo OrderID
 6. Rate limit: WithRespectRateLimit để client tự chờ; RateLimit() xem hạn mức còn lại;
    lỗi 429 là RateLimitError hoặc TradingWSError (Status, ErrorCode, ServerMessage)
 7. Lệnh ghi bị timeout KHÔNG bao giờ được gửi lại tự động (lệnh gốc có thể đã vào sàn):
    đối chiếu bằng sự kiện hoặc QueryOrderBook theo ClientRequestID
*/
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	ssisdk "github.com/SSI-Securities-Inc/ssi-sdk-go/v3"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/experimental"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/trading"
)

const (
	wsSymbol = "SSI"
	wsQty    = 100
)

func main() {
	fmt.Println("*** TRADING WEBSOCKET — THỬ NGHIỆM (experimental), API có thể đổi bất cứ lúc nào ***")

	config, app := loadConfig()
	mustHave("private_key", app.PrivateKey) // lệnh ghi qua WebSocket được ký RSA
	mustHave("equity_account", app.EquityAccount)
	accountNo := app.EquityAccount

	a := ssi.NewAuth(config)
	defer a.Close()
	ensureAuth(a, app.OTP)

	floor := floorPrice(ssi.NewData(a), wsSymbol)
	fmt.Printf("Tài khoản %s | %s giá sàn %.0f | KL %d\n", maskAccount(accountNo), wsSymbol, floor, wsQty)
	if !confirmRealOrders(fmt.Sprintf("sẽ gửi 1 lệnh MUA THẬT %s x %d @ %.0f (giá sàn) qua Trading WebSocket", wsSymbol, wsQty, floor)) {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// --- Bước 1: Kết nối. Lỗi từ chối lúc bắt tay (thiếu scope, chữ ký sai...) trả về ở đây. ---
	fmt.Println("\nĐang mở Trading WebSocket...")
	ws, err := experimental.TradingWS(ctx, a,
		experimental.WithRequestTimeoutMs(15000),
		experimental.WithRespectRateLimit(true),
		experimental.WithMaxRateLimitWaitMs(5000),
	)
	if err != nil {
		reportWSError("TradingWS", err)
		return
	}
	defer ws.Close()
	fmt.Printf("IsConnected=%v\n", ws.Client().IsConnected())

	// --- Bước 2: Sự kiện lệnh. Không in topic vì topic chứa số tài khoản. ---
	events := make(chan string, 32)
	ws.Client().SetOnEvent(func(frame map[string]interface{}) {
		data, _ := frame["data"].(map[string]interface{})
		line := fmt.Sprintf("channel=%v eventType=%v", frame["channel"], data["eventType"])
		if st, ok := data["status"]; ok {
			line += fmt.Sprintf(" status=%v", st)
		}
		fmt.Printf("  [EVENT] %s\n", line)
		select {
		case events <- line:
		default:
		}
	})
	if err := ws.SubscribeOrderEvents(accountNo); err != nil {
		reportWSError("SubscribeOrderEvents", err)
		return
	}
	defer ws.UnsubscribeAllEvents()

	// --- Bước 3: Đặt lệnh LO ở giá sàn ---
	clientRequestID := "W" + time.Now().Format("0102150405")
	fmt.Printf("\n--- PlaceOrder (ClientRequestID %s) ---\n", clientRequestID)
	placed, err := ws.PlaceOrder(ctx, accountNo, wsSymbol, trading.OrderSideBuy, wsQty, floor, trading.OrderTypeLO, clientRequestID)
	if err != nil {
		reportWSError("PlaceOrder", err)
		fmt.Println("  Không gửi lại tự động. Kiểm tra QueryOrderBook theo ClientRequestID trước khi đặt lại.")
		queryOrderBook(ctx, ws, accountNo)
		return
	}
	fmt.Printf("  OrderID: %s | Status: %s\n", placed.OrderID, placed.Status)

	select {
	case <-events:
	case <-time.After(5 * time.Second):
		fmt.Println("  (chưa thấy sự kiện lệnh trong 5 giây, vẫn tiếp tục)")
	}

	// --- Bước 4: Truy vấn qua WebSocket ---
	fmt.Println("\n--- QueryPosition ---")
	pos, err := ws.QueryPosition(ctx, accountNo, false)
	if err != nil {
		reportWSError("QueryPosition", err)
	} else {
		fmt.Printf("  Status: %d | %s\n", pos.Status, describeResult(pos.Result))
	}
	orderRows := queryOrderBook(ctx, ws, accountNo)

	// --- Bước 5: Hủy lệnh ---
	fmt.Println("\n--- CancelOrder ---")
	cancelled := false
	if placed.OrderID != "" {
		res, err := ws.CancelOrder(ctx, accountNo, placed.OrderID, "")
		if err != nil {
			reportWSError("CancelOrder", err)
			fmt.Println("  Kiểm tra sổ lệnh và hủy thủ công nếu lệnh còn mở.")
		} else {
			cancelled = true
			fmt.Printf("  OrderID: %s | Status: %s\n", res.OrderID, res.Status)
		}
	}

	select {
	case <-events:
	case <-time.After(5 * time.Second):
	}

	// --- Bước 6: Hạn mức rate limit mà server báo gần nhất ---
	fmt.Printf("\nRate limit gần nhất: %s\n", formatRateLimit(ws.Client().RateLimit()))
	if rej := ws.Client().Rejection(); rej != nil {
		fmt.Printf("Server từ chối kết nối: status=%d code=%s %s\n", rej.Status, rej.ErrorCode, rej.ServerMessage)
	}

	// --- Response Summary ---
	fmt.Println("\n[Response] placed|cancelled|order_book_rows")
	fmt.Printf("%v|%v|%d\n", placed.OrderID != "", cancelled, orderRows)
}

// queryOrderBook truy vấn sổ lệnh trong ngày của mã đang thử, trả về số dòng (-1 nếu không đếm được).
func queryOrderBook(ctx context.Context, ws *experimental.Service, accountNo string) int {
	fmt.Println("\n--- QueryOrderBook ---")
	today := time.Now().Format("2006/01/02")
	res, err := ws.QueryOrderBook(ctx, experimental.OrderBookQuery{
		AccountNo: accountNo,
		FromDate:  today,
		ToDate:    today,
		Symbol:    wsSymbol,
		PageIndex: 1,
		PageSize:  20,
	})
	if err != nil {
		reportWSError("QueryOrderBook", err)
		return -1
	}
	fmt.Printf("  Status: %d | %s\n", res.Status, describeResult(res.Result))
	if rows, ok := res.Result.([]interface{}); ok {
		return len(rows)
	}
	return -1
}

// describeResult chỉ mô tả hình dạng kết quả, không in nội dung (có thể chứa số tài khoản).
func describeResult(v interface{}) string {
	switch r := v.(type) {
	case []interface{}:
		return fmt.Sprintf("danh sách %d phần tử", len(r))
	case map[string]interface{}:
		return fmt.Sprintf("đối tượng %d trường", len(r))
	case nil:
		return "không có dữ liệu"
	default:
		return fmt.Sprintf("kiểu %T", v)
	}
}

func formatRateLimit(rl map[string]interface{}) string {
	if len(rl) == 0 {
		return "(server chưa báo)"
	}
	keys := make([]string, 0, len(rl))
	for k := range rl {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", k, rl[k]))
	}
	return strings.Join(parts, " ")
}

// reportWSError phân loại lỗi bằng errors.As, chỉ in thông điệp, không in frame hay token.
func reportWSError(op string, err error) {
	var rl *ssisdk.RateLimitError
	var wsErr *ssisdk.TradingWSError
	var vErr *ssisdk.ValidationError
	switch {
	case errors.As(err, &rl):
		wait := "không rõ"
		if rl.RetryAfterMs != nil {
			wait = fmt.Sprintf("%d ms", *rl.RetryAfterMs)
		}
		fmt.Printf("  %s: bị giới hạn tốc độ, thử lại sau %s\n", op, wait)
	case errors.As(err, &wsErr):
		fmt.Printf("  %s: server từ chối status=%d code=%s: %s\n", op, wsErr.Status, wsErr.ErrorCode, wsErr.ServerMessage)
	case errors.As(err, &vErr):
		fmt.Printf("  %s: dữ liệu không hợp lệ (chưa gửi): %v\n", op, vErr)
	case errors.Is(err, context.DeadlineExceeded):
		fmt.Printf("  %s: hết thời gian chờ, chưa rõ lệnh đã vào sàn hay chưa\n", op)
	default:
		fmt.Printf("  %s: lỗi: %v\n", op, err)
	}
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
