# マルチステージビルド: Stage 1 - ビルド環境
FROM golang:1.21-alpine AS builder

# ビルドに必要なツールをインストール
RUN apk add --no-cache git

# 作業ディレクトリを設定
WORKDIR /app

# 依存関係のファイルを先にコピー（キャッシュ効率化）
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# バイナリをビルド
# -ldflags="-s -w" でバイナリサイズを削減（デバッグ情報を削除）
# CGO_ENABLED=0 で静的リンク（依存ライブラリを含めない）
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .

# マルチステージビルド: Stage 2 - 実行環境
FROM alpine:latest

# セキュリティアップデートとCA証明書のインストール
# （HTTPS通信を行う場合に必要）
RUN apk --no-cache add ca-certificates

# 非rootユーザーを作成（セキュリティのため）
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser

# 作業ディレクトリを設定
WORKDIR /app

# ビルドステージからバイナリをコピー
COPY --from=builder /app/server .

# （オプション）マイグレーションファイルなどもコピー
# COPY --from=builder /app/db/migrations ./db/migrations

# ファイルの所有者を非rootユーザーに変更
RUN chown -R appuser:appgroup /app

# 非rootユーザーに切り替え
USER appuser

# アプリケーションが使用するポートを公開
EXPOSE 8080

# ヘルスチェック（オプション）
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1

# アプリケーションを実行
CMD ["/app/server"]
