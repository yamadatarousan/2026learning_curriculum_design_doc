# Day 59: Docker化 実行手順

## 今日の成果物

1. `day59_Dockerfile` - マルチステージビルドのDockerfile
2. `day59_.dockerignore` - Dockerビルド時に除外するファイル
3. `day59_docker-compose.yml` - アプリケーション全体の構成
4. `day59_.env.example` - 環境変数の例
5. `day59_docker_guide.md` - Docker化の完全ガイド
6. このファイル - 実行手順書

---

## 前提条件

Dockerがインストールされていること：

```bash
# Docker のバージョン確認
docker --version

# Docker Compose のバージョン確認
docker compose version
```

---

## 実行手順

### 手順1: ファイルを配置

day54のプロジェクトにDockerファイルを配置します：

```bash
cd /Users/user/Development/2026learning_curriculum_design_doc/go/day54

# Dockerfileをコピー
cp ../examples/day59_Dockerfile ./Dockerfile

# .dockerignoreをコピー
cp ../examples/day59_.dockerignore ./.dockerignore

# docker-compose.ymlをコピー
cp ../examples/day59_docker-compose.yml ./docker-compose.yml

# .env.exampleをコピー
cp ../examples/day59_.env.example ./.env.example

# .envファイルを作成（.env.exampleをコピー）
cp .env.example .env
```

---

### 手順2: .envファイルを編集

```bash
# .envファイルを編集
vi .env
```

最低限、以下を設定：

```
JWT_SECRET=your-super-secret-jwt-key-change-this
```

---

### 手順3: Dockerイメージをビルド

```bash
cd /Users/user/Development/2026learning_curriculum_design_doc/go/day54

# イメージをビルド
docker build -t todo-app:latest .
```

**注目ポイント①: マルチステージビルドの動作**

ビルド中の出力を確認：

```
[builder 1/6] FROM docker.io/library/golang:1.21-alpine
[builder 2/6] RUN apk add --no-cache git
[builder 3/6] WORKDIR /app
[builder 4/6] COPY go.mod go.sum ./
[builder 5/6] RUN go mod download
[builder 6/6] COPY . .
[builder 7/7] RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

[stage-1 1/5] FROM docker.io/library/alpine:latest
[stage-1 2/5] RUN apk --no-cache add ca-certificates
[stage-1 3/5] RUN addgroup -g 1000 appgroup && adduser -D -u 1000 -G appgroup appuser
[stage-1 4/5] COPY --from=builder /app/server .
...
```

- `[builder ...]`: ビルドステージ
- `[stage-1 ...]`: 実行ステージ
- **ビルドステージの成果物（バイナリ）だけが最終イメージに含まれる**

---

### 手順4: イメージサイズを確認（重要な学び）

```bash
# ビルドしたイメージのサイズを確認
docker images todo-app:latest
```

**期待される結果**:

```
REPOSITORY   TAG       IMAGE ID       CREATED         SIZE
todo-app     latest    abc123def456   10 seconds ago  20-30MB
```

**確認すべきこと**:
✅ イメージサイズが20-30MB程度（マルチステージビルドなし だと1GB超）
✅ **約40-50分の1のサイズ削減**

**比較のため、シングルステージでビルドしてみる（オプション）**:

```dockerfile
# 一時的にDockerfileを以下に変更
FROM golang:1.21
WORKDIR /app
COPY . .
RUN go build -o server
CMD ["/app/server"]
```

```bash
docker build -t todo-app:single .
docker images | grep todo-app
```

**結果比較**:
- シングルステージ: 約1GB
- マルチステージ: 約20-30MB

---

### 手順5: イメージの詳細を確認

```bash
# レイヤーごとのサイズを確認
docker history todo-app:latest
```

**注目ポイント②: レイヤー構造**

```
IMAGE          CREATED         CREATED BY                                      SIZE
abc123def456   2 minutes ago   CMD ["/app/server"]                             0B
def456ghi789   2 minutes ago   EXPOSE map[8080/tcp:{}]                         0B
ghi789jkl012   2 minutes ago   USER appuser                                    0B
jkl012mno345   2 minutes ago   RUN /bin/sh -c chown -R appuser:appgroup...     0B
mno345pqr678   2 minutes ago   COPY /app/server . # buildkit                   15MB
...
```

**確認すべきこと**:
- `COPY /app/server` のレイヤーが最も大きい（バイナリサイズ）
- その他のレイヤーは数MB以下

---

### 手順6: Docker Composeでアプリケーション全体を起動

```bash
# データベースとアプリケーションを起動
docker compose up -d

# ログを確認
docker compose logs -f
```

**注目ポイント③: 起動順序の制御**

```
[+] Running 3/3
 ✔ Network day54_default  Created
 ✔ Container todo_db      Healthy    # データベースが先に起動
 ✔ Container todo_app     Started    # データベースの健全性チェック後に起動
```

**確認すべきこと**:
✅ データベースが先に起動し、healthyになる
✅ アプリケーションはデータベースのhealth checkが通ってから起動
✅ `depends_on` と `healthcheck` の連携

---

### 手順7: 動作確認

```bash
# ヘルスチェック
curl http://localhost:8080/health

# ユーザー登録
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"docker-test@example.com","password":"password123"}'

# ログイン
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"docker-test@example.com","password":"password123"}'
```

---

### 手順8: コンテナ内部の確認（セキュリティ）

```bash
# コンテナに入る
docker exec -it todo_app sh

# 実行ユーザーを確認
whoami
# 出力: appuser （非rootユーザー）

# プロセスを確認
ps aux
# 出力: appuser が /app/server を実行

# exit で抜ける
exit
```

**注目ポイント④: 非rootユーザー**

**確認すべきこと**:
✅ アプリケーションが `appuser` で実行されている（rootではない）
✅ セキュリティのベストプラクティスに従っている

---

### 手順9: 環境変数の確認

```bash
# コンテナの環境変数を確認
docker exec todo_app env | grep JWT_SECRET
docker exec todo_app env | grep DB_HOST
```

**注目ポイント⑤: 環境変数の注入**

**確認すべきこと**:
✅ `JWT_SECRET` が設定されている
✅ `DB_HOST=db` （docker-compose.ymlで定義したサービス名）
✅ `.env` ファイルから環境変数が読み込まれている

---

### 手順10: ログの確認

```bash
# アプリケーションのログ
docker compose logs app

# データベースのログ
docker compose logs db

# リアルタイムでログを確認
docker compose logs -f
```

---

### 手順11: 停止とクリーンアップ

```bash
# 停止
docker compose down

# ボリュームも削除（データベースのデータも削除）
docker compose down -v

# イメージも削除
docker rmi todo-app:latest
```

---

## 重要な学び

### 1. マルチステージビルドの効果

| 項目 | シングルステージ | マルチステージ | 削減率 |
|------|-----------------|---------------|--------|
| イメージサイズ | 約1GB | 約20-30MB | **97%削減** |
| 含まれる内容 | Go コンパイラ、ツール、ソース | バイナリのみ | - |
| セキュリティ | 攻撃面が広い | 攻撃面が小さい | - |

### 2. 非rootユーザーの重要性

```dockerfile
USER appuser
```

- コンテナ脱獄のリスクを軽減
- 本番環境では必須

### 3. 依存関係のキャッシュ最適化

```dockerfile
# 依存関係を先にコピー
COPY go.mod go.sum ./
RUN go mod download

# ソースコードは後
COPY . .
```

- ソースコードが変わっても、依存関係は再ダウンロードしない
- ビルド時間の短縮

### 4. 環境変数の管理

- `.env` ファイルで管理
- `.env` は `.gitignore` に追加
- `.env.example` をコミット

### 5. docker-composeの利便性

- 複数コンテナの管理が簡単
- 起動順序の制御
- ネットワークの自動構成

---

## トラブルシューティング

### イメージのビルドが失敗

```bash
# キャッシュを使わずにビルド
docker build --no-cache -t todo-app:latest .
```

### コンテナが起動しない

```bash
# ログを確認
docker compose logs app

# コンテナの状態を確認
docker compose ps
```

### データベースに接続できない

```bash
# データベースのヘルスチェックを確認
docker compose ps

# ネットワークを確認
docker network inspect day54_default
```

### ポートが既に使用されている

```
Error: bind: address already in use
```

**対策**:
- `docker-compose.yml` のポート番号を変更（例: 8080 → 8081）
- 既存のプロセスを停止

---

## 実務での活用

### 開発環境

```bash
# ホットリロード付きで起動（ソースコードの変更を自動反映）
docker compose watch
```

### 本番環境

```bash
# イメージをビルドしてレジストリにプッシュ
docker build -t registry.example.com/todo-app:v1.0.0 .
docker push registry.example.com/todo-app:v1.0.0

# 本番環境で起動
docker pull registry.example.com/todo-app:v1.0.0
docker run -d --env-file .env.production registry.example.com/todo-app:v1.0.0
```

---

## まとめ

### Docker化のメリット

1. **環境の一貫性**: 開発・ステージング・本番で同じ環境
2. **ポータビリティ**: どこでも動く
3. **リソース効率**: 軽量（VMより小さい）
4. **デプロイの簡素化**: イメージをpullして起動するだけ

### マルチステージビルドのメリット

1. **イメージサイズの削減**: 97%削減（1GB → 20-30MB）
2. **セキュリティ向上**: 攻撃面を減らす
3. **ビルド時間の短縮**: キャッシュの活用

### セキュリティのベストプラクティス

1. **非rootユーザーで実行**
2. **最小限のベースイメージ**（Alpine Linux）
3. **機密情報をイメージに含めない**
4. **定期的なイメージ更新**

---

**完了条件**:
- マルチステージビルドの仕組みを理解
- イメージサイズの削減効果を確認（97%削減）
- 非rootユーザーの重要性を理解
- 環境変数の管理方法を理解
- docker-composeでアプリケーション全体を起動できる

---

**次のステップ（Day 60）**: CI + API仕様（OpenAPI）+ Runbook（最終まとめ）
