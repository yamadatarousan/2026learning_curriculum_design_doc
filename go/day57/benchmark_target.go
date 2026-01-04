package main

import (
	"fmt"
	"strings"
	"unicode"
)

// IsPalindrome は文字列が回文かどうかを判定する（改善前のバージョン）
func IsPalindrome(s string) bool {
	// 空文字は回文とみなす
	if len(s) == 0 {
		return true
	}

	// 文字列を逆順にして比較（非効率）
	reversed := ""
	for i := len(s) - 1; i >= 0; i-- {
		reversed += string(s[i])
	}

	return s == reversed
}

// IsPalindromeOptimized は文字列が回文かどうかを判定する（改善後のバージョン）
func IsPalindromeOptimized(s string) bool {
	// 空文字は回文とみなす
	if len(s) == 0 {
		return true
	}

	// 両端から中央に向かって比較（効率的）
	for i := 0; i < len(s)/2; i++ {
		if s[i] != s[len(s)-1-i] {
			return false
		}
	}

	return true
}

// CountWords は文字列内の単語数をカウントする（改善前）
func CountWords(s string) int {
	if len(s) == 0 {
		return 0
	}

	// strings.Splitで分割（余分なメモリ確保）
	words := strings.Fields(s)
	return len(words)
}

// CountWordsOptimized は文字列内の単語数をカウントする（改善後）
func CountWordsOptimized(s string) int {
	if len(s) == 0 {
		return 0
	}

	count := 0
	inWord := false

	// 1文字ずつ走査（メモリ効率的）
	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}

	return count
}

// StringConcat は文字列を連結する（改善前）
func StringConcat(strs []string) string {
	result := ""
	for _, s := range strs {
		result += s // 非効率な連結
	}
	return result
}

// StringConcatOptimized は文字列を連結する（改善後）
func StringConcatOptimized(strs []string) string {
	var builder strings.Builder
	for _, s := range strs {
		builder.WriteString(s) // 効率的な連結
	}
	return builder.String()
}

// StringConcatJoin は文字列を連結する（さらに改善）
func StringConcatJoin(strs []string) string {
	return strings.Join(strs, "") // 標準ライブラリを活用
}

// FindMax はスライスの最大値を見つける（改善前）
func FindMax(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	// append を使った非効率な実装
	var sorted []int
	for _, n := range nums {
		sorted = append(sorted, n)
	}

	// バブルソート（O(n²)）
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] < sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted[0]
}

// FindMaxOptimized はスライスの最大値を見つける（改善後）
func FindMaxOptimized(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	max := nums[0]
	for _, n := range nums {
		if n > max {
			max = n
		}
	}

	return max
}

func main() {
	// IsPalindromeの例
	fmt.Println("IsPalindrome('racecar'):", IsPalindrome("racecar"))
	fmt.Println("IsPalindromeOptimized('racecar'):", IsPalindromeOptimized("racecar"))

	// CountWordsの例
	text := "Hello World this is a test"
	fmt.Println("CountWords:", CountWords(text))
	fmt.Println("CountWordsOptimized:", CountWordsOptimized(text))

	// StringConcatの例
	strs := []string{"Go", "is", "awesome"}
	fmt.Println("StringConcat:", StringConcat(strs))
	fmt.Println("StringConcatOptimized:", StringConcatOptimized(strs))

	// FindMaxの例
	nums := []int{3, 7, 2, 9, 1, 5}
	fmt.Println("FindMax:", FindMax(nums))
	fmt.Println("FindMaxOptimized:", FindMaxOptimized(nums))
}
