# Docker化ガイド

このドキュメントは、GoアプリケーションをDocker化する際のベストプラクティスと実行手順をまとめたものです。

---

## マルチステージビルドとは

### 従来の方法（シングルステージ）

```dockerfile
FROM golang:1.21
WORKDIR /app
COPY . .
RUN go build -o server
CMD ["/app/server"]
```

**問題点**:
- イメージサイズが大きい（約1GB）
- Go コンパイラや不要なツールが含まれる
- セキュリティリスク（攻撃面が広い）

### マルチステージビルド

```dockerfile
# Stage 1: ビルド
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server

# Stage 2: 実行
FROM alpine:latest
COPY --from=builder /app/server .
CMD ["/app/server"]
```

**メリット**:
- イメージサイズが小さい（約20MB）
- 実行に必要なものだけを含む
- セキュリティ向上

---

## Dockerfile の構成要素

### 1. ビルドステージ

```dockerfile
FROM golang:1.21-alpine AS builder
```

- `golang:1.21-alpine`: 軽量なAlpine Linuxベース
- `AS builder`: このステージに名前を付ける

```dockerfile
WORKDIR /app
```

- 作業ディレクトリを設定

```dockerfile
COPY go.mod go.sum ./
RUN go mod download
```

- 依存関係のファイルを先にコピー
- Dockerのレイヤーキャッシュを活用（ソースコードが変わっても依存関係は再ダウンロードしない）

```dockerfile
COPY . .
```

- ソースコード全体をコピー

```dockerfile
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .
```

- `CGO_ENABLED=0`: C言語ライブラリへの依存をなくす（静的リンク）
- `GOOS=linux`: Linuxバイナリをビルド
- `-ldflags="-s -w"`: デバッグ情報を削除してサイズ削減
  - `-s`: シンボルテーブルを削除
  - `-w`: DWARFデバッグ情報を削除

### 2. 実行ステージ

```dockerfile
FROM alpine:latest
```

- 軽量なベースイメージ（約5MB）

```dockerfile
RUN apk --no-cache add ca-certificates
```

- HTTPS通信に必要なCA証明書をインストール
- 外部APIと通信する場合に必須

```dockerfile
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser
```

- 非rootユーザーを作成（セキュリティのため）
- `-D`: パスワードなし
- `-u 1000`: ユーザーID
- `-G appgroup`: グループに所属

```dockerfile
COPY --from=builder /app/server .
```

- ビルドステージからバイナリをコピー

```dockerfile
USER appuser
```

- 非rootユーザーに切り替え
- rootで実行すると、コンテナ脱獄のリスクが高まる

```dockerfile
EXPOSE 8080
```

- アプリケーションが使用するポートを文書化
- 実際のポート公開は `docker run -p` で指定

```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1
```

- コンテナの健全性をチェック
- `--interval=30s`: 30秒ごとにチェック
- `--timeout=3s`: 3秒でタイムアウト
- `--retries=3`: 3回失敗したら異常とみなす

```dockerfile
CMD ["/app/server"]
```

- コンテナ起動時に実行するコマンド

---

## .dockerignore の重要性

`.dockerignore` は、Dockerビルド時にコピーしないファイルを指定します。

### なぜ必要か

- ビルド速度の向上（不要なファイルをコピーしない）
- イメージサイズの削減
- セキュリティ（機密ファイルを含めない）

### 基本的な内容

```
.git
*.md
*_test.go
.env
tmp/
```

---

## 環境変数の扱い

### 方法1: docker run で渡す

```bash
docker run -e JWT_SECRET=my-secret -e DB_HOST=localhost app
```

### 方法2: .env ファイルを使う

```bash
docker run --env-file .env app
```

### 方法3: docker-compose.yml で定義

```yaml
services:
  app:
    environment:
      JWT_SECRET: ${JWT_SECRET:-default-value}
      DB_HOST: db
```

### ベストプラクティス

1. **機密情報は .env ファイルに**
2. **.env は .gitignore に追加**
3. **.env.example をコミット**（値は空またはダミー）
4. **docker-compose.yml にデフォルト値を設定**

---

## イメージサイズの最適化

### 1. マルチステージビルド

- 1GB → 20MB（約50分の1）

### 2. Alpine Linux を使用

- `golang:1.21` (約1GB) → `golang:1.21-alpine` (約300MB)
- `ubuntu` (約70MB) → `alpine` (約5MB)

### 3. ビルドフラグ

```bash
go build -ldflags="-s -w"
```

- バイナリサイズが30-40%削減

### 4. 不要なファイルを含めない

- `.dockerignore` で除外

### サイズ比較例

| 方法 | イメージサイズ |
|------|---------------|
| シングルステージ (golang:1.21) | 約1GB |
| シングルステージ (golang:1.21-alpine) | 約300MB |
| マルチステージ (alpine) | 約20MB |
| マルチステージ + 最適化フラグ | 約15MB |

---

## セキュリティのベストプラクティス

### 1. 非rootユーザーで実行

```dockerfile
RUN adduser -D appuser
USER appuser
```

**理由**: rootで実行すると、脆弱性があった場合にホストOSへの影響が大きい

### 2. 最小限のベースイメージ

```dockerfile
FROM alpine:latest  # または scratch
```

**理由**: 攻撃面を減らす（不要なツールがない）

### 3. 定期的なイメージ更新

```bash
docker pull alpine:latest
docker build --no-cache -t app:latest .
```

**理由**: セキュリティパッチを適用

### 4. 機密情報をイメージに含めない

- `.env` ファイルを `.dockerignore` に追加
- 環境変数で渡す

### 5. マルチステージビルド

- ビルドツールを含めない
- ソースコードを含めない（バイナリのみ）

---

## トラブルシューティング

### イメージサイズが大きい

```bash
# レイヤーごとのサイズを確認
docker history app:latest
```

**対策**:
- マルチステージビルドを使う
- Alpine Linux を使う
- ビルドフラグで最適化

### バイナリが実行できない

```
standard_init_linux.go:xxx: exec user process caused: no such file or directory
```

**原因**: 動的リンクのバイナリをscratchで実行

**対策**:
```dockerfile
RUN CGO_ENABLED=0 go build  # 静的リンク
```

### CA証明書のエラー

```
x509: certificate signed by unknown authority
```

**対策**:
```dockerfile
RUN apk --no-cache add ca-certificates
```

### パーミッションエラー

```
permission denied
```

**対策**:
```dockerfile
RUN chown -R appuser:appgroup /app
USER appuser
```

---

## 実務での運用

### 開発環境

```bash
# ホットリロード付きで起動
docker-compose -f docker-compose.dev.yml up
```

**docker-compose.dev.yml**:
```yaml
services:
  app:
    build:
      target: builder  # ビルドステージで止める
    command: go run main.go
    volumes:
      - .:/app  # ソースコードをマウント
```

### ステージング/本番環境

```bash
# イメージをビルド
docker build -t app:v1.0.0 .

# タグ付け
docker tag app:v1.0.0 registry.example.com/app:v1.0.0

# レジストリにプッシュ
docker push registry.example.com/app:v1.0.0

# 本番環境で起動
docker pull registry.example.com/app:v1.0.0
docker run -d --env-file .env.production registry.example.com/app:v1.0.0
```

---

## CI/CDでのビルド

### GitHub Actions 例

```yaml
name: Build and Push Docker Image

on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build Docker image
        run: docker build -t app:latest .

      - name: Run tests in container
        run: docker run app:latest go test ./...

      - name: Push to registry
        run: |
          echo ${{ secrets.REGISTRY_PASSWORD }} | docker login -u ${{ secrets.REGISTRY_USER }} --password-stdin
          docker tag app:latest registry.example.com/app:latest
          docker push registry.example.com/app:latest
```

---

## 参考資料

- [Docker公式ドキュメント: Best practices](https://docs.docker.com/develop/develop-images/dockerfile_best-practices/)
- [Docker公式ドキュメント: Multi-stage builds](https://docs.docker.com/build/building/multi-stage/)
- [Go公式: Building minimal Docker containers](https://golang.org/doc/articles/wiki/)

---

**最終更新**: 2026-01-04
