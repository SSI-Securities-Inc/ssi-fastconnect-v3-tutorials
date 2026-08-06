/*
Sample 02 — Yêu cầu & Xác thực OTP (Request OTP & Verify OTP / Smart OTP)
========================================================================
Hướng dẫn luồng Request OTP và Polling Smart OTP Push Notification.
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/SSI-Securities-Inc/ssi-sdk-go/v3/ssi"
)

func main() {
	config, err := GetConfig()
	if err != nil {
		log.Fatalf("Không đọc được cấu hình: %v", err)
	}

	auth := ssi.NewAuth(config)
	defer auth.Close()

	fmt.Println("=== Bước 1: Yêu cầu OTP (Request OTP) ===")
	otpRes, err := auth.RequestOTP()
	if err != nil {
		log.Fatalf("Lỗi Request OTP: %v", err)
	}
	fmt.Printf("Request OTP Response: %+v\n", otpRes)

	var transactionID string
	if dataMap, ok := otpRes["data"].(map[string]interface{}); ok {
		if tid, ok := dataMap["transactionId"].(string); ok {
			transactionID = tid
		}
	}
	if transactionID == "" {
		if tid, ok := otpRes["transactionId"].(string); ok {
			transactionID = tid
		}
	}

	if transactionID != "" {
		fmt.Printf("\n[Smart OTP] Đã nhận Transaction ID: %s\n", transactionID)
		fmt.Println("Vui lòng mở ứng dụng SSI trên điện thoại và bấm APPROVE (Xác nhận)...")
		fmt.Println("SDK đang Polling chờ bạn bấm phê duyệt...")

		accessToken, err := auth.EnsureAuthenticated("", transactionID)
		if err != nil {
			log.Fatalf("\n[LỖI/TIMEOUT] Phê duyệt Smart OTP thất bại: %v", err)
		}

		fmt.Println("\n[THÀNH CÔNG] Đã xác thực Smart OTP!")
		fmt.Printf("Access Token: %s...\n", accessToken[:minInt(40, len(accessToken))])
	} else {
		fmt.Print("\n[OTP Thường / Smart OTP lấy trực tiếp trên App] Nhập mã OTP 6 số: ")
		reader := bufio.NewReader(os.Stdin)
		userOTP, _ := reader.ReadString('\n')
		userOTP = strings.TrimSpace(userOTP)

		if userOTP != "" {
			token, err := auth.Authenticate(userOTP)
			if err != nil {
				log.Fatalf("Lỗi xác thực OTP: %v", err)
			}
			fmt.Println("\n[THÀNH CÔNG] Đã xác thực mã OTP!")
			fmt.Printf("Access Token: %s...\n", token.AccessToken[:minInt(40, len(token.AccessToken))])
		}
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
