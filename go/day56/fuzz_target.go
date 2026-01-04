package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ParseEmail は簡易的なメールアドレスパーサー
// fuzz testingで脆弱性を見つけるための例として使用
func ParseEmail(email string) (username, domain string, err error) {
	// 空文字チェック
	if email == "" {
		return "", "", errors.New("email cannot be empty")
	}

	// 長さチェック（254文字がRFC上限）
	if len(email) > 254 {
		return "", "", errors.New("email too long")
	}

	// @で分割
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "", "", errors.New("email must contain exactly one @")
	}

	username = parts[0]
	domain = parts[1]

	// ユーザー名チェック
	if username == "" {
		return "", "", errors.New("username cannot be empty")
	}

	// ドメインチェック
	if domain == "" {
		return "", "", errors.New("domain cannot be empty")
	}

	// ドメインに少なくとも1つのドットが必要
	if !strings.Contains(domain, ".") {
		return "", "", errors.New("domain must contain at least one dot")
	}

	return username, domain, nil
}

// SanitizeInput は入力文字列をサニタイズする
// fuzz testingでパニックやメモリ問題を見つけるための例
func SanitizeInput(input string) (string, error) {
	// UTF-8の妥当性チェック
	if !utf8.ValidString(input) {
		return "", errors.New("invalid UTF-8 string")
	}

	// 制御文字を除去
	var result strings.Builder
	for _, r := range input {
		// 制御文字（0x00-0x1F, 0x7F-0x9F）をスキップ
		if r >= 0x20 && r < 0x7F || r >= 0xA0 {
			result.WriteRune(r)
		}
	}

	return result.String(), nil
}

// CalculateDiscount は割引計算を行う
// fuzz testingで算術エラー（オーバーフロー、ゼロ除算など）を見つけるための例
func CalculateDiscount(price int, discountPercent int) (int, error) {
	if price < 0 {
		return 0, errors.New("price cannot be negative")
	}

	if discountPercent < 0 || discountPercent > 100 {
		return 0, errors.New("discount percent must be between 0 and 100")
	}

	// オーバーフロー対策
	if price > 1000000000 {
		return 0, errors.New("price too large")
	}

	discount := (price * discountPercent) / 100
	finalPrice := price - discount

	return finalPrice, nil
}

// ParseUserAge はユーザーの年齢文字列をパースする
// fuzz testingで変換エラーを見つけるための例
func ParseUserAge(ageStr string) (int, error) {
	// 空文字チェック
	ageStr = strings.TrimSpace(ageStr)
	if ageStr == "" {
		return 0, errors.New("age cannot be empty")
	}

	// 数値変換
	age, err := strconv.Atoi(ageStr)
	if err != nil {
		return 0, fmt.Errorf("invalid age format: %w", err)
	}

	// 範囲チェック
	if age < 0 {
		return 0, errors.New("age cannot be negative")
	}

	if age > 150 {
		return 0, errors.New("age too large (must be <= 150)")
	}

	return age, nil
}

func main() {
	// 正常系の例
	username, domain, err := ParseEmail("user@example.com")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Email parsed: username=%s, domain=%s\n", username, domain)
	}

	// SanitizeInputの例
	sanitized, err := SanitizeInput("Hello\x00World\nTest")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Sanitized: %s\n", sanitized)
	}

	// CalculateDiscountの例
	finalPrice, err := CalculateDiscount(1000, 20)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Final price: %d\n", finalPrice)
	}

	// ParseUserAgeの例
	age, err := ParseUserAge("25")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Age: %d\n", age)
	}
}
