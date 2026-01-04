package main

import (
	"fmt"
	"sync"
	"time"
)

// Example 1: 共有変数への非同期アクセス（data race）
func badCounter() {
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++ // race condition: 複数goroutineが同時に読み書き
		}()
	}

	wg.Wait()
	fmt.Println("Bad counter result:", counter) // 期待は100だが、結果は不定
}

// Example 1 Fixed: mutexで保護
func goodCounter() {
	counter := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println("Good counter result:", counter) // 常に100
}

// Example 2: マップへの並行書き込み（data race + panic）
func badMapWrite() {
	m := make(map[int]int)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m[n] = n * 2 // race + パニックの可能性
		}(i)
	}

	wg.Wait()
	fmt.Println("Bad map size:", len(m))
}

// Example 2 Fixed: sync.Mapを使用
func goodMapWrite() {
	var m sync.Map
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Store(n, n*2)
		}(i)
	}

	wg.Wait()
	count := 0
	m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	fmt.Println("Good map size:", count)
}

// Example 3: クロージャでのループ変数キャプチャ問題
func badLoopCapture() {
	var wg sync.WaitGroup
	items := []string{"apple", "banana", "cherry"}

	for _, item := range items {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// itemはループ変数への参照なので、最後の値（cherry）になりがち
			fmt.Println("Bad:", item)
			time.Sleep(10 * time.Millisecond)
		}()
	}

	wg.Wait()
}

// Example 3 Fixed: 引数として渡す
func goodLoopCapture() {
	var wg sync.WaitGroup
	items := []string{"apple", "banana", "cherry"}

	for _, item := range items {
		wg.Add(1)
		go func(i string) {
			defer wg.Done()
			fmt.Println("Good:", i)
			time.Sleep(10 * time.Millisecond)
		}(item) // ループ変数をコピーして引数に渡す
	}

	wg.Wait()
}

func main() {
	fmt.Println("=== Example 1: Counter ===")
	badCounter()
	goodCounter()

	fmt.Println("\n=== Example 2: Map ===")
	// badMapWrite() // パニックする可能性があるためコメントアウト
	goodMapWrite()

	fmt.Println("\n=== Example 3: Loop Capture ===")
	badLoopCapture()
	goodLoopCapture()
}
