package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	// PostgreSQL用のドライバをインポート
	_ "github.com/jackc/pgx/v5/stdlib"
)

type Todo struct {
	ID   int    `json:"id"`
	Name string `json:"name" binding:"required"`
}

var db *sql.DB

type AppHandler func(c *gin.Context) error

func errorHandler(handler AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := handler(c); err != nil {
			log.Printf("Error occurred: %v", err)
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Validation Failed", "details": err.Error()})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
	}
}

// initDB はPostgreSQLデータベースに接続し、テーブルを作成します。
func initDB() {
	var err error
	// --- PostgreSQLへの接続情報 (DSN: Data Source Name) ---
	// docker-compose.ymlで設定した値に合わせて接続文字列を作成します。
	dsn := "host=localhost user=user password=password dbname=todo_db port=5432 sslmode=disable"

	// ドライバ名 "pgx" を使用してDBに接続します。
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	// 接続確認
	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Successfully connected to PostgreSQL!")

	// --- テーブル作成SQLをPostgreSQL用に変更 ---
	// AUTOINCREMENTの代わりに SERIAL PRIMARY KEY を使用します。
	createTableSQL := `CREATE TABLE IF NOT EXISTS todos (
    "id" SERIAL PRIMARY KEY,
    "name" TEXT
  );`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %v", err)
	}
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		uuidObj, _ := uuid.NewRandom()
		requestID := uuidObj.String()
		c.Set("RequestID", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

func getTodos(c *gin.Context) error {
	rows, err := db.Query("SELECT id, name FROM todos")
	if err != nil {
		return err
	}
	defer rows.Close()
	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return err
		}
		todos = append(todos, t)
	}
	c.JSON(http.StatusOK, todos)
	return nil
}

// createTodo のID取得方法をPostgreSQL用に変更
func createTodo(c *gin.Context) error {
	var newTodo Todo
	if err := c.BindJSON(&newTodo); err != nil {
		return err
	}

	// --- ID取得方法の変更 ---
	// LastInsertId()が使えないため、INSERT文に "RETURNING id" を追加し、
	// QueryRow().Scan() を使って新しく生成されたIDを取得します。
	var id int
	err := db.QueryRow("INSERT INTO todos (name) VALUES ($1) RETURNING id", newTodo.Name).Scan(&id)
	if err != nil {
		return err
	}
	newTodo.ID = id

	c.JSON(http.StatusCreated, newTodo)
	return nil
}

func main() {
	initDB()
	router := gin.New()
	logFormatter := func(param gin.LogFormatterParams) string {
		requestID := param.Keys["RequestID"]
		return fmt.Sprintf("%s | %s | %s | %3d | %13v | %15s | %s\n",
			param.TimeStamp.Format(time.RFC3339), requestID, param.Method,
			param.StatusCode, param.Latency, param.ClientIP, param.Path)
	}
	router.Use(gin.Recovery())
	router.Use(requestIDMiddleware())
	router.Use(gin.LoggerWithFormatter(logFormatter))
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/todos", errorHandler(getTodos))
	router.POST("/todos", errorHandler(createTodo))

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Starting server at port 8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	log.Println("Server exiting")
}
