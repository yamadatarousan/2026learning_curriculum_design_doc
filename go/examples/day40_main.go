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
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Todo構造体は、複数のファイルで使われるため、main.goに残します。
// （より大きなアプリケーションでは、`model`や`domain`といった共通パッケージに置かれます）
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

// --- ハンドラ層の定義 ---
// TodoHandler は、リポジトリを依存関係として持ちます。
type TodoHandler struct {
	repo *TodoRepository
}

// NewTodoHandler は新しいTodoHandlerのインスタンスを作成します。
func NewTodoHandler(repo *TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

// getTodos はHTTPリクエストを処理し、リポジトリを呼び出します。
func (h *TodoHandler) getTodos(c *gin.Context) error {
	todos, err := h.repo.FindAll()
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, todos)
	return nil
}

// createTodo はHTTPリクエストを処理し、リポジトリを呼び出します。
func (h *TodoHandler) createTodo(c *gin.Context) error {
	var newTodo Todo
	if err := c.BindJSON(&newTodo); err != nil {
		return err
	}

	createdTodo, err := h.repo.Create(newTodo)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, createdTodo)
	return nil
}

// --- DB初期化、ミドルウェア (変更なし) ---
func initDB() {
	var err error
	dsn := "host=localhost user=user password=password dbname=todo_db port=5433 sslmode=disable"
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL!")
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

func main() {
	initDB()

	// --- 依存関係の構築 (DI: Dependency Injection) ---
	// 1. リポジトリのインスタンスを作成
	todoRepo := NewTodoRepository(db)
	// 2. ハンドラのインスタンスを作成し、リポジトリを注入
	todoHandler := NewTodoHandler(todoRepo)

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

	// --- ルーティング ---
	// ハンドラのメソッドを登録
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
	router.GET("/todos", errorHandler(todoHandler.getTodos))
	router.POST("/todos", errorHandler(todoHandler.createTodo))

	// --- サーバー起動 (変更なし) ---
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
