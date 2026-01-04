# Day 57: Benchmark Testing 実行手順

## 今日の成果物

1. `day57_benchmark_target.go` - ベンチマーク対象の関数（改善前/改善後の実装）
2. `day57_benchmark_test.go` - ベンチマークテスト
3. `day57_benchmark_guide.md` - ベンチマーク完全ガイド
4. このファイル - 実行手順書

---

## 実行手順

### 手順1: サンプルコードの動作確認

```bash
cd /Users/user/Development/2026learning_curriculum_design_doc/go/examples

# 通常の実行（正常系の確認）
go run day57_benchmark_target.go
```

**期待される結果**: 各関数が正常に動作することを確認

---

### 手順2: 通常のテストを実行

```bash
# 機能テスト（改善前と改善後が同じ結果を返すことを確認）
go test -v -run TestIsPalindrome day57_benchmark_test.go day57_benchmark_target.go
```

**注目ポイント**:
- 改善前と改善後で同じ結果を返すことを確認
- パフォーマンス改善は**機能を壊さない**ことが大前提

---

### 手順3: 基本的なベンチマークを実行（重要）

```bash
# すべてのベンチマークを実行
go test -bench=. day57_benchmark_test.go day57_benchmark_target.go
```

**注目ポイント①: 基本的な出力**

```
BenchmarkIsPalindrome-8                   5000000        300 ns/op
BenchmarkIsPalindromeOptimized-8         20000000         75 ns/op
```

**各フィールドの意味**:
- `BenchmarkIsPalindrome-8` - ベンチマーク名（-8はCPUコア数）
- `5000000` - 実行回数（Go が自動調整）
- `300 ns/op` - 1回あたりの実行時間（ナノ秒）

**確認すべきこと**:
✅ Optimized版が**4倍高速**（300ns → 75ns）
✅ 実行回数は処理速度によって自動調整される

---

### 手順4: メモリ統計付きベンチマーク（重要）

```bash
# メモリ割り当ても計測
go test -bench=. -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**注目ポイント②: メモリ統計の出力**

```
BenchmarkStringConcat/Original-8          500000    3000 ns/op    512 B/op    5 allocs/op
BenchmarkStringConcat/Optimized-8        5000000     300 ns/op     48 B/op    1 allocs/op
```

**追加フィールド**:
- `512 B/op` - 1回あたりのメモリ割り当て量（バイト）
- `5 allocs/op` - 1回あたりのメモリ割り当て回数

**確認すべきこと**:
✅ Optimized版は**10倍高速**（3000ns → 300ns）
✅ メモリ使用量が**90%削減**（512B → 48B）
✅ メモリ割り当て回数が**80%削減**（5回 → 1回）

---

### 手順5: 特定のベンチマークのみ実行

```bash
# 回文判定のみ
go test -bench=Palindrome -benchmem day57_benchmark_test.go day57_benchmark_target.go

# 文字列連結のみ
go test -bench=StringConcat -benchmem day57_benchmark_test.go day57_benchmark_target.go

# 最大値探索のみ
go test -bench=FindMax -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**注目ポイント**:
- 特定の関数に絞って計測できる
- 正規表現でフィルタ可能

---

### 手順6: 実行時間を調整

```bash
# デフォルトは1秒、3秒間実行してより正確な結果を得る
go test -bench=StringConcat -benchtime=3s -benchmem day57_benchmark_test.go day57_benchmark_target.go

# 回数指定（1000回）
go test -bench=StringConcat -benchtime=1000x -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**注目ポイント**:
- 長時間実行するとより安定した結果が得られる
- 実行回数を指定することもできる

---

### 手順7: 結果を保存して比較（重要な学び）

#### ステップ1: 現在の結果を保存

```bash
# 改善前（Original）の結果を保存
go test -bench=StringConcat/Original -benchmem day57_benchmark_test.go day57_benchmark_target.go > before.txt

# 結果を確認
cat before.txt
```

#### ステップ2: 改善後の結果を保存

```bash
# 改善後（Optimized）の結果を保存
go test -bench=StringConcat/Optimized -benchmem day57_benchmark_test.go day57_benchmark_target.go > after.txt

# 結果を確認
cat after.txt
```

#### ステップ3: benchstatで比較（オプション）

```bash
# benchstatをインストール（初回のみ）
go install golang.org/x/perf/cmd/benchstat@latest

# 2つの結果を比較
benchstat before.txt after.txt
```

**期待される出力**:

```
name              old time/op    new time/op    delta
StringConcat-8      3000ns ± 5%     300ns ± 3%  -90.00%

name              old alloc/op   new alloc/op   delta
StringConcat-8       512B ± 0%       48B ± 0%  -90.62%
```

**読み方**:
- `-90.00%` - 90%高速化（10倍速い）
- `-90.62%` - メモリ使用量が90%削減

---

### 手順8: 大量データでのベンチマーク

```bash
# 1000個の文字列連結
go test -bench=StringConcatLarge -benchmem day57_benchmark_test.go day57_benchmark_target.go

# 1000個の数値で最大値探索
go test -bench=FindMaxLarge -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**注目ポイント**:
- データサイズが大きくなると、最適化の効果が**より顕著**になる
- `FindMaxLarge/Original` は O(n²) なので、極端に遅い
- `FindMaxLarge/Optimized` は O(n) なので、高速

**確認すべきこと**:
✅ アルゴリズムの違いが大きく影響する
✅ データサイズによって最適な実装が変わる

---

### 手順9: 並列ベンチマーク（オプション）

```bash
# 並列実行のベンチマーク
go test -bench=Parallel -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**注目ポイント**:
- 複数のgoroutineで並行実行した場合のパフォーマンス
- 並行処理が安全かどうかも確認できる

---

## ベンチマーク結果の読み方（まとめ）

### 基本的な出力

```
BenchmarkIsPalindrome-8        5000000        300 ns/op
```

| 項目 | 意味 |
|------|------|
| `BenchmarkIsPalindrome` | ベンチマーク名 |
| `-8` | 使用CPUコア数（GOMAXPROCS） |
| `5000000` | 実行回数（b.N） |
| `300 ns/op` | 1回あたりの実行時間 |

### メモリ統計付き出力

```
BenchmarkStringConcat-8    500000    3000 ns/op    512 B/op    5 allocs/op
```

| 項目 | 意味 |
|------|------|
| `3000 ns/op` | 1回あたりの実行時間 |
| `512 B/op` | 1回あたりのメモリ割り当て量 |
| `5 allocs/op` | 1回あたりのメモリ割り当て回数 |

### 改善の判断基準

| 改善度 | 評価 |
|--------|------|
| 2倍以上高速化 | 明確な改善 ✅ |
| 1.5〜2倍 | 有意な改善 ✅ |
| 1.1〜1.5倍 | 小さな改善 |
| 1.1倍未満 | 誤差の範囲かも |

**メモリ割り当て**:
- 割り当て回数が減ると、GCの負荷が下がる
- 長時間稼働するサーバーでは重要

---

## 各ベンチマークで学ぶこと

### 1. IsPalindrome（回文判定）

**改善内容**: 文字列連結 → 両端比較

```bash
go test -bench=IsPalindrome -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**学び**:
- 文字列の `+=` は遅い（毎回新しいメモリを確保）
- アルゴリズムの工夫で大幅に改善

---

### 2. CountWords（単語カウント）

**改善内容**: `strings.Fields` → 1文字ずつ走査

```bash
go test -bench=CountWords -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**学び**:
- 標準ライブラリは便利だが、必ずしも最速ではない
- 用途に特化した実装の方が速い場合もある

---

### 3. StringConcat（文字列連結）

**改善内容**: `+=` → `strings.Builder` → `strings.Join`

```bash
go test -bench=StringConcat -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**学び**:
- `strings.Builder` は文字列連結の定番
- `strings.Join` はさらに高速（特化した実装）
- **10倍〜100倍の差**が出る

---

### 4. FindMax（最大値探索）

**改善内容**: バブルソート（O(n²)） → 1回走査（O(n)）

```bash
go test -bench=FindMax -benchmem day57_benchmark_test.go day57_benchmark_target.go
```

**学び**:
- アルゴリズムの計算量が最も重要
- O(n²) → O(n) は劇的な改善（**100倍以上**）
- データサイズが大きいほど差が顕著

---

## パフォーマンス改善のサイクル

### 1. 計測

```bash
go test -bench=. -benchmem > before.txt
```

### 2. 分析

- どこが遅いか？
- なぜ遅いか？
- どう改善できるか？

### 3. 改善

- コードを修正
- アルゴリズムを変更
- データ構造を見直し

### 4. 再計測

```bash
go test -bench=. -benchmem > after.txt
benchstat before.txt after.txt
```

### 5. 検証

- 機能テストで正しさを確認
- 改善効果を確認

### 6. 繰り返し

---

## 実務での使いどころ

### 使うべき場面

✅ **パフォーマンスが重要な関数**
- APIのホットパス（頻繁に呼ばれる処理）
- データ処理のループ
- 暗号化/シリアライズ

✅ **リファクタリング時**
- 変更前後でパフォーマンス劣化がないことを確認

✅ **アルゴリズム選定**
- 複数の実装を比較
- データサイズに応じた選択

✅ **ライブラリ評価**
- 複数のライブラリを比較

### 使わなくていい場面

❌ **パフォーマンスが重要でない箇所**
- 初期化処理（起動時に1回だけ）
- エラーハンドリング（頻繁には起きない）

❌ **早すぎる最適化**
- まず動くものを作る
- ボトルネックが明確になってから最適化

---

## トラブルシューティング

### ベンチマーク結果がばらつく

```bash
# 複数回実行して平均を取る
go test -bench=. -count=10

# 実行時間を長くする
go test -bench=. -benchtime=5s
```

### ベンチマークが速すぎる

- コンパイラが最適化している可能性
- 結果を使う（グローバル変数に代入）

### メモリ統計が表示されない

```bash
# -benchmem フラグを付ける
go test -bench=. -benchmem
```

---

## 学習のポイント

### 1. 推測ではなく計測

- 「このコードは遅いはず」という推測は当てにならない
- 必ず計測して判断する

### 2. 改善前後を比較

- ベンチマークは比較が重要
- 改善効果を数値で示す

### 3. メモリも重要

- 実行時間だけでなく、メモリ使用量も見る
- GCの負荷を減らすことも重要

### 4. データサイズを変える

- 小さいデータと大きいデータで結果が変わる
- 実際の使用例に近いサイズでテストする

### 5. 正しさが最優先

- パフォーマンスより正しさが重要
- 通常のテストで正しさを保証してから最適化

---

**完了条件**:
- ベンチマークの基本的な実行方法を理解
- ベンチマーク結果の読み方を理解
- 改善前後の比較方法を理解
- メモリ統計の見方を理解
- パフォーマンス改善のサイクルを理解

---

**次のステップ（Day 58）**: コードレビュー観点表（FW視点も追加）
