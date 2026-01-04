package main

import (
	"strings"
	"testing"
)

// BenchmarkIsPalindrome は改善前の回文判定のベンチマーク
func BenchmarkIsPalindrome(b *testing.B) {
	testString := "racecar"

	// b.N回繰り返し実行される
	for i := 0; i < b.N; i++ {
		IsPalindrome(testString)
	}
}

// BenchmarkIsPalindromeOptimized は改善後の回文判定のベンチマーク
func BenchmarkIsPalindromeOptimized(b *testing.B) {
	testString := "racecar"

	for i := 0; i < b.N; i++ {
		IsPalindromeOptimized(testString)
	}
}

// BenchmarkIsPalindromeLong は長い文字列での回文判定のベンチマーク
func BenchmarkIsPalindromeLong(b *testing.B) {
	testString := strings.Repeat("racecar", 100) // 700文字

	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsPalindrome(testString)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			IsPalindromeOptimized(testString)
		}
	})
}

// BenchmarkCountWords は単語カウントのベンチマーク
func BenchmarkCountWords(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog"

	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CountWords(text)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			CountWordsOptimized(text)
		}
	})
}

// BenchmarkStringConcat は文字列連結のベンチマーク
func BenchmarkStringConcat(b *testing.B) {
	strs := []string{"Go", "is", "awesome", "and", "fast"}

	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StringConcat(strs)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StringConcatOptimized(strs)
		}
	})

	b.Run("Join", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StringConcatJoin(strs)
		}
	})
}

// BenchmarkStringConcatLarge は大量の文字列連結のベンチマーク
func BenchmarkStringConcatLarge(b *testing.B) {
	// 1000個の文字列
	strs := make([]string, 1000)
	for i := range strs {
		strs[i] = "test"
	}

	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StringConcat(strs)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StringConcatOptimized(strs)
		}
	})

	b.Run("Join", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			StringConcatJoin(strs)
		}
	})
}

// BenchmarkFindMax は最大値探索のベンチマーク
func BenchmarkFindMax(b *testing.B) {
	nums := []int{3, 7, 2, 9, 1, 5, 4, 8, 6, 10}

	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindMax(nums)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindMaxOptimized(nums)
		}
	})
}

// BenchmarkFindMaxLarge は大量データでの最大値探索のベンチマーク
func BenchmarkFindMaxLarge(b *testing.B) {
	// 1000個の数値
	nums := make([]int, 1000)
	for i := range nums {
		nums[i] = i
	}

	b.Run("Original", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindMax(nums)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			FindMaxOptimized(nums)
		}
	})
}

// BenchmarkWithSetup はセットアップコストを除外するベンチマークの例
func BenchmarkWithSetup(b *testing.B) {
	// セットアップ（計測対象外）
	testData := make([]string, 1000)
	for i := range testData {
		testData[i] = "test"
	}

	// タイマーをリセット
	b.ResetTimer()

	// ここから計測開始
	for i := 0; i < b.N; i++ {
		StringConcatOptimized(testData)
	}
}

// BenchmarkWithMemoryAllocation はメモリ割り当てを計測するベンチマーク
func BenchmarkWithMemoryAllocation(b *testing.B) {
	strs := []string{"Go", "is", "awesome"}

	b.Run("Original", func(b *testing.B) {
		// メモリ割り当てを報告
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			StringConcat(strs)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()

		for i := 0; i < b.N; i++ {
			StringConcatOptimized(strs)
		}
	})
}

// BenchmarkParallel は並列実行のベンチマーク
func BenchmarkParallel(b *testing.B) {
	testString := "racecar"

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			IsPalindromeOptimized(testString)
		}
	})
}

// 通常のテスト（ベンチマークと併用）
func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"racecar", true},
		{"hello", false},
		{"", true},
		{"a", true},
		{"ab", false},
		{"aba", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsPalindrome(tt.input)
			if got != tt.want {
				t.Errorf("IsPalindrome(%q) = %v, want %v", tt.input, got, tt.want)
			}

			// 最適化版も同じ結果を返すことを確認
			gotOptimized := IsPalindromeOptimized(tt.input)
			if gotOptimized != tt.want {
				t.Errorf("IsPalindromeOptimized(%q) = %v, want %v", tt.input, gotOptimized, tt.want)
			}
		})
	}
}
