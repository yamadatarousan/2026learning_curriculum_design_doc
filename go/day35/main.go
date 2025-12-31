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

type Todo struct {
  ID   int    `json:"id"`
  Name string `json:"name" binding:"required"`
}

var db *sql.DB

func initDB() {
  var err error
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

func requestIDMiddleware() gin.HandlerFunc {  
  return func(c *gin.Context) {
    uuidObj, _ := uuid.NewRandom()
    requestID := uuidObj.String()

    c.Set("RequestID", requestID)

    c.Header("X-Request-ID", requestID)

    c.Next()
  }
}

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

func createTodo(c *gin.Context) {
  var newTodo Todo
  if err := c.BindJSON(&newTodo); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{
      "error": "Validation Failed",
      "details": err.Error(),
    })
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

  router := gin.New()

  logFormatter := func(param gin.LogFormatterParams) string {  
    requestID := param.Keys["RequestID"]

    return fmt.Sprintf("%s | %s | %s | %3d | %13v | %15s | %s\n",
      param.TimeStamp.Format(time.RFC3339), // timeパッケージの定数を明示的に使用
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

  router.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
  })
  router.GET("/todos", getTodos)
  router.POST("/todos", createTodo)

  log.Println("Starting server at port 8080")
  router.Run()
}

