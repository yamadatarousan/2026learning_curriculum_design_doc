package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

// データベースを準備し、テスト用のデータを入れるヘルパー関数
func setupTestDB() {
	// 既存のテーブルを削除して、まっさらな状態から始める
	db.Exec("DROP TABLE IF EXISTS todos")
	initDB() // 再度テーブルを作成

	// テスト用のデータを1件挿入
	db.Exec(`INSERT INTO todos (id, name) VALUES (1, "Test Todo 1")`)
}

func TestGetTodosHandler(t *testing.T) {
	// --- 準備 (Arrange) ---
	setupTestDB() // データベースを初期化

	// テスト用のリクエストとレスポンスレコーダーを作成
	req := httptest.NewRequest("GET", "/todos", nil)
	rr := httptest.NewRecorder()

	// --- 実行 (Act) ---
	// ハンドラを直接呼び出す
	getTodosHandler(rr, req)

	// --- 検証 (Assert) ---
	// 1. ステータスコードを検証
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// 2. レスポンスボディ（JSON）を検証
	// 期待されるレスポンスボディを作成
	expected := []Todo{{ID: 1, Name: "Test Todo 1"}}

	// 実際のレスポンスボディをデコード
	var actual []Todo
	if err := json.NewDecoder(rr.Body).Decode(&actual); err != nil {
		t.Fatalf("Could not decode response body: %v", err)
	}

	// デコードした結果と期待値を比較
	// reflect.DeepEqual は、スライスや構造体など複雑なデータ型を比較するのに便利
	if !reflect.DeepEqual(actual, expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			actual, expected)
	}
}
