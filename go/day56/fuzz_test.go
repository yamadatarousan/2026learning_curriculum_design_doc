package main

import (
	"strings"
	"testing"
	"unicode/utf8"
)

// FuzzParseEmail はParseEmail関数のfuzzテスト
// Go 1.18以降のネイティブfuzzingを使用
func FuzzParseEmail(f *testing.F) {
	// Seed corpus: 初期テストケース（既知の入力パターン）
	f.Add("user@example.com")
	f.Add("test@test.co.jp")
	f.Add("")
	f.Add("@")
	f.Add("user@")
	f.Add("@domain.com")
	f.Add("user@@example.com")
	f.Add("user@domain@com")
	f.Add(strings.Repeat("a", 300)) // 長い文字列

	// Fuzz関数: ランダムな入力を受け取ってテストを実行
	f.Fuzz(func(t *testing.T, email string) {
		// ParseEmailを実行
		username, domain, err := ParseEmail(email)

		// クラッシュしないことを確認（panicが起きないか）
		// エラーが返ってもOK、重要なのはpanicしないこと

		// エラーがない場合の不変条件をチェック
		if err == nil {
			// エラーがない場合、usernameとdomainは空でないはず
			if username == "" {
				t.Errorf("username is empty but no error returned for email: %q", email)
			}
			if domain == "" {
				t.Errorf("domain is empty but no error returned for email: %q", email)
			}

			// 再構築したメールアドレスに@が1つだけ含まれるはず
			reconstructed := username + "@" + domain
			if strings.Count(reconstructed, "@") != 1 {
				t.Errorf("reconstructed email %q has invalid @ count", reconstructed)
			}

			// ドメインにドットが含まれているはず
			if !strings.Contains(domain, ".") {
				t.Errorf("domain %q does not contain a dot", domain)
			}
		}
	})
}

// FuzzSanitizeInput はSanitizeInput関数のfuzzテスト
func FuzzSanitizeInput(f *testing.F) {
	// Seed corpus
	f.Add("Hello World")
	f.Add("")
	f.Add("Test\x00String")
	f.Add("Unicode: こんにちは")
	f.Add("\n\r\t")
	f.Add(strings.Repeat("A", 10000)) // 長い文字列

	f.Fuzz(func(t *testing.T, input string) {
		// SanitizeInputを実行
		result, err := SanitizeInput(input)

		// UTF-8として無効な文字列の場合、エラーが返るべき
		if !utf8.ValidString(input) {
			if err == nil {
				t.Errorf("expected error for invalid UTF-8 input, got nil")
			}
			return
		}

		// 有効なUTF-8の場合、結果も有効なUTF-8であるべき
		if err == nil {
			if !utf8.ValidString(result) {
				t.Errorf("sanitized result is not valid UTF-8: %q", result)
			}

			// 結果は元の入力以下の長さであるべき（制御文字を除去するため）
			if len(result) > len(input) {
				t.Errorf("sanitized result is longer than input: input=%d, result=%d", len(input), len(result))
			}
		}
	})
}

// FuzzCalculateDiscount はCalculateDiscount関数のfuzzテスト
func FuzzCalculateDiscount(f *testing.F) {
	// Seed corpus
	f.Add(1000, 10)
	f.Add(0, 0)
	f.Add(100, 100)
	f.Add(-100, 50)
	f.Add(1000, -10)
	f.Add(999999999, 50)

	f.Fuzz(func(t *testing.T, price int, discountPercent int) {
		// CalculateDiscountを実行
		finalPrice, err := CalculateDiscount(price, discountPercent)

		// 負の価格はエラーになるべき
		if price < 0 {
			if err == nil {
				t.Errorf("expected error for negative price %d, got nil", price)
			}
			return
		}

		// 無効な割引率はエラーになるべき
		if discountPercent < 0 || discountPercent > 100 {
			if err == nil {
				t.Errorf("expected error for invalid discount percent %d, got nil", discountPercent)
			}
			return
		}

		// 有効な入力の場合
		if err == nil {
			// 最終価格は0以上であるべき
			if finalPrice < 0 {
				t.Errorf("final price is negative: %d (price=%d, discount=%d%%)", finalPrice, price, discountPercent)
			}

			// 最終価格は元の価格以下であるべき
			if finalPrice > price {
				t.Errorf("final price %d is greater than original price %d", finalPrice, price)
			}

			// 100%割引の場合、最終価格は0であるべき
			if discountPercent == 100 && finalPrice != 0 {
				t.Errorf("100%% discount should result in 0, got %d", finalPrice)
			}
		}
	})
}

// FuzzParseUserAge はParseUserAge関数のfuzzテスト
func FuzzParseUserAge(f *testing.F) {
	// Seed corpus
	f.Add("25")
	f.Add("0")
	f.Add("150")
	f.Add("-1")
	f.Add("999")
	f.Add("")
	f.Add("abc")
	f.Add("12.5")
	f.Add("  30  ")

	f.Fuzz(func(t *testing.T, ageStr string) {
		// ParseUserAgeを実行
		age, err := ParseUserAge(ageStr)

		// エラーがない場合の不変条件
		if err == nil {
			// 年齢は0-150の範囲であるべき
			if age < 0 || age > 150 {
				t.Errorf("age %d is out of valid range [0, 150] for input %q", age, ageStr)
			}
		}

		// 負の年齢を返すことは絶対にない
		if age < 0 {
			t.Errorf("age cannot be negative: got %d for input %q", age, ageStr)
		}
	})
}

// 通常のテスト（fuzz testと併用可能）
func TestParseEmail(t *testing.T) {
	tests := []struct {
		input        string
		wantUsername string
		wantDomain   string
		wantErr      bool
	}{
		{"user@example.com", "user", "example.com", false},
		{"test@test.co.jp", "test", "test.co.jp", false},
		{"", "", "", true},
		{"@", "", "", true},
		{"user@", "", "", true},
		{"@domain.com", "", "", true},
		{"nodomain", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			username, domain, err := ParseEmail(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEmail(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if username != tt.wantUsername || domain != tt.wantDomain {
					t.Errorf("ParseEmail(%q) = (%q, %q), want (%q, %q)",
						tt.input, username, domain, tt.wantUsername, tt.wantDomain)
				}
			}
		})
	}
}
