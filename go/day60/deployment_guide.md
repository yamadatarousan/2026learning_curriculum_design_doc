# デプロイ・ロールバック手順書

## 概要

このドキュメントは、TODO APIのデプロイとロールバックの手順を記載しています。

### 前提条件

- Gitリポジトリへのアクセス権限
- 本番環境へのSSH/デプロイ権限
- Docker / docker-composeがインストール済み
- データベースマイグレーションツールが利用可能

### デプロイの種類

| 種類 | 説明 | リスク | 使用タイミング |
|------|------|--------|----------------|
| **ローリングデプロイ** | 1台ずつ順次更新 | 低 | 通常のリリース |
| **Blue-Green** | 環境を丸ごと切り替え | 低 | 大規模変更 |
| **Canary** | 一部トラフィックのみ新版 | 最低 | 高リスク変更 |
| **ホットフィックス** | 緊急パッチ適用 | 中 | 重大なバグ修正 |

---

## デプロイフロー全体図

```
[開発] → [レビュー] → [マージ] → [CI実行] → [ステージング] → [本番デプロイ] → [監視]
                                                                    ↓（問題発生時）
                                                              [ロールバック]
```

---

## 1. デプロイ前チェックリスト

### 必須確認事項

- [ ] PRがレビュー承認済み
- [ ] CIが全てグリーン（lint/test/build）
- [ ] ステージング環境で動作確認済み
- [ ] マイグレーションが必要な変更の有無を確認
- [ ] 破壊的変更（Breaking Changes）の有無を確認
- [ ] ロールバック手順を確認済み
- [ ] デプロイ時間帯が適切（推奨：平日10-16時）
- [ ] 関係者への事前通知済み（大規模変更の場合）

### リスク評価

以下の質問に答えて、デプロイのリスクレベルを判断：

1. DBスキーマ変更を含むか？ → **YES: 高リスク**
2. 認証/認可ロジックの変更か？ → **YES: 高リスク**
3. 外部API連携の変更か？ → **YES: 中リスク**
4. UIのみの変更か？ → **YES: 低リスク**

**高リスクの場合**: Canaryデプロイまたはメンテナンス時間を設定

---

## 2. デプロイ手順（Docker + ローリングデプロイ）

### Step 1: リリースタグを作成

```bash
# 1. mainブランチを最新にする
git checkout main
git pull origin main

# 2. バージョン番号を決定（セマンティックバージョニング）
# 例: v1.2.3 → v1.2.4 (パッチ)
#     v1.2.3 → v1.3.0 (マイナー)
#     v1.2.3 → v2.0.0 (メジャー)

# 3. タグを作成・プッシュ
git tag -a v1.2.4 -m "Release v1.2.4: Fix todo creation bug"
git push origin v1.2.4

# 4. タグが正しく作成されたことを確認
git tag -l | tail -5
```

### Step 2: Dockerイメージをビルド

```bash
# 1. リポジトリをクローン or プル
cd /path/to/todo-api
git fetch --tags
git checkout v1.2.4

# 2. イメージをビルド
docker build -t todo-api:v1.2.4 -t todo-api:latest .

# 3. ビルドが成功したことを確認
docker images | grep todo-api

# 4. イメージをレジストリにプッシュ（Docker Hubやプライベートレジストリ）
docker push todo-api:v1.2.4
docker push todo-api:latest
```

### Step 3: データベースマイグレーション（必要な場合）

**重要**: マイグレーションは必ずアプリケーションデプロイ**前**に実行

```bash
# 1. バックアップを取得（必須）
pg_dump -h localhost -U user -d todo_db > backup_$(date +%Y%m%d_%H%M%S).sql

# 2. マイグレーションファイルを確認
ls -lh migrations/

# 3. ドライラン（可能な場合）
# 実際には実行せず、SQL文を確認
# ツールによって方法は異なる

# 4. マイグレーション実行
# 例: golang-migrate
migrate -path ./migrations -database "postgres://user:password@localhost:5432/todo_db?sslmode=disable" up

# 5. マイグレーション結果を確認
psql -h localhost -U user -d todo_db -c "\dt"
psql -h localhost -U user -d todo_db -c "SELECT version FROM schema_migrations ORDER BY version DESC LIMIT 5;"
```

**マイグレーションが失敗した場合**:
1. すぐにロールバック（`migrate down`）
2. バックアップから復元
3. デプロイを中止

### Step 4: アプリケーションをデプロイ

#### パターンA: Docker Composeを使用

```bash
# 1. 本番サーバーにSSH
ssh user@production-server

# 2. docker-compose.ymlを更新（イメージタグを指定）
cd /opt/todo-api
vim docker-compose.yml
# services.app.image を todo-api:v1.2.4 に変更

# 3. 新しいイメージをプル
docker-compose pull app

# 4. アプリケーションを再起動（ダウンタイム有り）
docker-compose up -d app

# 5. 起動確認
docker-compose ps
docker-compose logs -f app

# 6. ヘルスチェック
curl http://localhost:8080/health
```

#### パターンB: 複数サーバーでのローリングデプロイ

```bash
# サーバー1台ずつ順次実行

for server in server1 server2 server3; do
  echo "Deploying to $server..."

  # 1. ロードバランサーから切り離し
  # 例: AWS ALBの場合
  # aws elbv2 deregister-targets --target-group-arn <arn> --targets Id=$server

  # 2. デプロイ実行
  ssh user@$server << 'EOF'
    cd /opt/todo-api
    docker-compose pull app
    docker-compose up -d app
    sleep 10
EOF

  # 3. ヘルスチェック
  health_status=$(ssh user@$server "curl -s http://localhost:8080/health | jq -r .status")

  if [ "$health_status" != "ok" ]; then
    echo "Health check failed on $server. Aborting deployment."
    exit 1
  fi

  # 4. ロードバランサーに戻す
  # aws elbv2 register-targets --target-group-arn <arn> --targets Id=$server

  echo "Deployed to $server successfully."
  sleep 30  # 次のサーバーに進む前に少し待機
done

echo "Rolling deployment completed!"
```

### Step 5: デプロイ後の動作確認

```bash
# 1. ヘルスチェック
curl http://localhost:8080/health

# 2. 主要エンドポイントのスモークテスト
# ログイン
TOKEN=$(curl -s -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"testpass123"}' \
  | jq -r .token)

# TODO一覧取得
curl -s http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer $TOKEN" | jq

# TODO作成
curl -s -X POST http://localhost:8080/api/v1/todos \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"デプロイ確認用"}' | jq

# 3. ログ確認（エラーがないか）
docker-compose logs app --tail 50 | grep ERROR

# 4. メトリクスを確認
# - レスポンスタイム
# - エラー率
# - CPU/メモリ使用率
```

### Step 6: 監視（デプロイ後15分）

デプロイ後は必ず**15分以上監視**してください：

- [ ] エラー率が正常範囲内（< 0.1%）
- [ ] レスポンスタイムが正常範囲内（P95 < 500ms）
- [ ] CPU/メモリ使用率が急増していない
- [ ] アラートが発火していない
- [ ] ユーザーからの問い合わせがない

**異常があればすぐにロールバック**

---

## 3. ロールバック手順

### ロールバックを実行すべき状況

- エラー率が通常の10倍以上
- 重大なバグが見つかった
- パフォーマンスが著しく劣化
- データ不整合が発生

### 緊急ロールバック（5分以内）

```bash
# 1. 前バージョンのタグを確認
git tag -l | tail -10
# 例: v1.2.3 に戻す

# 2. docker-compose.ymlを修正
cd /opt/todo-api
vim docker-compose.yml
# services.app.image を todo-api:v1.2.3 に変更

# 3. 再起動
docker-compose pull app
docker-compose up -d app

# 4. ヘルスチェック
curl http://localhost:8080/health

# 5. 動作確認
# 主要エンドポイントをテスト

# 6. 関係者に通知
echo "Rollback completed. Reverted to v1.2.3" | notify
```

### データベースマイグレーションのロールバック

**注意**: DBロールバックはデータ損失のリスクがあります。慎重に実行してください。

```bash
# 1. 現在のマイグレーションバージョンを確認
migrate -path ./migrations -database "postgres://..." version

# 2. 1つ前のバージョンにロールバック
migrate -path ./migrations -database "postgres://..." down 1

# 3. 結果を確認
psql -h localhost -U user -d todo_db -c "\dt"

# データが失われた場合はバックアップから復元
psql -h localhost -U user -d todo_db < backup_20240101_120000.sql
```

### ロールバック後の対応

- [ ] 関係者に通知（ロールバック完了を報告）
- [ ] 原因を調査（ログ、エラーメッセージを収集）
- [ ] インシデントチケット作成
- [ ] ポストモーテムをスケジュール

---

## 4. Blue-Greenデプロイ（推奨：大規模変更時）

Blue-Green デプロイは、2つの本番環境（BlueとGreen）を用意し、トラフィックを切り替える方法です。

### メリット
- ダウンタイムゼロ
- 問題発生時に即座にロールバック可能
- 新環境で十分にテストしてから切り替え可能

### 手順

```bash
# 前提: Blue環境（現行）が稼働中、Green環境（新版）を準備

# 1. Green環境に新バージョンをデプロイ
ssh user@green-server
cd /opt/todo-api
docker-compose pull app
docker-compose up -d app

# 2. Green環境で動作確認
curl http://green-server:8080/health
# 主要エンドポイントをテスト

# 3. ロードバランサーでトラフィックを切り替え
# 例: AWS ALB、Nginx、HAProxyなど
# Blue → Green へトラフィックを向ける

# 4. 監視（15分）
# メトリクスを確認、異常があればBlueに戻す

# 5. 問題なければBlue環境を停止
ssh user@blue-server
docker-compose down
```

---

## 5. Canaryデプロイ（推奨：高リスク変更時）

Canaryデプロイは、新バージョンに少量のトラフィック（例: 10%）を流し、問題ないことを確認してから全体に展開する方法です。

### 手順

```bash
# 1. 新バージョンを1台だけデプロイ
ssh user@canary-server
cd /opt/todo-api
docker-compose pull app
docker-compose up -d app

# 2. ロードバランサーで10%のトラフィックをCanaryサーバーに向ける
# 例: Nginxの設定
# upstream backend {
#   server blue-server:8080 weight=9;
#   server canary-server:8080 weight=1;
# }

# 3. Canaryサーバーのメトリクスを監視（30分）
# エラー率、レスポンスタイム、ログを確認

# 4. 問題なければトラフィックを段階的に増やす
# 10% → 25% → 50% → 100%

# 5. 全サーバーに展開
# ローリングデプロイで残りのサーバーを更新
```

---

## 6. ホットフィックス手順（緊急パッチ）

重大なバグが本番で見つかった場合の緊急対応手順です。

### 前提
- 通常のレビュープロセスをスキップ可能（ただし事後報告必須）
- ステージング確認を簡略化可能
- リスクを理解した上で実行

### 手順

```bash
# 1. mainブランチから hotfix ブランチを作成
git checkout main
git pull origin main
git checkout -b hotfix/fix-critical-bug

# 2. バグ修正を実装・コミット
# ... コードを修正 ...
git add .
git commit -m "hotfix: Fix critical bug in todo creation"

# 3. テストを実行（最低限）
go test ./...

# 4. mainブランチにマージ
git checkout main
git merge hotfix/fix-critical-bug
git push origin main

# 5. ホットフィックスタグを作成
git tag -a v1.2.4-hotfix.1 -m "Hotfix: Fix critical bug"
git push origin v1.2.4-hotfix.1

# 6. 通常のデプロイ手順に従ってリリース
# （Step 2以降を実行）

# 7. デプロイ後、すぐに動作確認

# 8. 関係者に報告
# - 何が起きたか
# - どう修正したか
# - 今後の対策
```

---

## 7. デプロイ自動化（CI/CD）

GitHub Actionsを使った自動デプロイの例：

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    tags:
      - 'v*.*.*'

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Build Docker image
        run: |
          docker build -t todo-api:${{ github.ref_name }} .
          docker tag todo-api:${{ github.ref_name }} todo-api:latest

      - name: Push to registry
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker push todo-api:${{ github.ref_name }}
          docker push todo-api:latest

      - name: Deploy to production
        uses: appleboy/ssh-action@master
        with:
          host: ${{ secrets.PROD_HOST }}
          username: ${{ secrets.PROD_USER }}
          key: ${{ secrets.SSH_PRIVATE_KEY }}
          script: |
            cd /opt/todo-api
            docker-compose pull app
            docker-compose up -d app
            sleep 10
            curl -f http://localhost:8080/health || exit 1

      - name: Notify on failure
        if: failure()
        run: |
          # Slackなどに通知
          echo "Deployment failed!"
```

---

## 8. トラブルシューティング

### デプロイが失敗した場合

| 症状 | 原因 | 対応 |
|------|------|------|
| イメージがプルできない | レジストリ接続エラー | 認証情報確認、ネットワーク確認 |
| コンテナが起動しない | 設定ミス、依存関係不足 | ログ確認、設定ファイル確認 |
| ヘルスチェック失敗 | DB接続失敗、初期化エラー | DBの状態確認、環境変数確認 |
| マイグレーション失敗 | SQL文エラー、制約違反 | ロールバック、マイグレーションファイル確認 |

### デプロイ後にパフォーマンスが悪化した場合

```bash
# 1. 新旧バージョンのメトリクスを比較
# - レスポンスタイム
# - CPU/メモリ使用率
# - DB接続数

# 2. スロークエリログを確認
psql -h localhost -U user -d todo_db -c "SELECT query, calls, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"

# 3. 問題が大きければロールバック
# 小さければ監視を続けて次回修正
```

---

## 9. チェックリスト（印刷用）

### デプロイ前

- [ ] PRレビュー済み
- [ ] CI全てグリーン
- [ ] ステージング確認済み
- [ ] マイグレーション確認
- [ ] ロールバック手順確認
- [ ] 関係者通知済み

### デプロイ中

- [ ] バックアップ取得
- [ ] マイグレーション実行（必要な場合）
- [ ] アプリケーションデプロイ
- [ ] ヘルスチェック成功
- [ ] スモークテスト成功

### デプロイ後

- [ ] エラー率正常
- [ ] レスポンスタイム正常
- [ ] CPU/メモリ正常
- [ ] 15分監視完了
- [ ] デプロイ完了通知

---

## 10. 用語集

| 用語 | 説明 |
|------|------|
| **ローリングデプロイ** | サーバーを1台ずつ順次更新する方法 |
| **Blue-Green** | 2つの環境を用意し、トラフィックを切り替える方法 |
| **Canary** | 一部トラフィックのみ新版に流してテストする方法 |
| **ホットフィックス** | 緊急のバグ修正パッチ |
| **ロールバック** | 前のバージョンに戻すこと |
| **マイグレーション** | データベーススキーマの変更 |

---

## 更新履歴

| 日付 | 更新者 | 内容 |
|------|--------|------|
| 2024-01-01 | - | 初版作成 |
