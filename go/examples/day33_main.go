package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
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
	// Day30では "todos.db" でしたが、Day33用にファイルを分けるために "todos_day33.db" とします。
	db, err = sql.Open("sqlite3", "./todos_day33.db")
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
	// c.BindJSON を使ってリクエストボディを構造体にバインドします。
	if err := c.BindJSON(&newTodo); err != nil {
		// バインドに失敗した場合は 400 Bad Request を返します。
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
	// データベースを初期化
	initDB()

	// Ginルーターを初期化
	router := gin.Default()

	// ルーティングを設定
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/todos", getTodos)
	router.POST("/todos", createTodo)

	// サーバーを :8080 ポートで起動
	log.Println("Starting server at port 8080")
	router.Run()
}
