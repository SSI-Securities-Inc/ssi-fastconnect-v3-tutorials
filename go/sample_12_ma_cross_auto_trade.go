/*
Sample 11 — MA Cross Signal + Auto Place & Monitor
====================================================
Tự động giao dịch khi MA5 cắt MA10.

Luồng:
 1. Lấy OHLC theo chu kỳ để tính MA(5), MA(10)
 2. Kiểm tra điều kiện giao cắt tại nến hiện tại so với nến trước
 3. Có tín hiệu thì kiểm tra balance/risk rule trước khi vào lệnh
 4. Tự động đặt lệnh (thường MARKET để ưu tiên khớp nhanh)
 5. Theo dõi trạng thái đến khi FILLED, timeout thì hủy lệnh
 6. Ghi log kết quả giao dịch và tính P&L cơ bản
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

const (
	maAccountNo  = "<ACCOUNT_NO>"
	maSymbol     = "SSI"
	maFast       = 5
	maSlow       = 10
	maQuantity   = 100
	maxPoll      = 10
	pollInterval = 3 * time.Second
)

// Trạng thái kết thúc
var maTerminalStatuses = map[string]bool{
	string(trading.OrderStatusFilled):           true,
	string(trading.OrderStatusCancelled):        true,
	string(trading.OrderStatusRejected):         true,
	string(trading.OrderStatusExpired):          true,
	string(trading.OrderStatusPartialCancelled): true,
}

// OHLCBar đại diện cho 1 nến OHLC (dùng nội bộ cho tính MA)
type OHLCBar struct {
	ClosePrice float64
}

// calculateMA tính Moving Average từ danh sách close price.
func calculateMA(closePrices []float64, period int) *float64 {
	if len(closePrices) < period {
		return nil
	}
	sum := 0.0
	for _, p := range closePrices[len(closePrices)-period:] {
		sum += p
	}
	result := sum / float64(period)
	return &result
}

// detectCross phát hiện tín hiệu giao cắt MA.
// Returns: "BUY" (golden cross), "SELL" (death cross), hoặc "".
func detectCross(closePrices []float64, fastPeriod, slowPeriod int) string {
	if len(closePrices) < slowPeriod+1 {
		return ""
	}

	// MA tại nến hiện tại
	maFastNow := calculateMA(closePrices, fastPeriod)
	maSlowNow := calculateMA(closePrices, slowPeriod)

	// MA tại nến trước
	prevPrices := closePrices[:len(closePrices)-1]
	maFastPrev := calculateMA(prevPrices, fastPeriod)
	maSlowPrev := calculateMA(prevPrices, slowPeriod)

	if maFastNow == nil || maSlowNow == nil || maFastPrev == nil || maSlowPrev == nil {
		return ""
	}

	// Golden Cross: MA5 cắt lên MA10
	if *maFastPrev <= *maSlowPrev && *maFastNow > *maSlowNow {
		return "BUY"
	}

	// Death Cross: MA5 cắt xuống MA10
	if *maFastPrev >= *maSlowPrev && *maFastNow < *maSlowNow {
		return "SELL"
	}

	return ""
}

func main() {
	config := ssi.NewConfig("<CLIENT_ID>")
	config.APIKey = "<API_KEY>"
	config.APISecret = "<API_SECRET>"
	config.PrivateKey = "<PRIVATE_KEY_CONTENT>"

	auth := ssi.NewAuth(config)
	defer auth.Close()

	ensureAuth(auth, "<OTP>")

	data := ssi.NewData(auth)
	t := ssi.NewTrading(auth)

	// ===== Bước 1: Lấy dữ liệu OHLC =====
	fmt.Printf("--- Lấy dữ liệu OHLC %s (1 ngày) ---\n", maSymbol)
	bars, err := data.MarketData.GetOHLC1DayHistorical(
		maSymbol, "2026/01/01 00:00:00", "2026/04/22 23:59:59", 1, 100,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Số nến: %d\n", len(bars))

	if len(bars) < maSlow+1 {
		fmt.Printf("  Không đủ dữ liệu (cần ít nhất %d nến). Dừng.\n", maSlow+1)
		return
	}

	// Trích xuất close prices
	closePrices := make([]float64, len(bars))
	for i, bar := range bars {
		closePrices[i] = bar.ClosePrice
	}

	// ===== Bước 2: Tính MA và phát hiện tín hiệu =====
	maFastVal := calculateMA(closePrices, maFast)
	maSlowVal := calculateMA(closePrices, maSlow)
	fmt.Printf("\n  MA%d = %.2f\n", maFast, *maFastVal)
	fmt.Printf("  MA%d = %.2f\n", maSlow, *maSlowVal)

	signal := detectCross(closePrices, maFast, maSlow)
	if signal == "" {
		fmt.Println("  Tín hiệu: Không có tín hiệu giao cắt")
		fmt.Println("\nKhông có tín hiệu, không đặt lệnh.")
		return
	}
	fmt.Printf("  Tín hiệu: %s\n", signal)

	var side trading.OrderSide
	if signal == "BUY" {
		side = trading.OrderSideBuy
	} else {
		side = trading.OrderSideSell
	}

	// ===== Bước 3: Kiểm tra balance/risk =====
	fmt.Println("\n--- Kiểm tra sức mua/bán ---")
	maxBS, err := t.Trading.GetMaxBuySellAtMarketPrice(maAccountNo, maSymbol)
	if err != nil {
		log.Fatal(err)
	}

	var maxQty int
	if signal == "BUY" {
		maxQty = maxBS.MaxBuyQuantity
	} else {
		maxQty = maxBS.MaxSellQuantity
	}
	fmt.Printf("  Max %s: %d cổ phiếu\n", signal, maxQty)

	if maxQty < maQuantity {
		fmt.Printf("  Không đủ (%d cần, %d có). Dừng.\n", maQuantity, maxQty)
		return
	}

	// ===== Bước 4: Đặt lệnh Market =====
	fmt.Printf("\n--- Đặt lệnh MARKET %s %s x%d ---\n", signal, maSymbol, maQuantity)
	result, err := t.Trading.PlaceMarketOrder(
		maAccountNo, maSymbol, side, maQuantity,
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  Kết quả: %v\n", result)

	// ===== Bước 5: Theo dõi đến khi FILLED hoặc timeout =====
	fmt.Println("\n--- Theo dõi trạng thái lệnh ---")
	var filledOrder *struct {
		FilledQuantity int
		AvgPrice       float64
		Status         string
	}

	for i := 1; i <= maxPoll; i++ {
		orders, err := t.Portfolio.GetTodayOrders(maAccountNo)
		if err != nil {
			log.Fatal(err)
		}
		if len(orders) == 0 {
			time.Sleep(pollInterval)
			continue
		}

		latest := orders[len(orders)-1]
		fmt.Printf("  Poll %d: %s | Status=%s | Khớp=%d/%d\n",
			i, latest.OrderID, latest.Status,
			latest.FilledQuantity, latest.Quantity)

		if maTerminalStatuses[string(latest.Status)] {
			filledOrder = &struct {
				FilledQuantity int
				AvgPrice       float64
				Status         string
			}{
				FilledQuantity: latest.FilledQuantity,
				AvgPrice:       latest.AvgPrice,
				Status:         string(latest.Status),
			}
			break
		}

		time.Sleep(pollInterval)
	}

	// Timeout — hủy lệnh nếu chưa kết thúc
	if filledOrder == nil {
		fmt.Println("  Timeout! Đang hủy lệnh...")
		orders, _ := t.Portfolio.GetTodayOrders(maAccountNo)
		if len(orders) > 0 {
			latest := orders[len(orders)-1]
			if !maTerminalStatuses[string(latest.Status)] {
				t.Trading.CancelOrderByOrderID(maAccountNo, latest.OrderID)
				fmt.Printf("  Đã gửi hủy lệnh: %s\n", latest.OrderID)
			}
		}
	}

	// ===== Bước 6: Ghi log P&L =====
	fmt.Println("\n--- Kết quả giao dịch ---")
	if filledOrder != nil && filledOrder.FilledQuantity > 0 {
		cost := float64(filledOrder.FilledQuantity) * filledOrder.AvgPrice
		fmt.Printf("  %s %s: %d CP @ %.0f\n", signal, maSymbol, filledOrder.FilledQuantity, filledOrder.AvgPrice)
		fmt.Printf("  Tổng giá trị: %.0f VND\n", cost)
		fmt.Printf("  Trạng thái: %s\n", filledOrder.Status)
	} else {
		fmt.Println("  Lệnh không khớp hoặc đã bị hủy.")
	}

	// Số dư sau giao dịch
	balance, err := t.Portfolio.GetEquityBalance(maAccountNo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n  Tiền mặt còn: %.0f VND\n", balance.AvailableCash)
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
