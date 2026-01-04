# Day 55: Race Detector 実行手順

## 今日の成果物

1. `day55_race_examples.go` - よくあるdata raceのサンプルコード（Bad/Good例）
2. `day55_race_test.go` - race detectorの動作を確認するテストコード
3. `day55_race_detector_guide.md` - race detector使用ガイド（完全版）
4. このファイル - 実行手順書

## 実行手順

### 1. サンプルコードの動作確認

```bash
cd /Users/user/Development/2026learning_curriculum_design_doc/go/examples

# 通常の実行（raceがあっても動作する可能性あり）
go run day55_race_examples.go
```

### 2. Race Detectorでサンプルコードを実行

```bash
# race detectorを有効にして実行（警告が表示される）
go run -race day55_race_examples.go
```

**期待される結果**: `badCounter()` や `badLoopCapture()` で "WARNING: DATA RACE" が表示される

### 3. テストでRace Detectorを実行

```bash
# 通常のテスト（raceがあっても一部は成功する）
go test -v day55_race_test.go day55_race_examples.go

# race detectorを有効にしてテスト（raceを検出）
go test -race -v day55_race_test.go day55_race_examples.go
```

**期待される結果**:
- `TestRaceConditionDetection` でdata raceが報告される
- `TestMapRace` でdata race（またはpanic）が報告される
- `TestNoRaceWithMutex` と `TestNoMapRaceWithSyncMap` は問題なく通過

### 4. 既存のアプリケーション（day54）でRace Detectorを実行

```bash
cd /Users/user/Development/2026learning_curriculum_design_doc/go/day54

# 既存のテストをrace detectorで実行
go test -race -v

# ベンチマークもrace detectorで実行（もしあれば）
go test -race -bench=.
```

**期待される結果**: 既存のコードにdata raceがなければ、すべてのテストが通過する

### 5. アプリケーションサーバーをrace detectorで起動（短時間）

```bash
cd /Users/user/Development/2026learning_curriculum_design_doc/go/day54

# Docker Composeでデータベースを起動
docker-compose -f ../../docker-compose.yml up -d

# race detectorを有効にしてサーバーを起動（パフォーマンスが低下する）
go run -race .
```

別のターミナルでAPIリクエストを送信：

```bash
# サインアップ
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"race-test@example.com","password":"password123"}'

# ログイン
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"race-test@example.com","password":"password123"}'

# 複数の並行リクエストを送信（data raceを検出しやすくする）
for i in {1..10}; do
  curl -X POST http://localhost:8080/login \
    -H "Content-Type: application/json" \
    -d '{"email":"race-test@example.com","password":"password123"}' &
done
wait
```

サーバーのログにdata raceの警告が出ないことを確認し、Ctrl+Cで停止。

## Race Detectorの出力例

### Data Raceが検出された場合

```
WARNING: DATA RACE
Read at 0x00c0001a0088 by goroutine 8:
  main.badCounter.func1()
      /path/to/file.go:15 +0x3e

Previous write at 0x00c0001a0088 by goroutine 7:
  main.badCounter.func1()
      /path/to/file.go:15 +0x54

Goroutine 8 (running) created at:
  main.badCounter()
      /path/to/file.go:13 +0x8e
```

この出力は以下を示しています：
- どのメモリアドレスでraceが発生したか
- どのgoroutineが読み書きしたか
- 該当するコードの行番号

## CI/CDへの組み込み（オプション）

### Makefileの作成（推奨）

`/Users/user/Development/2026learning_curriculum_design_doc/go/day54/Makefile` を作成：

```makefile
.PHONY: test
test:
	go test -v ./...

.PHONY: test-race
test-race:
	go test -race -v ./...

.PHONY: test-all
test-all: test test-race

.PHONY: run-race
run-race:
	go run -race .
```

実行：
```bash
make test-race
```

### GitHub Actions（参考）

`.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test -v ./...

      - name: Run tests with race detector
        run: go test -race -v ./...
```

## 学習のポイント

1. **Race Detectorは開発者の味方**: 並行処理のバグは再現が難しいため、自動検出できるのは非常に有用
2. **テストで必ず実行**: 並行処理を含むコードを書いたら、必ず `-race` でテストを実行する習慣をつける
3. **本番環境では使用しない**: パフォーマンスへの影響が大きいため、開発・テスト環境のみで使用
4. **警告は必ず修正**: race detectorの警告は false positive（誤検出）がほとんどないため、必ず修正する

## トラブルシューティング

### "race detector not supported" エラー
- 一部のOS/アーキテクチャではサポートされていない
- 主要なプラットフォーム（linux/amd64, darwin/amd64, windows/amd64など）では動作する

### テストがタイムアウトする
- race detectorは実行速度を2-20倍遅くする
- テストのタイムアウト時間を延長: `go test -race -timeout 5m`

### メモリ不足
- race detectorはメモリを5-10倍消費する
- テストを分割して実行

---

**完了条件**:
- サンプルコードでdata raceが検出されることを確認
- 既存のアプリケーション（day54）で `-race` を実行し、問題がないことを確認
- race detectorの使い方と運用方針を理解

---

**次のステップ（Day 56）**: fuzzの入口（入力境界の守り）
