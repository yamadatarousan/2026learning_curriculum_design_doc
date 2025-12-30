package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ginエンジンのインスタンスを作成します。
	// Default() はロガーとリカバリーのミドルウェアを自動で有効にします。
	router := gin.Default()

	// ルーティングを定義します。
	// GETリクエストで /health にアクセスがあった際に、
	// JSONレスポンス {"status": "ok"} を返します。
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	// サーバーを :8080 ポートで起動します。
	router.Run() // デフォルトは router.Run(":8080") と同じです
}
