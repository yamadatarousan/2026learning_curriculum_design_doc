package main

import (
	"sync"
	"testing"
)

// TestRaceConditionDetection は、race detectorが実際にdata raceを検出することを示すテスト
// go test -race を実行すると、data raceが報告される
func TestRaceConditionDetection(t *testing.T) {
	// このテストは通常のテストとしては成功するが、-race フラグを付けると失敗する
	counter := 0
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// ここでdata raceが発生する
			temp := counter
			temp++
			counter = temp
		}()
	}

	wg.Wait()

	// 結果は不定なので、期待値チェックはしない
	t.Logf("Counter value: %d (expected 10, but may differ due to race)", counter)
}

// TestNoRaceWithMutex は、mutexで保護した場合にrace detectorが問題を報告しないことを示すテスト
func TestNoRaceWithMutex(t *testing.T) {
	counter := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()

	if counter != 10 {
		t.Errorf("Expected counter to be 10, got %d", counter)
	}
}

// TestMapRace は、マップへの並行書き込みでdata raceを検出するテスト
func TestMapRace(t *testing.T) {
	m := make(map[int]int)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			// ここでdata race（とpanic）が発生する可能性がある
			m[n] = n * 2
		}(i)
	}

	wg.Wait()
	t.Logf("Map size: %d", len(m))
}

// TestNoMapRaceWithSyncMap は、sync.Mapを使った安全な実装
func TestNoMapRaceWithSyncMap(t *testing.T) {
	var m sync.Map
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
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

	if count != 10 {
		t.Errorf("Expected map size to be 10, got %d", count)
	}
}

// BenchmarkCounterWithMutex は、mutexを使ったカウンタのベンチマーク
// race detectorはベンチマークでも使用可能: go test -bench=. -race
func BenchmarkCounterWithMutex(b *testing.B) {
	counter := 0
	var mu sync.Mutex

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			counter++
			mu.Unlock()
		}
	})
}
