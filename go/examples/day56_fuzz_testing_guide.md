# Fuzz Testing 使用ガイド

## Fuzz Testing とは

**Fuzz Testing（ファジングテスト）** は、ランダムまたは変異させた入力を大量に生成してプログラムに与え、予期しない動作やバグを発見するテスト手法です。

### 通常のテストとの違い

| 項目 | 通常のテスト | Fuzz Testing |
|------|-------------|--------------|
| 入力 | 開発者が定義した固定値 | ランダム/自動生成 |
| 目的 | 期待する動作の確認 | 予期しない入力でのクラッシュ検出 |
| カバレッジ | 想定内のケース | 想定外のケースも含む |
| 実行時間 | 短い（秒〜分） | 長い（分〜時間〜継続実行） |

### Fuzz Testingが得意なこと

1. **予期しない入力でのバグ**
   - 境界値、空文字、特殊文字、巨大な入力
   - 開発者が想定していないパターン

2. **セキュリティ脆弱性**
   - バッファオーバーフロー
   - パニック（クラッシュ）
   - メモリ破壊

3. **入力検証の抜け漏れ**
   - バリデーション不足
   - エラーハンドリング漏れ

4. **算術エラー**
   - オーバーフロー/アンダーフロー
   - ゼロ除算

---

## Go 1.18以降のネイティブFuzzing

Go 1.18から、標準ライブラリでfuzz testingがサポートされています。

### 基本的な構造

```go
func FuzzXxx(f *testing.F) {
    // 1. Seed corpus（初期テストケース）を追加
    f.Add("valid input")
    f.Add("edge case")

    // 2. Fuzz関数でテストロジックを定義
    f.Fuzz(func(t *testing.T, input string) {
        // テスト対象の関数を実行
        result, err := YourFunction(input)

        // 不変条件をチェック
        if err == nil && result == nil {
            t.Error("result should not be nil when no error")
        }
    })
}
```

### Seed Corpus（初期テストケース）

Seed corpusは、fuzzingのスタート地点となる入力データです。

**役割**:
- fuzz engineが変異させる元データ
- 重要なエッジケースを明示的に含める
- コードカバレッジを向上させる

**良いseed corpusの例**:
```go
f.Add("") // 空文字
f.Add("normal input") // 正常系
f.Add(strings.Repeat("a", 1000)) // 長い入力
f.Add("特殊文字!@#$%") // 特殊文字
f.Add("\x00\xFF") // バイナリデータ
```

---

## Fuzz Testの書き方

### パターン1: パニックしないことを確認

最も基本的なfuzz test - クラッシュしないことを確認する

```go
func FuzzParseInput(f *testing.F) {
    f.Add("test")

    f.Fuzz(func(t *testing.T, input string) {
        // パニックしなければOK
        _, _ = ParseInput(input)
    })
}
```

### パターン2: 不変条件のチェック

エラーがない場合の条件をチェックする

```go
func FuzzParseEmail(f *testing.F) {
    f.Add("user@example.com")

    f.Fuzz(func(t *testing.T, email string) {
        username, domain, err := ParseEmail(email)

        // エラーがない場合の不変条件
        if err == nil {
            // usernameとdomainは空であってはならない
            if username == "" || domain == "" {
                t.Errorf("empty result without error for %q", email)
            }

            // 再構築したメールが妥当か
            reconstructed := username + "@" + domain
            if strings.Count(reconstructed, "@") != 1 {
                t.Errorf("invalid reconstructed email: %q", reconstructed)
            }
        }
    })
}
```

### パターン3: プロパティベーステスト

入出力の関係性をチェックする

```go
func FuzzReverseString(f *testing.F) {
    f.Add("hello")

    f.Fuzz(func(t *testing.T, input string) {
        reversed := ReverseString(input)

        // プロパティ1: 長さは同じ
        if len(reversed) != len(input) {
            t.Errorf("length mismatch: input=%d, reversed=%d", len(input), len(reversed))
        }

        // プロパティ2: 2回reverseすると元に戻る
        doubleReversed := ReverseString(reversed)
        if doubleReversed != input {
            t.Errorf("double reverse failed: %q != %q", doubleReversed, input)
        }
    })
}
```

### パターン4: 複数の引数をfuzz

```go
func FuzzCalculateDiscount(f *testing.F) {
    f.Add(1000, 10)

    f.Fuzz(func(t *testing.T, price int, discount int) {
        result, err := CalculateDiscount(price, discount)

        // 有効な入力の場合の不変条件
        if price >= 0 && discount >= 0 && discount <= 100 && err == nil {
            // 結果は元の価格以下
            if result > price {
                t.Errorf("discounted price %d > original %d", result, price)
            }

            // 結果は0以上
            if result < 0 {
                t.Errorf("negative result: %d", result)
            }
        }
    })
}
```

---

## Fuzz Testの実行方法

### 基本的な実行

```bash
# すべてのfuzz testを実行（デフォルト: 数秒間）
go test -fuzz=.

# 特定のfuzz testを実行
go test -fuzz=FuzzParseEmail

# 実行時間を指定（10秒間）
go test -fuzz=FuzzParseEmail -fuzztime=10s

# 実行回数を指定（1000回）
go test -fuzz=FuzzParseEmail -fuzztime=1000x
```

### 並列実行

```bash
# 4つのワーカーで並列実行
go test -fuzz=FuzzParseEmail -parallel=4
```

### Corpusの保存と再利用

fuzz testingで見つかった興味深い入力は、`testdata/fuzz/FuzzXxx/` に自動保存されます。

```
your_package/
  testdata/
    fuzz/
      FuzzParseEmail/
        abc123...  # 自動生成されたcorpus
        def456...
```

次回のfuzz実行時、これらのcorpusが自動的に使用されます。

### Corpusのクリア

```bash
# testdataディレクトリを削除してクリーンスタート
rm -rf testdata/fuzz
```

---

## Fuzz Testが失敗した場合

### 失敗例の出力

```
--- FAIL: FuzzParseEmail (0.13s)
    --- FAIL: FuzzParseEmail (0.00s)
        fuzz_test.go:25: username is empty but no error returned for email: "@@@"

    Failing input written to testdata/fuzz/FuzzParseEmail/abc123def...
    To re-run:
    go test -run=FuzzParseEmail/abc123def...
```

### 対応手順

1. **失敗を再現**
   ```bash
   go test -run=FuzzParseEmail/abc123def...
   ```

2. **コードを修正**
   - バグを修正する
   - または、入力バリデーションを追加

3. **再度fuzz testを実行**
   ```bash
   go test -fuzz=FuzzParseEmail -fuzztime=30s
   ```

4. **修正を確認**
   - 保存されたcorpusで自動的に回帰テストされる

---

## Fuzz Testingのベストプラクティス

### 1. 対象関数の選定

**Fuzzingが効果的な関数**:
✅ 外部入力を受け取る関数（API、パーサー、バリデーター）
✅ 文字列操作、バイト操作
✅ 数値計算（オーバーフロー検出）
✅ エンコード/デコード（JSON、XML、Base64など）
✅ セキュリティ関連（認証、暗号化）

**Fuzzingが不向きな関数**:
❌ 副作用が大きい関数（DB書き込み、ファイル操作）
❌ 決定的でない関数（ランダム、時刻依存）
❌ 複雑すぎる関数（テストのフィードバックループが遅い）

### 2. 良いSeed Corpusを作る

```go
// 良い例: エッジケースを網羅
f.Add("")                      // 空
f.Add("a")                     // 最小
f.Add(strings.Repeat("x", 1000)) // 大きい
f.Add("valid@example.com")     // 正常系
f.Add("@@@")                   // 異常系
f.Add("\x00\xFF")              // バイナリ
f.Add("日本語@例.jp")           // マルチバイト
```

### 3. 不変条件を明確にする

```go
f.Fuzz(func(t *testing.T, input string) {
    result, err := Parse(input)

    // 不変条件の例:
    // - エラーがない場合、resultは有効
    // - エラーがある場合、resultはゼロ値
    // - panicしない
    // - 入出力の長さ関係
    // - プロパティ（冪等性、可逆性など）
})
```

### 4. CI/CDでの運用

**短時間fuzz（PR毎）**:
```bash
# 各fuzz testを10秒ずつ実行
go test -fuzz=. -fuzztime=10s
```

**長時間fuzz（Nightly Build）**:
```bash
# 各fuzz testを1時間ずつ実行
go test -fuzz=. -fuzztime=1h
```

**継続的fuzz（専用サーバー）**:
- OSS-Fuzz、Google Cloud Fuzzingなどのサービスを利用
- 24時間365日fuzzingを実行し続ける

### 5. 通常のテストと併用

```go
// 通常のテスト: 既知のケースを高速にチェック
func TestParseEmail(t *testing.T) {
    tests := []struct{
        input string
        want  bool
    }{
        {"user@example.com", true},
        {"invalid", false},
    }
    // ...
}

// Fuzz test: 未知のケースを探索
func FuzzParseEmail(f *testing.F) {
    // ...
}
```

---

## 実務での使用シーン

### 1. API入力バリデーション

```go
func FuzzValidateUserInput(f *testing.F) {
    f.Fuzz(func(t *testing.T, username, email, password string) {
        // バリデーション関数がpanicしないことを確認
        _ = ValidateUserInput(username, email, password)
    })
}
```

### 2. パーサー/デコーダー

```go
func FuzzJSONParser(f *testing.F) {
    f.Add(`{"key": "value"}`)

    f.Fuzz(func(t *testing.T, jsonStr string) {
        var result map[string]interface{}
        // 不正なJSONでpanicしないことを確認
        _ = json.Unmarshal([]byte(jsonStr), &result)
    })
}
```

### 3. セキュリティ関連

```go
func FuzzSanitizeHTML(f *testing.F) {
    f.Add("<script>alert('xss')</script>")

    f.Fuzz(func(t *testing.T, html string) {
        sanitized := SanitizeHTML(html)

        // サニタイズ後にスクリプトタグが含まれていないことを確認
        if strings.Contains(strings.ToLower(sanitized), "<script") {
            t.Errorf("script tag found in sanitized HTML: %q", sanitized)
        }
    })
}
```

---

## トラブルシューティング

### Fuzzingが遅い

- **Seed corpusを減らす**: 不要なseedを削除
- **並列度を上げる**: `-parallel=8` など
- **関数を軽量化**: 重い処理をモック化

### メモリを大量に消費

- **入力サイズを制限**: fuzz関数内で入力長をチェック
  ```go
  if len(input) > 10000 {
      return
  }
  ```

### 同じバグが繰り返し見つかる

- **既知のバグを修正**: コードを修正してから再実行
- **Corpusをクリア**: `rm -rf testdata/fuzz`

---

## 参考資料

- [Go公式ブログ: Fuzzing is Beta Ready](https://go.dev/blog/fuzz-beta)
- [Go公式ドキュメント: Fuzzing](https://go.dev/doc/fuzz/)
- [Tutorial: Getting started with fuzzing](https://go.dev/doc/tutorial/fuzz)

---

**最終更新**: 2026-01-04
