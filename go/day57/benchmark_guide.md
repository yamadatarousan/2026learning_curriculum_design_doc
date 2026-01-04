# Benchmark Testing 使用ガイド

## Benchmark Testing とは

**ベンチマークテスト**は、コードのパフォーマンス（実行時間、メモリ使用量）を定量的に計測するテスト手法です。

### なぜベンチマークが必要か

1. **推測ではなく計測**
   - 「このコードは遅い」という推測ではなく、数値で判断
   - 直感に反する結果もある（小さな最適化は効果がないことも）

2. **パフォーマンス劣化の早期発見**
   - コード変更によるパフォーマンス低下を検出
   - リファクタリング前後の比較

3. **ボトルネックの特定**
   - どの部分が遅いかを客観的に把握
   - 最適化の優先順位を決定

4. **改善効果の検証**
   - 最適化が本当に効果があったかを確認
   - 改善前後を数値で比較

---

## 基本的な書き方

### 最小構成

```go
func BenchmarkXxx(b *testing.B) {
    for i := 0; i < b.N; i++ {
        // 計測したい処理
        YourFunction()
    }
}
```

### 重要なポイント

1. **関数名は `Benchmark` で始める**
2. **引数は `*testing.B`**
3. **`b.N` 回繰り返す** - Goが自動的に調整する回数

---

## 実行方法

### 基本的な実行

```bash
# すべてのベンチマークを実行
go test -bench=.

# 特定のベンチマークを実行
go test -bench=BenchmarkIsPalindrome

# 正規表現でフィルタ
go test -bench=Palindrome
```

### 実行時間の調整

```bash
# デフォルトは1秒、3秒間実行
go test -bench=. -benchtime=3s

# 回数指定（1000回）
go test -bench=. -benchtime=1000x
```

### メモリ割り当ての計測

```bash
# メモリ統計を表示
go test -bench=. -benchmem
```

### 結果の保存と比較

```bash
# 結果をファイルに保存
go test -bench=. -benchmem > old.txt

# コードを修正後、再度実行
go test -bench=. -benchmem > new.txt

# benchstatで比較（要インストール）
go install golang.org/x/perf/cmd/benchstat@latest
benchstat old.txt new.txt
```

---

## ベンチマーク結果の読み方

### 基本的な出力

```
BenchmarkIsPalindrome-8         10000000        150 ns/op
```

各フィールドの意味：
- `BenchmarkIsPalindrome` - ベンチマーク名
- `-8` - GOMAXPROCS（使用されたCPUコア数）
- `10000000` - 実行回数（b.N）
- `150 ns/op` - 1回あたりの実行時間（ナノ秒）

### メモリ統計付き出力

```
BenchmarkStringConcat-8         1000000         1200 ns/op        512 B/op        5 allocs/op
```

追加フィールド：
- `512 B/op` - 1回あたりのメモリ割り当て量（バイト）
- `5 allocs/op` - 1回あたりのメモリ割り当て回数

### サブベンチマークの出力

```
BenchmarkStringConcat/Original-8        500000    3000 ns/op
BenchmarkStringConcat/Optimized-8      5000000     300 ns/op
```

`/` の後がサブベンチマーク名

---

## ベンチマークの書き方パターン

### パターン1: 基本的なベンチマーク

```go
func BenchmarkMyFunction(b *testing.B) {
    for i := 0; i < b.N; i++ {
        MyFunction()
    }
}
```

### パターン2: セットアップを除外

```go
func BenchmarkWithSetup(b *testing.B) {
    // セットアップ（計測対象外）
    testData := prepareTestData()

    // タイマーをリセット
    b.ResetTimer()

    // ここから計測
    for i := 0; i < b.N; i++ {
        MyFunction(testData)
    }
}
```

### パターン3: サブベンチマーク（複数バージョンの比較）

```go
func BenchmarkComparison(b *testing.B) {
    b.Run("VersionA", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            VersionA()
        }
    })

    b.Run("VersionB", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            VersionB()
        }
    })
}
```

### パターン4: メモリ割り当ての計測

```go
func BenchmarkWithMemory(b *testing.B) {
    b.ReportAllocs() // メモリ統計を報告

    for i := 0; i < b.N; i++ {
        MyFunction()
    }
}
```

### パターン5: 並列ベンチマーク

```go
func BenchmarkParallel(b *testing.B) {
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            MyFunction()
        }
    })
}
```

### パターン6: 異なる入力サイズでのベンチマーク

```go
func BenchmarkDifferentSizes(b *testing.B) {
    sizes := []int{10, 100, 1000, 10000}

    for _, size := range sizes {
        b.Run(fmt.Sprintf("Size-%d", size), func(b *testing.B) {
            data := generateData(size)
            b.ResetTimer()

            for i := 0; i < b.N; i++ {
                MyFunction(data)
            }
        })
    }
}
```

---

## パフォーマンス改善のサイクル

### 1. 計測（Measure）

```bash
# 現在のパフォーマンスを計測
go test -bench=. -benchmem > before.txt
```

### 2. 分析（Analyze）

- ボトルネックを特定
- なぜ遅いのかを考える
- 仮説を立てる

### 3. 改善（Optimize）

- コードを修正
- 最適化を適用

### 4. 再計測（Re-measure）

```bash
# 改善後のパフォーマンスを計測
go test -bench=. -benchmem > after.txt

# 比較
benchstat before.txt after.txt
```

### 5. 検証（Verify）

- 機能が壊れていないか確認（通常のテスト）
- 期待通りの改善があったか確認

### 6. 繰り返し

- さらなる改善の余地があれば繰り返す

---

## benchstatによる比較

### インストール

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

### 使い方

```bash
# 2つのベンチマーク結果を比較
benchstat old.txt new.txt
```

### 出力例

```
name                old time/op    new time/op    delta
IsPalindrome-8        150ns ± 2%      75ns ± 1%  -50.00%  (p=0.000 n=10+10)
StringConcat-8       3000ns ± 5%     300ns ± 3%  -90.00%  (p=0.000 n=10+10)

name                old alloc/op   new alloc/op   delta
StringConcat-8         512B ± 0%       48B ± 0%  -90.62%  (p=0.000 n=10+10)

name                old allocs/op  new allocs/op  delta
StringConcat-8         5.00 ± 0%      1.00 ± 0%  -80.00%  (p=0.000 n=10+10)
```

**読み方**:
- `delta` - 変化率（負の値は改善）
- `-50.00%` - 50%高速化
- `-90.00%` - 90%高速化
- `p=0.000` - 統計的に有意な差

---

## よくある最適化パターンと効果

### 1. 文字列連結: `+=` → `strings.Builder`

**改善前**:
```go
result := ""
for _, s := range strs {
    result += s  // 毎回新しい文字列を確保
}
```

**改善後**:
```go
var builder strings.Builder
for _, s := range strs {
    builder.WriteString(s)  // 1つのバッファに追加
}
result := builder.String()
```

**効果**: 10倍〜100倍高速化、メモリ使用量90%削減

---

### 2. ループの最適化: 不要な処理を外に

**改善前**:
```go
for i := 0; i < len(data); i++ {
    prefix := "user_"  // 毎回確保
    result = prefix + data[i]
}
```

**改善後**:
```go
prefix := "user_"  // ループの外
for i := 0; i < len(data); i++ {
    result = prefix + data[i]
}
```

**効果**: 小さいが積み重なると大きい

---

### 3. アルゴリズムの改善

**改善前**: O(n²)
```go
// バブルソートで最大値を探す
for i := 0; i < len(nums); i++ {
    for j := i + 1; j < len(nums); j++ {
        if nums[i] < nums[j] {
            nums[i], nums[j] = nums[j], nums[i]
        }
    }
}
max := nums[0]
```

**改善後**: O(n)
```go
// 1回走査で最大値を見つける
max := nums[0]
for _, n := range nums {
    if n > max {
        max = n
    }
}
```

**効果**: 100倍〜10000倍高速化（データサイズによる）

---

### 4. メモリ割り当ての削減

**改善前**:
```go
var result []int
for _, n := range data {
    result = append(result, n*2)  // 何度もメモリ再確保
}
```

**改善後**:
```go
result := make([]int, 0, len(data))  // 事前に容量確保
for _, n := range data {
    result = append(result, n*2)
}
```

**効果**: メモリ割り当て回数が10分の1以下

---

## ベンチマークのベストプラクティス

### 1. 意味のあるベンチマークを書く

❌ **悪い例**: 現実的でない入力
```go
func BenchmarkBad(b *testing.B) {
    for i := 0; i < b.N; i++ {
        MyFunction("a")  // 実際の使用例とかけ離れている
    }
}
```

✅ **良い例**: 実際の使用例に近い入力
```go
func BenchmarkGood(b *testing.B) {
    input := generateRealisticInput()  // 実際のデータに近い
    for i := 0; i < b.N; i++ {
        MyFunction(input)
    }
}
```

### 2. セットアップを計測から除外

```go
func BenchmarkWithProperSetup(b *testing.B) {
    data := prepareData()  // セットアップ
    b.ResetTimer()         // タイマーリセット

    for i := 0; i < b.N; i++ {
        Process(data)      // ここだけ計測
    }
}
```

### 3. 複数のサイズでテスト

```go
func BenchmarkMultipleSizes(b *testing.B) {
    for _, size := range []int{10, 100, 1000} {
        b.Run(fmt.Sprintf("n=%d", size), func(b *testing.B) {
            data := generateData(size)
            b.ResetTimer()
            for i := 0; i < b.N; i++ {
                Process(data)
            }
        })
    }
}
```

### 4. コンパイラ最適化を防ぐ

```go
var result int  // グローバル変数

func BenchmarkPreventOptimization(b *testing.B) {
    var r int  // ローカル変数
    for i := 0; i < b.N; i++ {
        r = ExpensiveCalculation()  // 結果を使う
    }
    result = r  // コンパイラ最適化を防ぐ
}
```

### 5. 通常のテストも併用

```go
// 正しい結果を返すことを確認
func TestMyFunction(t *testing.T) {
    // ...
}

// パフォーマンスを計測
func BenchmarkMyFunction(b *testing.B) {
    // ...
}
```

---

## 実務での使用シーン

### 1. パフォーマンスが重要な関数

- データ処理のループ
- API のホットパス（頻繁に呼ばれる処理）
- 暗号化/復号化
- シリアライズ/デシリアライズ

### 2. リファクタリング時

- 変更前後でパフォーマンス劣化がないことを確認
- CI/CDで自動チェック

### 3. アルゴリズムの選定

- 複数の実装を比較
- データサイズに応じた適切なアルゴリズムを選択

### 4. ライブラリの評価

- 複数のライブラリを比較
- パフォーマンスとメモリ使用量で判断

---

## CI/CDでの運用

### GitHub Actionsの例

```yaml
name: Benchmarks

on: [push, pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Run benchmarks
        run: go test -bench=. -benchmem ./...

      - name: Compare with main branch
        run: |
          git fetch origin main
          git checkout origin/main
          go test -bench=. -benchmem > old.txt
          git checkout -
          go test -bench=. -benchmem > new.txt
          benchstat old.txt new.txt
```

---

## トラブルシューティング

### ベンチマーク結果がばらつく

- 他のプロセスを止める
- `-benchtime` を長くする（より多くのサンプル）
- `-count` で複数回実行（例: `-count=10`）

### ベンチマークが早すぎる

- コンパイラが最適化している可能性
- 結果を使う（グローバル変数に代入など）

### メモリベンチマークが表示されない

- `-benchmem` フラグを付ける
- または `b.ReportAllocs()` を呼ぶ

---

## 参考資料

- [Go公式ドキュメント: Benchmarks](https://pkg.go.dev/testing#hdr-Benchmarks)
- [Profiling Go Programs](https://go.dev/blog/pprof)
- [benchstat tool](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat)

---

**最終更新**: 2026-01-04
