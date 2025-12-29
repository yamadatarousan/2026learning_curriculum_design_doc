package main

import (
	"fmt"
	"time"
)

// 時間のかかる処理（のフリ）
func worker(ch chan string) {
	fmt.Println("...worker is doing something...")
	time.Sleep(2 * time.Second)
	fmt.Println("...worker finished!")

	// 処理結果をチャネルに送信
	ch <- "Job finished"
}

func main() {
	// string型をやり取りするチャネルを作成
	ch := make(chan string)

	// workerをゴルーチンとして起動
	go worker(ch)

	fmt.Println("Waiting for worker...")

	// チャネルからデータを受信するまで、ここで処理がブロックされる
	result := <-ch

	fmt.Println("Result from worker:", result)
	fmt.Println("Main function finished.")
}
