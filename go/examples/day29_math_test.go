package main

import "testing"

// TestAdd は Add 関数のテストです。
// 関数名は `Test` で始まり、テスト対象の関数名などを続けます。
// 引数には必ず `t *testing.T` を取ります。
func TestAdd(t *testing.T) {
	// 1. 準備 (Arrange)
	x, y := 2, 3
	expected := 5

	// 2. 実行 (Act)
	result := Add(x, y)

	// 3. 検証 (Assert)
	if result != expected {
		// もし結果が期待値と異なれば、t.Errorf でエラーを報告し、テストを失敗させる
		t.Errorf("Add(%d, %d) = %d; want %d", x, y, result, expected)
	}
}
