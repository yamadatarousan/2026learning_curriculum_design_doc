package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Todo はTODOアイテムの構造体です。
type Todo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var db *sql.DB

// initDB はデータベースを初期化し、テーブルを作成します。
func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./todos_day34.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	createTableSQL := `CREATE TABLE IF NOT EXISTS todos (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name" TEXT
  );`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

// requestIDMiddleware はリクエストにユニークなIDを付与するミドルウェアです。
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// UUIDを生成
		uuidObj, _ := uuid.NewRandom()
		requestID := uuidObj.String()

		// リクエストのコンテキストにIDを設定。これにより、後続のハンドラでIDを取得できる。
		c.Set("RequestID", requestID)

		// レスポンスヘッダーにIDを設定。クライアント側でもIDを確認できる。
		c.Header("X-Request-ID", requestID)

		// c.Next() を呼び出して、後続のミドルウェアまたはハンドラ処理を実行
		c.Next()
	}
}

// getTodos はすべてのTODOアイテムを取得してJSONで返します。
func getTodos(c *gin.Context) {
	rows, err := db.Query("SELECT id, name FROM todos")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		todos = append(todos, t)
	}

	c.JSON(http.StatusOK, todos)
}

// createTodo は新しいTODOアイテムを作成します。
func createTodo(c *gin.Context) {
	var newTodo Todo
	if err := c.BindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	result, err := db.Exec("INSERT INTO todos (name) VALUES (?)", newTodo.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	newTodo.ID = int(id)

	c.JSON(http.StatusCreated, newTodo)
}

func main() {
	initDB()

	// gin.Default() の代わりに gin.New() を使用して、ミドルウェアを自分で制御します。
	router := gin.New()

	// カスタムのログフォーマットを定義します。
	logFormatter := func(param gin.LogFormatterParams) string {
		// コンテキストから "RequestID" を取得します。
		requestID, _ := param.Context.Get("RequestID")

		// ログに RequestID を含めるようにフォーマットをカスタマイズします。
		return fmt.Sprintf("%s | %s | %s | %3d | %13v | %15s | %s\n",
			param.TimeStamp.Format("2006/01/02 - 15:04:05"),
			requestID,
			param.Method,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Path,
		)
	}

	// ミドルウェアを .Use() で適用します。適用した順に実行されます。
	// 1. Recovery: panicが発生してもサーバーが落ちないようにする。
	router.Use(gin.Recovery())
	// 2. RequestID: これ以降の処理（ロガーなど）で使えるようにIDを生成する。
	router.Use(requestIDMiddleware())
	// 3. Logger: カスタムフォーマットのロガーを適用する。
	router.Use(gin.LoggerWithFormatter(logFormatter))

	// ルーティングを設定
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/todos", getTodos)
	router.POST("/todos", createTodo)

	log.Println("Starting server at port 8080")
	router.Run()
}
