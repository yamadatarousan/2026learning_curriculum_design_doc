package main

import (
	"fmt"
	"math/rand"
	"time"
)

// 擬似的なAPI呼び出し関数
// apiName: どのAPIを呼んだかを示す文字列
// ch: 結果を書き込むためのチャネル
func fetchFromAPI(apiName string, ch chan string) {
	// 100ミリ秒から1秒の間のランダムな時間、処理を待機させる
	sleepTime := time.Duration(100 + rand.Intn(900)) * time.Millisecond
	time.Sleep(sleepTime)

	// 処理結果を作成し、チャネルに送信する
	result := fmt.Sprintf("Result from %s (took %v)", apiName, sleepTime)
	ch <- result
}

func main() {
	// 乱数のシードを初期化（毎回違う結果にするため）
	rand.Seed(time.Now().UnixNano())

	// 結果を受け取るためのチャネルを作成
	ch := make(chan string)

	// 呼び出したいAPIのリスト
	apis := []string{"Google API", "Facebook API", "Twitter API", "GitHub API"}

	// --- Step 1: 並行処理の開始 ---
	// 各APIに対して、ゴルーチンを起動する
	for _, api := range apis {
		fmt.Printf("Fetching from %s...\n", api)
		go fetchFromAPI(api, ch)
	}

	fmt.Println("\nWaiting for results...")

	// --- Step 2: 結果の集約 ---
	// 起動したゴルーチンの数だけ、チャネルから結果を受信する
	// ループは len(apis) 回、実行される
	for i := 0; i < len(apis); i++ {
		// チャネルから結果が送られてくるまで、ここでブロックされる
		result := <-ch
		fmt.Println(result)
	}

	fmt.Println("\nAll APIs have responded.")
}
