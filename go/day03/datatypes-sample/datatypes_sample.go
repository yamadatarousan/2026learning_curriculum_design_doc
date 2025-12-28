package main

import "fmt"

const (
	AppName     = "Go言語データ型サンプル" // アプリケーション名 (string)
	Version     = "1.0.0"             // バージョン (string)
	DefaultPort = 8080                // デフォルトポート番号 (int)
	Pi          = 3.14159             // 円周率 (float64)
)

func main() {
	var userName string = "山田"

	userAge := 32

	userHeight := 175.5

	isMember := true

	fmt.Println("--- 定数の表示 ---")
	fmt.Println("アプリ名:", AppName)
	fmt.Println("バージョン:", Version)
	fmt.Println("デフォルトポート:", DefaultPort)
	fmt.Println("円周率:", Pi)
	
	fmt.Println("\n--- 変数の表示 ---")
	fmt.Printf("名前: %v (型: %T)\n", userName, userName)
	fmt.Printf("年齢: %v (型: %T)\n", userAge, userAge)
	fmt.Printf("身長: %v (型: %T)\n", userHeight, userHeight)
	fmt.Printf("会員: %v (型: %T)\n", isMember, isMember)

	fmt.Println("\n--- 変数の値を変えてみる ---")
	userAge = 33
	fmt.Printf("新しい年齢: %v\n", userAge)
}