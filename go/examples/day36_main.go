package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Todo はTODOアイテムの構造体です。
type Todo struct {
	ID   int    `json:"id"`
	Name string `json:"name" binding:"required"`
}

var db *sql.DB

// AppHandler は、errorを返すことができるカスタムハンドラ型です。
// これにより、各ハンドラはビジネスロジックとエラー返却に専念できます。
type AppHandler func(c *gin.Context) error

// errorHandler は、AppHandlerを受け取り、gin.HandlerFuncに変換するアダプターです。
// この中で、AppHandlerから返されたエラーを一元的に処理し、HTTPレスポンスを生成します。
func errorHandler(handler AppHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		// カスタムハンドラを実行
		if err := handler(c); err != nil {
			log.Printf("Error occurred: %v", err) // エラーをログに出力

			// エラーの種類を判別
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				// バリデーションエラーの場合
				c.JSON(http.StatusBadRequest, gin.H{
					"error":   "Validation Failed",
					"details": err.Error(),
				})
				return
			}

			// その他のエラーはすべて「Internal Server Error」として処理
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error",
			})
		}
	}
}

func initDB() {
	var err error
	db, err = sql.Open("sqlite3", "./todos_day36.db")
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	createTableSQL := `CREATE TABLE IF NOT EXISTS todos (
    "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT, "name" TEXT);`
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

// getTodos は AppHandler 型の関数として定義します。
func getTodos(c *gin.Context) error {
	rows, err := db.Query("SELECT id, name FROM todos")
	if err != nil {
		return err // エラーが発生したら、errをそのまま返す
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return err // エラーが発生したら、errをそのまま返す
		}
		todos = append(todos, t)
	}

	c.JSON(http.StatusOK, todos)
	return nil // 成功した場合は nil を返す
}

// createTodo も AppHandler 型の関数として定義します。
func createTodo(c *gin.Context) error {
	var newTodo Todo
	if err := c.BindJSON(&newTodo); err != nil {
		return err // バリデーションエラーなどをそのまま返す
	}

	result, err := db.Exec("INSERT INTO todos (name) VALUES (?)", newTodo.Name)
	if err != nil {
		return err // エラーが発生したら、errをそのまま返す
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err // エラーが発生したら、errをそのまま返す
	}
	newTodo.ID = int(id)

	c.JSON(http.StatusCreated, newTodo)
	return nil // 成功した場合は nil を返す
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

	// 各ルートで、ハンドラを errorHandler アダプターでラップします。
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/todos", errorHandler(getTodos))
	router.POST("/todos", errorHandler(createTodo))

	log.Println("Starting server at port 8080")
	router.Run()
}
