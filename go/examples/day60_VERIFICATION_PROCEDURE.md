# Day60: 成果物の動作確認手順

## 概要

Day60では以下の4つの成果物を作成しました：

1. **GitHub Actions（CI/CD設定）** - `day60_github_actions.yml`
2. **API仕様（OpenAPI）** - `day60_openapi.yaml`
3. **Runbook** - `day60_runbook.md`
4. **デプロイ/ロールバック手順書** - `day60_deployment_guide.md`

このドキュメントでは、各成果物の動作確認方法を説明します。

---

## 前提条件

- Go 1.21以上がインストール済み
- Docker / docker-composeがインストール済み
- GitHubリポジトリへのアクセス権限（GitHub Actions確認用）
- OpenAPIツール（Swagger UIまたはRedocなど）が利用可能

---

## 1. GitHub Actions（CI/CD）の確認

### 目的
CI/CDパイプラインが正しく設定されているかを確認します。

### 手順

#### Step 1: GitHub Actionsワークフローファイルを配置

```bash
# プロジェクトルートに移動
cd /Users/user/Development/2026learning_curriculum_design_doc/go/day60

# .github/workflows ディレクトリを作成
mkdir -p .github/workflows

# ワークフローファイルをコピー
cp ../examples/day60_github_actions.yml .github/workflows/ci.yml
```

#### Step 2: ワークフローファイルの内容を確認

```bash
# ファイルが正しく配置されているか確認
ls -la .github/workflows/

# 内容を確認
cat .github/workflows/ci.yml
```

**確認ポイント**:
- [ ] `on.push.branches` が `main` と `develop` になっている
- [ ] `lint`、`test`、`build` の3つのジョブが定義されている
- [ ] `test` ジョブでPostgreSQLサービスが起動している
- [ ] `go test -race` でレースコンディションチェックが有効

#### Step 3: GitHubにプッシュして動作確認

```bash
# 変更をコミット
git add .github/workflows/ci.yml
git commit -m "Add CI workflow for Day60"

# リモートリポジトリにプッシュ
git push origin main
```

#### Step 4: GitHub Actionsの実行結果を確認

1. GitHubリポジトリのページを開く
2. **Actions** タブをクリック
3. 最新のワークフロー実行を確認

**期待される結果**:
- [ ] `lint` ジョブが成功（緑色のチェックマーク）
- [ ] `test` ジョブが成功
- [ ] `build` ジョブが成功
- [ ] すべてのジョブが5分以内に完了

**トラブルシューティング**:

| 問題 | 原因 | 対処法 |
|------|------|--------|
| lint失敗 | コードスタイル違反 | `golangci-lint run` を実行して修正 |
| test失敗 | テストコードのバグ | ログを確認してテストを修正 |
| build失敗 | コンパイルエラー | `go build` を実行してエラーを修正 |

#### Step 5: ローカルでもCI相当のチェックを実行

```bash
# lint実行
golangci-lint run --timeout=5m

# テスト実行（レースチェック付き）
go test -v -race -coverprofile=coverage.out ./...

# ビルド実行
go build -v -o ./bin/app ./main.go

# カバレッジ確認
go tool cover -html=coverage.out -o coverage.html
open coverage.html  # ブラウザで確認
```

**確認ポイント**:
- [ ] lintエラーが0件
- [ ] すべてのテストが成功
- [ ] レースコンディションが検出されない
- [ ] カバレッジが60%以上（目安）

---

## 2. API仕様（OpenAPI）の確認

### 目的
API仕様書が正しく記述され、READMEから参照できるかを確認します。

### 手順

#### Step 1: OpenAPIファイルをプロジェクトに配置

```bash
# docsディレクトリを作成
mkdir -p /Users/user/Development/2026learning_curriculum_design_doc/go/day60/docs

# OpenAPI仕様をコピー
cp ../examples/day60_openapi.yaml docs/openapi.yaml
```

#### Step 2: OpenAPI仕様の構文チェック

```bash
# OpenAPI Generatorをインストール（未インストールの場合）
# macOS
brew install openapi-generator

# または、Dockerを使用
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli validate -i /local/docs/openapi.yaml
```

**期待される結果**:
```
Validating spec (docs/openapi.yaml)
Spec is valid.
```

#### Step 3: Swagger UIで視覚的に確認

```bash
# Swagger UIをDockerで起動
docker run -p 8081:8080 -e SWAGGER_JSON=/docs/openapi.yaml -v ${PWD}/docs:/docs swaggerapi/swagger-ui

# ブラウザで確認
open http://localhost:8081
```

**確認ポイント**:
- [ ] すべてのエンドポイントが表示されている
  - `GET /health`
  - `POST /signup`
  - `POST /login`
  - `GET /api/v1/todos`
  - `POST /api/v1/todos`
  - `GET /api/v1/admin/users`
- [ ] リクエスト/レスポンスのスキーマが正しい
- [ ] 認証方式（JWT Bearer）が定義されている
- [ ] エラーレスポンスが統一されている

#### Step 4: READMEにAPI仕様へのリンクを追加

```bash
# READMEを編集
vim README.md
```

以下の内容を追加：

```markdown
## API仕様

API仕様書は [OpenAPI 3.0形式](./docs/openapi.yaml) で提供されています。

### 確認方法

#### Swagger UIで確認
```bash
docker run -p 8081:8080 -e SWAGGER_JSON=/docs/openapi.yaml -v ${PWD}/docs:/docs swaggerapi/swagger-ui
open http://localhost:8081
```

#### Redocで確認
```bash
docker run -p 8082:80 -e SPEC_URL=openapi.yaml -v ${PWD}/docs:/usr/share/nginx/html redocly/redoc
open http://localhost:8082
```
```

#### Step 5: APIクライアントコードを生成してテスト（オプション）

```bash
# TypeScript用のクライアントコードを生成
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli generate \
  -i /local/docs/openapi.yaml \
  -g typescript-axios \
  -o /local/generated/typescript-client

# 生成されたコードを確認
ls -la generated/typescript-client/
```

**確認ポイント**:
- [ ] クライアントコードがエラーなく生成される
- [ ] 型定義が正しく生成されている

---

## 3. Runbookの確認

### 目的
Runbookが実際の障害対応で使えるかを確認します。

### 手順

#### Step 1: Runbookをプロジェクトに配置

```bash
# docsディレクトリにコピー
cp ../examples/day60_runbook.md docs/runbook.md
```

#### Step 2: Runbookの内容を確認

```bash
# Markdownビューアで確認
open docs/runbook.md

# または、コマンドラインで確認
cat docs/runbook.md | less
```

**確認ポイント**:
- [ ] 障害パターンが4つ以上カバーされている
  - サーバーが起動しない
  - サービスが遅い
  - DBに繋がらない
  - 特定のエンドポイントが失敗
- [ ] 各パターンに「初動確認」と「復旧手順」が記載されている
- [ ] コマンド例が実行可能な形で記載されている
- [ ] エスカレーション基準が明確

#### Step 3: Runbookのコマンドを実際に試す

**シナリオ1: サーバーが起動しているか確認**

```bash
# プロセス確認
ps aux | grep app

# コンテナ確認（Docker環境）
docker ps -a

# ヘルスチェック
curl http://localhost:8080/health
```

**期待される結果**:
- [ ] プロセスまたはコンテナが実行中
- [ ] ヘルスチェックが成功（`{"status":"ok"}`）

**シナリオ2: DBに繋がるか確認**

```bash
# DB接続テスト
psql -h localhost -U user -d todo_db -c "SELECT 1;"

# Docker環境の場合
docker exec -it <postgres_container> psql -U user -d todo_db -c "SELECT 1;"
```

**期待される結果**:
- [ ] DB接続が成功
- [ ] `1` が返ってくる

**シナリオ3: ログ確認**

```bash
# アプリケーションログ
docker logs <container_id> --tail 100

# エラーログのみ
docker logs <container_id> | grep ERROR
```

**期待される結果**:
- [ ] ログが正常に出力されている
- [ ] 致命的なエラーがない

#### Step 4: READMEにRunbookへのリンクを追加

```markdown
## 障害対応

障害が発生した際は [Runbook](./docs/runbook.md) を参照してください。

主な障害パターン:
- [サーバーが起動しない](./docs/runbook.md#1-サーバーが起動しない--サービスが落ちている)
- [サービスが遅い](./docs/runbook.md#2-サービスが遅い--タイムアウトが発生している)
- [DBに繋がらない](./docs/runbook.md#3-dbに繋がらない)
```

---

## 4. デプロイ/ロールバック手順書の確認

### 目的
デプロイ手順が実行可能かを確認します。

### 手順

#### Step 1: 手順書をプロジェクトに配置

```bash
# docsディレクトリにコピー
cp ../examples/day60_deployment_guide.md docs/deployment_guide.md
```

#### Step 2: 手順書の内容を確認

```bash
# Markdownビューアで確認
open docs/deployment_guide.md
```

**確認ポイント**:
- [ ] デプロイフロー全体図が記載されている
- [ ] デプロイ前チェックリストがある
- [ ] ローリングデプロイ手順が詳細に記載されている
- [ ] Blue-Green、Canaryデプロイの説明がある
- [ ] ロールバック手順が明確
- [ ] トラブルシューティングが充実している

#### Step 3: デプロイ手順を試す（ローカル環境でシミュレーション）

**シナリオ1: リリースタグの作成**

```bash
# 現在のタグを確認
git tag -l | tail -5

# 新しいタグを作成（テスト）
git tag -a v1.0.0-test -m "Test release tag for Day60"
git tag -l | tail -5

# タグを削除（テストなので）
git tag -d v1.0.0-test
```

**シナリオ2: Dockerイメージのビルド**

```bash
# イメージをビルド
docker build -t todo-api:day60-test .

# イメージが作成されたか確認
docker images | grep todo-api

# イメージを起動
docker run -d -p 8080:8080 --name todo-api-test todo-api:day60-test

# ヘルスチェック
curl http://localhost:8080/health

# コンテナを停止・削除
docker stop todo-api-test
docker rm todo-api-test
docker rmi todo-api:day60-test
```

**期待される結果**:
- [ ] イメージが正常にビルドされる
- [ ] コンテナが正常に起動する
- [ ] ヘルスチェックが成功

**シナリオ3: ロールバック手順の確認**

```bash
# 現在のバージョンを確認
docker ps --format "table {{.Image}}\t{{.Status}}"

# 前バージョンに戻す（シミュレーション）
echo "前バージョンのタグ: v0.9.9"
echo "docker-compose.ymlを編集してイメージタグをv0.9.9に変更"
echo "docker-compose up -d app"
```

#### Step 4: READMEにデプロイ手順へのリンクを追加

```markdown
## デプロイ

デプロイ手順は [デプロイガイド](./docs/deployment_guide.md) を参照してください。

### クイックスタート

```bash
# 1. リリースタグを作成
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 2. Dockerイメージをビルド
docker build -t todo-api:v1.0.0 .

# 3. デプロイ
docker-compose up -d app

# 4. ヘルスチェック
curl http://localhost:8080/health
```

### ロールバック

問題が発生した場合は [ロールバック手順](./docs/deployment_guide.md#3-ロールバック手順) を参照してください。
```

---

## 5. 総合動作確認

### 目的
すべての成果物が統合されて機能するかを確認します。

### 手順

#### Step 1: プロジェクト構成を確認

```bash
cd /Users/user/Development/2026learning_curriculum_design_doc/go/day60

# ディレクトリ構成を確認
tree -L 2 -I 'vendor|node_modules'
```

**期待される構成**:
```
.
├── .github
│   └── workflows
│       └── ci.yml
├── docs
│   ├── openapi.yaml
│   ├── runbook.md
│   └── deployment_guide.md
├── main.go
├── repository.go
├── go.mod
├── go.sum
└── README.md
```

#### Step 2: README.mdの完成度を確認

```bash
cat README.md
```

**確認ポイント**:
- [ ] プロジェクト概要が記載されている
- [ ] セットアップ手順が記載されている
- [ ] API仕様へのリンクがある
- [ ] Runbookへのリンクがある
- [ ] デプロイ手順へのリンクがある
- [ ] CI/CDの説明がある

#### Step 3: すべてのリンクが正しいか確認

```bash
# READMEから相対パスのリンクを抽出
grep -o '\[.*\](\.\/.*\.md)' README.md

# 各リンク先のファイルが存在するか確認
test -f docs/openapi.yaml && echo "✓ openapi.yaml exists"
test -f docs/runbook.md && echo "✓ runbook.md exists"
test -f docs/deployment_guide.md && echo "✓ deployment_guide.md exists"
test -f .github/workflows/ci.yml && echo "✓ ci.yml exists"
```

**期待される結果**:
```
✓ openapi.yaml exists
✓ runbook.md exists
✓ deployment_guide.md exists
✓ ci.yml exists
```

#### Step 4: E2Eテスト（エンドツーエンド）

```bash
# 1. アプリケーションを起動
docker-compose up -d

# 2. ヘルスチェック
curl -f http://localhost:8080/health || echo "Health check failed"

# 3. ユーザー登録
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}' | jq

# 4. ログイン
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}' | jq -r .token)

echo "Token: $TOKEN"

# 5. TODO作成
curl -X POST http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Day60の動作確認"}' | jq

# 6. TODO一覧取得
curl http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer $TOKEN" | jq

# 7. 後片付け
docker-compose down
```

**期待される結果**:
- [ ] すべてのAPIが正常にレスポンスを返す
- [ ] ユーザー登録が成功
- [ ] ログインが成功してトークンが取得できる
- [ ] TODO作成が成功
- [ ] TODO一覧取得が成功

---

## 6. 動作確認完了チェックリスト

### GitHub Actions

- [ ] ワークフローファイルが `.github/workflows/` に配置されている
- [ ] GitHubでワークフローが実行され、すべてのジョブが成功
- [ ] ローカルでもlint/test/buildが成功

### API仕様（OpenAPI）

- [ ] OpenAPI仕様ファイルが `docs/openapi.yaml` に配置されている
- [ ] 構文チェックが成功
- [ ] Swagger UIで仕様が正しく表示される
- [ ] READMEから仕様へのリンクがある

### Runbook

- [ ] Runbookが `docs/runbook.md` に配置されている
- [ ] 主要な障害パターンがカバーされている
- [ ] コマンド例が実行可能
- [ ] READMEからRunbookへのリンクがある

### デプロイ/ロールバック手順書

- [ ] 手順書が `docs/deployment_guide.md` に配置されている
- [ ] デプロイフローが明確
- [ ] ロールバック手順が詳細
- [ ] READMEから手順書へのリンクがある

### 総合

- [ ] プロジェクト構成が正しい
- [ ] README.mdがすべての成果物へリンクしている
- [ ] E2Eテストが成功
- [ ] すべてのリンクが正しく機能している

---

## 7. トラブルシューティング

### 問題: GitHub Actionsが失敗する

**原因**: ワークフローファイルの構文エラー

**対処法**:
```bash
# YAMLの構文チェック
yamllint .github/workflows/ci.yml

# オンラインバリデーターを使用
# https://www.yamllint.com/
```

### 問題: OpenAPI仕様のバリデーションエラー

**原因**: スキーマ定義の不備

**対処法**:
```bash
# 詳細なエラーメッセージを確認
docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli validate -i /local/docs/openapi.yaml --recommend

# オンラインエディタで確認
# https://editor.swagger.io/
```

### 問題: E2Eテストが失敗する

**原因**: データベース未起動、環境変数未設定

**対処法**:
```bash
# データベースの状態を確認
docker-compose ps

# ログを確認
docker-compose logs db
docker-compose logs app

# 環境変数を確認
docker-compose config
```

---

## 8. 次のステップ

Day60の動作確認が完了したら：

1. **成果物をGitHubにプッシュ**
   ```bash
   git add .
   git commit -m "Complete Day60: CI, API docs, Runbook, Deployment guide"
   git push origin main
   ```

2. **チームメンバーにレビューを依頼**
   - API仕様が要件を満たしているか
   - Runbookが実際の障害対応で使えるか
   - デプロイ手順が明確か

3. **Day61以降の準備**
   - Day60までの学習内容を振り返る
   - 次のフェーズ（実践的なアプリケーション構築）に進む

---

## まとめ

Day60では「現場で戦えるAPI」を完成させるための運用ドキュメントを整備しました：

- **GitHub Actions**: 自動化されたCI/CDパイプライン
- **API仕様**: 標準化されたOpenAPI形式のドキュメント
- **Runbook**: 障害対応の初動手順
- **デプロイ手順**: 本番運用を想定した手順書

これらの成果物により、あなたのTODO APIは「動くだけ」から「運用に耐える」レベルに進化しました。

次のステップでは、これらの知識を活かして実際のプロジェクトを構築していきます！
