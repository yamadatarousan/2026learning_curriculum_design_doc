# マルチステージビルド: Stage 1 - ビルド環境
# バージョン指定なし = 最新の安定版を使用
FROM golang:alpine AS builder
RUN apk add --no-cache git
# GOTOOLCHAIN=auto で go.mod が要求するバージョンを自動使用
ENV GOTOOLCHAIN=auto
WORKDIR /app
# 親ディレクトリのgo.modとgo.sumをコピー
COPY go.mod go.sum ./
RUN go mod download
# day54のソースコードをコピー
COPY day54/ ./day54/
# day54ディレクトリでビルド
WORKDIR /app/day54
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/server .
# Stage 2: 実行環境
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN addgroup -g 1000 appgroup && \
    adduser -D -u 1000 -G appgroup appuser
WORKDIR /app
COPY --from=builder /app/server .
RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --quiet --tries=1 --spider http://localhost:8080/health || exit 1
CMD ["/app/server"]