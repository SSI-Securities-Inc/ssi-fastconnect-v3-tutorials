/*
Sample 18 — Xử lý lỗi của SDK (error handling)
===============================================
Không gửi lệnh thật. Phần lớn là mô phỏng offline bằng các hàm tạo lỗi của SDK
(ssisdk.New...Error), chỉ có 2 lời gọi "thật" mà SDK chặn ngay tại client (không gọi mạng).

Luồng:
 1. classify: phân loại lỗi bằng errors.As. THỨ TỰ QUAN TRỌNG vì các lỗi con Unwrap về lỗi cha
    (DuplicateRequestError -> APIError -> SSIError, SmartOTP* -> AuthenticationError, ...),
    nên phải kiểm tra loại cụ thể trước loại chung
 2. RateLimitError: RetryAfterMs (mili-giây) và retry có giới hạn số lần
 3. ReauthenticationRequiredError / SmartOTPPendingError / SmartOTPRejectedError: khi nào
    chờ, khi nào phải đăng nhập lại
 4. DuplicateRequestError (409): KHÔNG tự đặt lại lệnh, đối chiếu sổ lệnh
 5. ValidationError: dữ liệu sai, bị chặn tại SDK
 6. Mã lỗi API (APIError.Code, chuỗi) và ssisdk.ServerErrorCode (số, có LookupServerErrorCode)
 7. Timeout tính bằng mili-giây (Config.TimeoutMs, OTPPollIntervalMs, OTPPollMaxWaitMs, Stream.Wait)
 8. Không bao giờ in ResponseBody, Headers, token, OTP hay số tài khoản đầy đủ ra log
*/
package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	ssisdk "github.com/SSI-Securities-Inc/ssi-sdk-go/v3"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/auth"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/trading"
)

func main() {
	config, app := loadConfig()
	mustHave("equity_account", app.EquityAccount)
	accountNo := app.EquityAccount

	a := ssi.NewAuth(config)
	defer a.Close()
	ensureAuth(a, app.OTP)

	handled := 0

	// --- Bước 1: Phân loại các loại lỗi (mô phỏng offline) ---
	fmt.Println("\n--- 1. Phân loại lỗi ---")
	retryAfter := 1500
	samples := []error{
		ssisdk.NewRateLimitError("Quá nhiều request", &retryAfter),
		ssisdk.NewSmartOTPPendingError("Chưa xác nhận trên Smart OTP", "401114", 401, nil),
		ssisdk.NewSmartOTPRejectedError("Smart OTP bị từ chối", "401115", 401, nil),
		ssisdk.NewReauthenticationRequiredError("Phiên hết hạn", "401", 401, nil),
		ssisdk.NewDuplicateRequestError("Trùng ClientRequestID", "409", nil),
		ssisdk.NewValidationError("quantity phải > 0"),
		ssisdk.NewAPIError("Sai chữ ký", "401013", 401, nil),
		ssisdk.NewAPIError("Lỗi lạ", "999999", 400, nil),
		ssisdk.NewTradingWSError(429, "TOO_MANY", "vượt hạn mức"),
		fmt.Errorf("gọi đặt lệnh: %w", ssisdk.NewRateLimitError("Bọc thêm một lớp", nil)),
	}
	for _, err := range samples {
		fmt.Printf("  %s\n", classify(err))
		handled++
	}

	// --- Bước 2: Retry khi bị rate limit, tôn trọng RetryAfterMs ---
	fmt.Println("\n--- 2. Retry khi bị rate limit ---")
	attempts := 0
	err := retryOnRateLimit(3, func() error {
		attempts++
		if attempts < 3 {
			ms := 200 * attempts
			return ssisdk.NewRateLimitError("Quá nhiều request", &ms)
		}
		return nil
	})
	fmt.Printf("  Kết quả sau %d lần gọi: %v\n", attempts, err)
	handled++

	// --- Bước 3: Smart OTP và đăng nhập lại ---
	fmt.Println("\n--- 3. Smart OTP / đăng nhập lại ---")
	fmt.Printf("  Poll Smart OTP mỗi %d ms, tối đa %d ms (Config.OTPPollIntervalMs / OTPPollMaxWaitMs)\n",
		config.OTPPollIntervalMs, config.OTPPollMaxWaitMs)
	fmt.Println("  Chờ thủ công: auth.EnsureOptions{TransactionID, PollIntervalMs, PollMaxRetries} với EnsureAuthenticatedWith")
	handled++

	// --- Bước 4 & 5: Lỗi thật mà SDK chặn tại client, không gọi mạng ---
	fmt.Println("\n--- 4-5. Lỗi dữ liệu thật (SDK chặn trước khi gửi) ---")
	t := ssi.NewTrading(a)
	fmt.Printf("  Đặt lệnh LO giá 0 trên tài khoản %s:\n", maskAccount(accountNo))
	_, err = t.Trading.PlaceLimitOrder(accountNo, "SSI", trading.OrderSideBuy, 100, 0)
	fmt.Printf("    %s\n", classify(err))
	if err == nil {
		log.Fatal("SDK không chặn lệnh giá 0, dừng để kiểm tra")
	}
	handled++

	d := ssi.NewData(a)
	fmt.Println("  Nến tuần (server không phục vụ):")
	_, err = d.MarketData.GetOHLC1WeekHistorical("SSI", "2026/03/01 00:00:00", "2026/03/27 23:59:59", 1, 100)
	fmt.Printf("    %s\n", classify(err))
	handled++

	// --- Bước 6: Mã lỗi API và ServerErrorCode ---
	fmt.Println("\n--- 6. Mã lỗi server ---")
	for _, code := range []string{"401013", "401114", "999999", "abc"} {
		fmt.Printf("  Code %q -> %s\n", code, describeServerCode(code))
	}
	handled++

	// --- Bước 7: Timeout tính bằng mili-giây ---
	fmt.Println("\n--- 7. Timeout ---")
	fmt.Printf("  Config.TimeoutMs=%d | MaxRetries=%d | RetryDelayMs=%d (gấp đôi mỗi lần) | RateLimitPerSecond=%d\n",
		config.TimeoutMs, config.MaxRetries, config.RetryDelayMs, config.RateLimitPerSecond)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	<-ctx.Done()
	fmt.Printf("  context quá hạn: errors.Is(DeadlineExceeded)=%v\n", errors.Is(ctx.Err(), context.DeadlineExceeded))
	fmt.Println("  Stream.Wait nhận *int mili-giây; lệnh ghi bị timeout KHÔNG được gửi lại, hãy đối chiếu sổ lệnh")
	handled++

	// --- Response Summary ---
	fmt.Println("\n[Response] scenarios_handled")
	fmt.Printf("%d\n", handled)
}

// classify trả về mô tả ngắn của lỗi. Chỉ dùng Message/Code/StatusCode; không bao giờ in
// ResponseBody hay Headers vì có thể chứa dữ liệu nhạy cảm.
func classify(err error) string {
	if err == nil {
		return "không có lỗi"
	}
	var dup *ssisdk.DuplicateRequestError
	var pending *ssisdk.SmartOTPPendingError
	var rejected *ssisdk.SmartOTPRejectedError
	var reauth *ssisdk.ReauthenticationRequiredError
	var rl *ssisdk.RateLimitError
	var vErr *ssisdk.ValidationError
	var wsErr *ssisdk.TradingWSError
	var apiErr *ssisdk.APIError
	var authErr *ssisdk.AuthenticationError
	switch {
	case errors.As(err, &dup):
		return "TRÙNG yêu cầu (409): không tự đặt lại, đối chiếu sổ lệnh | " + dup.Message
	case errors.As(err, &pending):
		return "Smart OTP CHỜ xác nhận: tiếp tục poll với TransactionID | " + pending.Message
	case errors.As(err, &rejected):
		return "Smart OTP BỊ TỪ CHỐI: dừng, hỏi người dùng có muốn thử lại | " + rejected.Message
	case errors.As(err, &reauth):
		return "CẦN ĐĂNG NHẬP LẠI bằng OTP mới | " + reauth.Message
	case errors.As(err, &rl):
		wait := "không rõ"
		if rl.RetryAfterMs != nil {
			wait = fmt.Sprintf("%d ms", *rl.RetryAfterMs)
		}
		return fmt.Sprintf("RATE LIMIT: chờ %s rồi thử lại | %s", wait, rl.Message)
	case errors.As(err, &vErr):
		return "DỮ LIỆU KHÔNG HỢP LỆ (chưa gọi server): sửa đầu vào | " + vErr.Message
	case errors.As(err, &wsErr):
		return fmt.Sprintf("TRADING WS: status=%d code=%s | %s", wsErr.Status, wsErr.ErrorCode, wsErr.ServerMessage)
	case errors.As(err, &apiErr):
		return fmt.Sprintf("API (HTTP %d, code %s: %s) | %s", apiErr.StatusCode, apiErr.Code, describeServerCode(apiErr.Code), apiErr.Message)
	case errors.As(err, &authErr):
		return "XÁC THỰC thất bại | " + authErr.Message
	case errors.Is(err, context.DeadlineExceeded):
		return "HẾT THỜI GIAN chờ: chưa rõ server đã xử lý hay chưa"
	default:
		return "lỗi khác: " + err.Error()
	}
}

// describeServerCode đổi mã lỗi dạng chuỗi của API sang ssisdk.ServerErrorCode nếu SDK biết mã đó.
func describeServerCode(code string) string {
	n, err := strconv.Atoi(strings.TrimSpace(code))
	if err != nil {
		return "không phải mã số"
	}
	sc, ok := ssisdk.LookupServerErrorCode(n)
	if !ok || !sc.IsKnown() {
		return fmt.Sprintf("mã %d chưa có trong SDK", n)
	}
	return fmt.Sprintf("ServerErrorCode %d (đã biết)", int(sc))
}

// retryOnRateLimit gọi fn tối đa maxAttempts lần; chỉ retry khi gặp RateLimitError và chờ đúng
// RetryAfterMs (mặc định 1000 ms nếu server không nói). Lỗi khác trả về ngay.
func retryOnRateLimit(maxAttempts int, fn func() error) error {
	var err error
	for i := 1; i <= maxAttempts; i++ {
		if err = fn(); err == nil {
			return nil
		}
		var rl *ssisdk.RateLimitError
		if !errors.As(err, &rl) || i == maxAttempts {
			return err
		}
		waitMs := 1000
		if rl.RetryAfterMs != nil {
			waitMs = *rl.RetryAfterMs
		}
		fmt.Printf("  Lần %d bị rate limit, chờ %d ms...\n", i, waitMs)
		time.Sleep(time.Duration(waitMs) * time.Millisecond)
	}
	return err
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
