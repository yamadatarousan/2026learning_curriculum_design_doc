# 追加30日（Day 31–60）：Goで「現場で戦えるAPI」へ育てる

前提：Day1–30で作ったTODO API（Go）をベースに、**1日1〜2時間**で「運用・保守に耐える」形へ近づけます。  
※毎日、手を動かした結果として **必ず成果物が1つ以上残る**ようにしています。

---

## Week 6：現場っぽいAPIの土台（運用・保守に耐える形）
**目標:** “動く”から “運用できる” へ（ログ/設定/終了処理/エラー設計）。

### Day 31: プロジェクト構成の再整理
- 学習項目: `cmd/`, `internal/` の分割、依存方向、責務の切り方
- 成果物: ディレクトリ再構成 + READMEに構成方針追記

### Day 32: 設定管理（env / flags）
- 学習項目: `os.Getenv`、flag、設定の優先順位（env>flag>default）
- 成果物: `config` パッケージ導入（DB接続等をここに寄せる）

### Day 33: Contextの実戦
- 学習項目: `context.Context` の伝播、タイムアウト、キャンセル
- 成果物: DBアクセス/外部呼び出しに `ctx` を通す

### Day 34: 構造化ログ
- 学習項目: `log/slog` などで key/value ログ、request_id
- 成果物: リクエスト単位のログ（request_id付き）

### Day 35: エラー設計（APIの失敗を統一）
- 学習項目: エラー型、HTTPステータス、エラーレスポンスの統一
- 成果物: エラーレスポンス仕様（JSON）+ 実装

### Day 36: Middleware（横断関心の分離）
- 学習項目: ログ、認証前提の枠、panic recovery
- 成果物: recovery middleware + logging middleware

### Day 37: Graceful shutdown
- 学習項目: `http.Server` / `Shutdown(ctx)`、シグナル処理
- 成果物: Ctrl+Cでも安全に終了するサーバ

---

## Week 7：DBを「現場のDB」に寄せる（マイグレーション/トランザクション/設計）
**目標:** DB操作を “それっぽく” する（移行・ロック・整合性の入口）。

### Day 38: SQLite→PostgreSQLへ移行（ローカルでOK）
- 学習項目: docker-composeでPostgres、接続文字列
- 成果物: compose追加 + 接続先をPostgresへ

### Day 39: マイグレーション導入
- 学習項目: マイグレーションツール（何でもOK）と運用イメージ
- 成果物: `migrations/` 作成、`up/down` 手順をREADMEへ

### Day 40: インデックスとクエリ（超入門）
- 学習項目: どこにindex貼るか、`EXPLAIN` の入口
- 成果物: TODO一覧のクエリに対する index 追加 + メモ

### Day 41: トランザクション（必須の型）
- 学習項目: `BEGIN/COMMIT/ROLLBACK`、一貫性が必要な場面
- 成果物: 何か1つ “Txが必要な処理” をTx化

### Day 42: Repository層の整理
- 学習項目: handler→service→repo の薄い分離（やりすぎない）
- 成果物: DBアクセスをrepoに寄せる

### Day 43: 冪等性の考え方（DB寄り）
- 学習項目: 二重送信、ユニーク制約、再試行
- 成果物: “二重作成を防ぐ” 方針を1つ入れる（制約 or 実装）

### Day 44: 週次まとめ（DB運用の足場）
- 学習項目: “移行手順/戻し手順” を文章化する
- 成果物: READMEに「DB初期化/移行/戻す」を追記

---

## Week 8：認証・認可（現場で要求されやすい）
**目標:** “誰が何できるか” をコードとテストに落とす。

### Day 45: 認証方式を決める（方針だけでOK）
- 学習項目: Cookieセッション / JWT の違い、選定軸
- 成果物: READMEに「今回の方式」を明文化

### Day 46: ユーザー登録・ログイン（最小）
- 学習項目: パスワードハッシュ（bcrypt等）、認証失敗の扱い
- 成果物: `/signup` `/login` の最小実装

### Day 47: 認証middleware
- 学習項目: 認証情報を `context` に載せる
- 成果物: 認証必須APIの導入（TODO作成など）

### Day 48: 認可（RBAC最小）
- 学習項目: role（admin/user）で守る
- 成果物: admin専用APIを1つ作る

### Day 49: マルチテナントの入口（概要でOK）
- 学習項目: tenant_id、データ分離の考え方
- 成果物: “やるならこうする” のメモ（実装は軽くてもOK）

### Day 50: セキュリティの最低限
- 学習項目: 入力検証、CORS、Secrets（env）、ログに出してはいけない情報
- 成果物: CORS設定 + Secrets方針メモ

### Day 51: 週次まとめ（権限事故を防ぐ）
- 学習項目: 権限表（誰が何できるか）を1枚にする
- 成果物: 権限表 + 認証/認可のテスト追加

---

## Week 9：テスト戦略・品質ゲート（足切りを超える）
**目標:** “テスト書けます” ではなく “壊れにくい運用” に寄せる。

### Day 52: httptestでハンドラ結合テスト
- 学習項目: `httptest.NewServer` / request組み立て
- 成果物: APIの結合テスト数本

### Day 53: DB込みの統合テスト
- 学習項目: テスト用DBの立ち上げ方（compose使い回しでOK）
- 成果物: DB込みのテスト1本（migrate→test→cleanup）

### Day 54: テストデータ設計（fixture/seed）
- 学習項目: 何を固定し何をランダムにするか
- 成果物: `testdata` / seed関数整備

### Day 55: race detector
- 学習項目: `go test -race` の意味、どんなバグが出るか
- 成果物: CIでも `-race` を回す方針メモ（実装は任意）

### Day 56: fuzz / property の入口（軽くでOK）
- 学習項目: Goのfuzzの使いどころ（入力境界）
- 成果物: 小さな関数にfuzzテスト1つ

### Day 57: ベンチの入口（計測癖）
- 学習項目: `testing.B`、ボトルネック探しの入口
- 成果物: 1ベンチ + 計測メモ

### Day 58: コードレビュー観点表（セルフ運用）
- 学習項目: Correctness/Design/Readability/Test/Security/Perf
- 成果物: 観点表をREADMEに置く + それで自分のPRを直す

---

## Week 10：DevOps/運用（ローカル中心で“現場っぽさ”）
**目標:** “ローカルで動く” を “再現可能に動く” へ（Docker/CI/Runbook）。

### Day 59: Docker化（マルチステージ）
- 学習項目: 小さいイメージ、実行ユーザー、環境変数
- 成果物: Dockerfile（マルチステージ）+ 起動手順

### Day 60: CI + リリース手順 + Runbook（最終まとめ）
- 学習項目: CIで lint/test、リリース時の手順、障害時の最初の動き
- 成果物:
  - GitHub Actions（lint/test）
  - READMEに「デプロイ/ロールバック手順（演習レベル）」追記
  - Runbook雛形（落ちた/遅い/DB繋がらない の初動）

---
