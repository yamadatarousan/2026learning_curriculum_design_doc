package main

import (
	"errors"
	"fmt"
)

// divide関数
// 引数: a, b (両方ともfloat64)
// 戻り値: 結果(float64) と エラー(error) の2つ
func divide(a, b float64) (float64, error) {
	// もし b が 0 だったら...
	if b == 0 {
		// 結果として0 と、errors.New()で作った新しいエラーを返す
		return 0, errors.New("ゼロで割ることはできません")
	}

	// エラーがなければ、計算結果と nil (エラーなし) を返す
	return a / b, nil
}

func main() {
	// --- 成功するケース ---
	fmt.Println("--- 成功ケース ---")
	result, err := divide(10.0, 2.0)
	if err != nil {
		fmt.Println("エラー発生:", err)
	} else {
		fmt.Println("結果:", result)
	}

	// --- 失敗するケース ---
	fmt.Println("\n--- 失敗ケース ---")
	result, err = divide(10.0, 0.0)
	if err != nil {
		fmt.Println("エラー発生:", err)
	} else {
		fmt.Println("結果:", result)
	}
}
