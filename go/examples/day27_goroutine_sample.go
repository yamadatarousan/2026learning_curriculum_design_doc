package main

import (
	"fmt"
	"time"
)

func say(s string) {
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(s)
	}
}

func main() {
	go say("World") // say("World")をゴルーチンとして起動
	say("Hello")   // say("Hello")は通常の関数として実行
}
