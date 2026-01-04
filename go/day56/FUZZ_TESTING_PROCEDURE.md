# Day 56: Fuzz Testing 実行手順

## 今日の成果物

1. `day56_fuzz_target.go` - fuzz testingの対象となる関数（4種類）
2. `day56_fuzz_test.go` - fuzz testのサンプルコード
3. `day56_fuzz_testing_guide.md` - fuzz testing完全ガイド
4. このファイル - 実行手順書

## 実行手順

### 1. サンプルコードの動作確認

```bash
cd /Users/user/Development/2026learning_curriculum_design_doc/go/examples

# 通常の実行（正常系の動作確認）
go run day56_fuzz_target.go
```

**期待される結果**:
```
Email parsed: username=user, domain=example.com
Sanitized: HelloWorldTest
Final price: 800
Age: 25
```

---

### 2. 通常のテストを実行（fuzzなし）

```bash
# 通常のテスト（固定入力）
go test -v -run TestParseEmail day56_fuzz_test.go day56_fuzz_target.go
```

**期待される結果**: すべてのテストケースがPASS

---

### 3. Fuzz Testを実行（短時間）

#### 3-1. ParseEmailのfuzz test（10秒間）

```bash
go test -fuzz=FuzzParseEmail -fuzztime=10s day56_fuzz_test.go day56_fuzz_target.go
```

**注目ポイント**:
- ランダムな文字列が大量に生成されてテストされる
- 実行中の統計情報が表示される:
  ```
  fuzz: elapsed: 0s, execs: 1234 (12340/sec), new interesting: 5 (total: 10)
  fuzz: elapsed: 3s, execs: 45678 (15000/sec), new interesting: 8 (total: 13)
  ```

**確認すべきこと**:
✅ パニックせずに完了する
✅ `new interesting` が増えている（新しいコードパスが発見されている）
✅ 失敗した場合、どの入力で失敗したかが表示される

#### 3-2. SanitizeInputのfuzz test（10秒間）

```bash
go test -fuzz=FuzzSanitizeInput -fuzztime=10s day56_fuzz_test.go day56_fuzz_target.go
```

#### 3-3. CalculateDiscountのfuzz test（10秒間）

```bash
go test -fuzz=FuzzCalculateDiscount -fuzztime=10s day56_fuzz_test.go day56_fuzz_target.go
```

#### 3-4. ParseUserAgeのfuzz test（10秒間）

```bash
go test -fuzz=FuzzParseUserAge -fuzztime=10s day56_fuzz_test.go day56_fuzz_target.go
```

---

### 4. すべてのFuzz Testを実行

```bash
# すべてのfuzz testを各10秒ずつ実行
# 注意: これは順番に実行されるため、合計40秒かかる
go test -fuzz=. -fuzztime=10s day56_fuzz_test.go day56_fuzz_target.go
```

---

### 5. Corpusの確認

fuzz testを実行すると、興味深い入力が自動保存されます。

```bash
# corpusディレクトリを確認
ls -la testdata/fuzz/

# 特定のfuzz testのcorpusを確認
ls -la testdata/fuzz/FuzzParseEmail/

# corpusの内容を表示（バイナリの場合は注意）
cat testdata/fuzz/FuzzParseEmail/*
```

**注目ポイント**:
- fuzzingで見つかった「興味深い」入力が保存されている
- 次回の実行時、これらのcorpusが自動的に使用される
- バグを再現するための入力データ

---

### 6. 失敗を意図的に発生させる（学習用）

ParseEmail関数に意図的なバグを入れて、fuzzingで検出できることを確認します。

#### 手順:

1. `day56_fuzz_target.go` の `ParseEmail` 関数を一時的に修正:

```go
// 元のコード:
if len(email) > 254 {
    return "", "", errors.New("email too long")
}

// バグを入れる（コメントアウト）:
// if len(email) > 254 {
//     return "", "", errors.New("email too long")
// }
```

2. fuzz testを実行:

```bash
go test -fuzz=FuzzParseEmail -fuzztime=30s day56_fuzz_test.go day56_fuzz_target.go
```

3. **期待される結果**: 長すぎる入力でバグが検出される

```
--- FAIL: FuzzParseEmail (X.XXs)
    --- FAIL: FuzzParseEmail (0.00s)
        day56_fuzz_test.go:XX: ...

    Failing input written to testdata/fuzz/FuzzParseEmail/abc123...
    To re-run:
    go test -run=FuzzParseEmail/abc123...
```

4. 失敗した入力で再現:

```bash
go test -run=FuzzParseEmail/abc123... day56_fuzz_test.go day56_fuzz_target.go
```

5. バグを修正して元に戻す

6. corpusをクリーンアップ（オプション）:

```bash
rm -rf testdata/fuzz
```

---

### 7. 長時間のFuzz Test（時間がある場合）

```bash
# 1時間実行（実用的な時間）
go test -fuzz=FuzzParseEmail -fuzztime=1h day56_fuzz_test.go day56_fuzz_target.go

# 10000回実行
go test -fuzz=FuzzParseEmail -fuzztime=10000x day56_fuzz_test.go day56_fuzz_target.go
```

**注意**: 長時間実行する場合は、バックグラウンドで実行するか、別ターミナルで実行してください。

---

### 8. 並列実行（パフォーマンス向上）

```bash
# 4つのワーカーで並列実行
go test -fuzz=FuzzParseEmail -fuzztime=1m -parallel=4 day56_fuzz_test.go day56_fuzz_target.go
```

**注目ポイント**:
- CPU使用率が上がる
- `execs/sec` の値が増える
- 短時間でより多くの入力をテストできる

---

## Fuzz Testingの出力の読み方

### 正常終了の例

```
fuzz: elapsed: 0s, execs: 0 (0/sec), new interesting: 0 (total: 9)
fuzz: elapsed: 3s, execs: 45234 (15078/sec), new interesting: 2 (total: 11)
fuzz: elapsed: 6s, execs: 92156 (15640/sec), new interesting: 0 (total: 11)
fuzz: elapsed: 9s, execs: 138923 (15596/sec), new interesting: 1 (total: 12)
fuzz: elapsed: 10s, execs: 154321 (15432/sec), new interesting: 0 (total: 12)
PASS
ok      command-line-arguments  10.234s
```

**各項目の意味**:
- `elapsed`: 経過時間
- `execs`: 実行された入力の総数
- `execs/sec`: 1秒あたりの実行回数（高いほど効率的）
- `new interesting`: 新しく見つかった興味深い入力（新しいコードパスをカバー）
- `total`: 累計の興味深い入力数

### 失敗の例

```
fuzz: elapsed: 3s, execs: 45234 (15078/sec), new interesting: 5 (total: 14)
--- FAIL: FuzzParseEmail (3.12s)
    --- FAIL: FuzzParseEmail (0.00s)
        day56_fuzz_test.go:25: username is empty but no error returned for email: "user@domain@com"

    Failing input written to testdata/fuzz/FuzzParseEmail/a1b2c3d4e5f6...
    To re-run:
    go test -run=FuzzParseEmail/a1b2c3d4e5f6...
FAIL
exit status 1
FAIL    command-line-arguments  3.234s
```

**失敗時の対応**:
1. エラーメッセージを確認（どの不変条件が違反されたか）
2. 失敗した入力を確認（`"user@domain@com"`）
3. コードを修正
4. 保存されたcorpusで再現テスト

---

## 学習のポイント

### 1. Fuzz Testingの価値

- **開発者が想定していない入力を自動生成**
- 通常のテストでは見つけにくいバグを発見
- セキュリティ脆弱性の早期発見

### 2. 不変条件の重要性

Fuzz testでは「正しい出力」ではなく、「守るべきルール」をチェックする：

```go
// 良い不変条件の例:
// - エラーがない場合、結果は有効
// - 負の値を返さない
// - 入力より大きい値を返さない
// - panicしない
```

### 3. Seed Corpusの戦略

良いseed corpusは、fuzzingの効率を大きく向上させる：
- エッジケース（空、最小、最大）
- 正常系の代表例
- 既知の問題パターン

### 4. 実務での使い分け

| テスト種類 | 使いどころ |
|-----------|-----------|
| 通常のテスト | 既知のケース、リグレッション防止 |
| Fuzz test | 未知のバグ発見、セキュリティチェック |
| Race detector | 並行処理のバグ |

---

## CI/CDへの組み込み例

### Makefile

```makefile
.PHONY: test-fuzz-short
test-fuzz-short:
	go test -fuzz=. -fuzztime=10s ./...

.PHONY: test-fuzz-long
test-fuzz-long:
	go test -fuzz=. -fuzztime=1h ./...
```

### GitHub Actions

```yaml
name: Fuzz Tests

on: [push, pull_request]

jobs:
  fuzz-short:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run fuzz tests (10s each)
        run: go test -fuzz=. -fuzztime=10s ./...

  fuzz-long:
    runs-on: ubuntu-latest
    if: github.event_name == 'schedule'  # Nightly build
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run fuzz tests (1h each)
        run: go test -fuzz=. -fuzztime=1h ./...
```

---

## トラブルシューティング

### "fuzz tests are only supported at Go 1.18 or higher"

- Go 1.18以上にアップグレードしてください
- `go version` で確認

### メモリ不足

```bash
# 入力サイズを制限
f.Fuzz(func(t *testing.T, input string) {
    if len(input) > 10000 {
        return  // 大きすぎる入力はスキップ
    }
    // ...
})
```

### 実行が遅い

- `-parallel` オプションで並列度を上げる
- seed corpusを減らす
- 対象関数を軽量化

---

**完了条件**:
- fuzz testの基本的な実行方法を理解
- fuzz testingの出力の読み方を理解
- 不変条件のチェック方法を理解
- 実務での使いどころを把握

---

**次のステップ（Day 57）**: ベンチの入口（計測癖）
