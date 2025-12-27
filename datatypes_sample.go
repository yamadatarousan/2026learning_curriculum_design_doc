package main

import "fmt"

// 定数を宣言します。
// constブロックで複数の定数をまとめて定義できます。
const (
	AppName     = "Go言語データ型サンプル" // アプリケーション名 (string)
	Version     = "1.0.0"             // バージョン (string)
	DefaultPort = 8080                // デフォルトポート番号 (int)
	Pi          = 3.14159             // 円周率 (float64)
)

func main() {
	// --- 変数の宣言 ---
	// string: 文字列を格納します
	var userName string = "山田"

	// int: 整数を格納します
	// `:=` を使って型を推論させています
	userAge := 32

	// float64: 浮動小数点数を格納します
	userHeight := 175.5

	// bool: true (真) または false (偽) を格納します
	isMember := true

	// --- 値の出力 ---
	fmt.Println("--- 定数の表示 ---")
	fmt.Println("アプリ名:", AppName)
	fmt.Println("バージョン:", Version)
	fmt.Println("デフォルトポート:", DefaultPort)
	fmt.Println("円周率:", Pi)

	fmt.Println("\n--- 変数の表示 ---")
	// fmt.Printf を使うと、書式を指定して出力できます
	// %v は値 (value) を、%T は型 (Type) を出力します
	fmt.Printf("名前: %v (型: %T)\n", userName, userName)
	fmt.Printf("年齢: %v (型: %T)\n", userAge, userAge)
	fmt.Printf("身長: %v (型: %T)\n", userHeight, userHeight)
	fmt.Printf("会員: %v (型: %T)\n", isMember, isMember)

	// --- 値の変更 ---
	fmt.Println("\n--- 変数の値を変えてみる ---")
	userAge = 33 // `:=` は最初の宣言時のみ。再代入は `=` を使います
	fmt.Printf("新しい年齢: %v\n", userAge)
}
