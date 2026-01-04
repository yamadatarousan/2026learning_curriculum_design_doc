# Race Detector 使用ガイド

## Race Detector とは

Go言語に組み込まれている並行処理のバグ（data race）を検出するツールです。
複数のgoroutineが同じメモリ領域に同時にアクセスし、少なくとも1つが書き込みを行う場合に発生する問題を自動検出します。

## 基本的な使い方

### テストでの使用（推奨）

```bash
# 通常のテスト実行
go test -race

# 詳細なログ付き
go test -race -v

# 特定のパッケージ
go test -race ./...

# ベンチマークでも使用可能
go test -race -bench=.
```

### アプリケーション実行時の使用

```bash
# ビルド時に有効化
go run -race main.go

# バイナリに組み込む（本番環境では非推奨：パフォーマンス低下）
go build -race
./app
```

## Data Race のよくあるパターン

### 1. 共有変数への非同期アクセス

**問題のあるコード:**
```go
counter := 0
var wg sync.WaitGroup

for i := 0; i < 100; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        counter++ // data race!
    }()
}
wg.Wait()
```

**修正方法:**
```go
// 方法1: Mutexで保護
var mu sync.Mutex
mu.Lock()
counter++
mu.Unlock()

// 方法2: atomic操作を使用
atomic.AddInt64(&counter, 1)
```

### 2. マップへの並行書き込み

**問題のあるコード:**
```go
m := make(map[int]int)
for i := 0; i < 100; i++ {
    go func(n int) {
        m[n] = n * 2 // data race + panic の可能性
    }(i)
}
```

**修正方法:**
```go
// sync.Mapを使用
var m sync.Map
m.Store(key, value)
```

### 3. クロージャでのループ変数キャプチャ

**問題のあるコード:**
```go
for _, item := range items {
    go func() {
        fmt.Println(item) // すべて最後の値になる可能性
    }()
}
```

**修正方法:**
```go
for _, item := range items {
    go func(i string) {
        fmt.Println(i)
    }(item) // 値をコピーして引数に渡す
}
```

### 4. スライスへの並行書き込み

**問題のあるコード:**
```go
slice := make([]int, 0)
for i := 0; i < 100; i++ {
    go func(n int) {
        slice = append(slice, n) // data race!
    }(i)
}
```

**修正方法:**
```go
// 方法1: Mutexで保護
var mu sync.Mutex
mu.Lock()
slice = append(slice, n)
mu.Unlock()

// 方法2: チャネルを使用
ch := make(chan int, 100)
for i := 0; i < 100; i++ {
    go func(n int) {
        ch <- n
    }(i)
}
// 別のgoroutineで受信して追加
```

## Race Detector の制限と注意点

### 制限
1. **実行時検出のみ**: コードが実際に実行されたパスのみ検出（静的解析ではない）
2. **パフォーマンス低下**: メモリ使用量が5-10倍、実行速度が2-20倍遅くなる
3. **メモリオーバーヘッド**: 追跡用のメタデータが必要

### 本番環境での使用について
- **推奨しない**: パフォーマンスとメモリのオーバーヘッドが大きい
- **代替案**: 開発・テスト・ステージング環境で徹底的に実行

### false negative（検出漏れ）
- race detectorは実行されたコードパスのみ検出
- すべてのdata raceを保証するわけではない
- **対策**: テストカバレッジを高め、並行処理のパターンを網羅する

## CI/CD での組み込み方

### GitHub Actions の例

```yaml
name: Go Tests with Race Detector

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests with race detector
        run: go test -race -v ./...

      - name: Run race detector on main app (short run)
        run: timeout 30s go run -race . || true
```

### Makefileでの定義

```makefile
.PHONY: test-race
test-race:
	go test -race -v ./...

.PHONY: test-all
test-all: test test-race
```

## 運用上の推奨事項

### 開発フロー
1. **ローカル開発時**: 並行処理を含むコードを書いたら必ず `-race` でテスト
2. **PR作成時**: CI/CDで自動実行（必須）
3. **リリース前**: 統合テストで `-race` を実行

### 優先順位
- **High**: API サーバーのハンドラ（複数リクエストが並行実行される）
- **High**: 共有リソースへのアクセス（キャッシュ、DB接続プールなど）
- **Medium**: バックグラウンドワーカー
- **Low**: 単一goroutineのみで動作するCLIツール

### ベストプラクティス
1. **テストを書く**: 並行処理のコードには必ずテストを書く
2. **シンプルに保つ**: 並行処理は複雑になりやすいため、必要最小限に留める
3. **チャネルを活用**: 共有メモリよりチャネルでの通信を優先（"Share memory by communicating"）
4. **race detectorを信頼する**: 警告が出たら必ず修正する

## トラブルシューティング

### race detectorが反応しない
- コードが実際に並行実行されていない可能性
- テストケースで並行性を確保（`t.Parallel()` や `sync.WaitGroup`）

### 大量の警告が出る
- 既知の問題がある場合は、まず根本原因を修正
- サードパーティライブラリの問題の場合は、イシューを報告またはライブラリを変更

### パフォーマンスが悪すぎる
- テストを分割して並列実行
- `-race` は全テストではなく、並行処理を含む部分のみに適用

## 参考資料

- [Go公式ブログ: Data Race Detector](https://go.dev/blog/race-detector)
- [Go公式ドキュメント: Race Detector](https://go.dev/doc/articles/race_detector)
- ThreadSanitizer（race detectorの基盤技術）

---

**最終更新**: 2026-01-04
