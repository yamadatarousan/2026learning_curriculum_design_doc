package main

import (
  "database/sql"
  "log"
  "net/http"

  "github.com/gin-gonic/gin"
  _ "github.com/mattn/go-sqlite3"
)

type Todo struct {
  ID   int    `json:"id"`
  Name string `json:"name"`
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

  router := gin.Default()

  router.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
  })
  router.GET("/todos", getTodos)
  router.POST("/todos", createTodo)

  log.Println("Starting server at port 8080")
  router.Run()
}

