package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

var testDB *sql.DB

// TestMainは、パッケージ内のテストが実行される前に一度だけ呼ばれる特別な関数です。
func TestMain(m *testing.M) {
	// --- セットアップ ---
	log.Println("Spinning up test database...")
	// --waitフラグでhealthcheckが通るまで待機
	cmd := exec.Command("docker-compose", "-f", "../../docker-compose.test.yml", "up", "-d", "--wait")
	if err := cmd.Run(); err != nil {
		log.Fatalf("Could not start test database: %v", err)
	}

	// deferでテスト終了時に必ずDBコンテナを破棄する
	defer func() {
		log.Println("Tearing down test database...")
		cmd := exec.Command("docker-compose", "-f", "../../docker-compose.test.yml", "down")
		if err := cmd.Run(); err != nil {
			log.Printf("Could not stop test database: %v", err)
		}
	}()

	// テスト用DBへの接続
	dsnForGo := "host=localhost user=user password=password dbname=todo_test_db port=5434 sslmode=disable"
	dsnForMigrate := "postgres://user:password@localhost:5434/todo_test_db?sslmode=disable"

	var err error
	// DBが完全に準備が整うまでリトライ
	for i := 0; i < 5; i++ {
		testDB, err = sql.Open("pgx", dsnForGo)
		if err == nil {
			if err = testDB.Ping(); err == nil {
				break
			}
		}
		log.Printf("Could not connect to test DB, retrying... (%v)", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("Could not connect to test database after retries: %v", err)
	}

	// マイグレーションの実行
	log.Println("Running migrations on test database...")
	migrateCmd := exec.Command("migrate", "-database", dsnForMigrate, "-path", "db/migrations", "up")
	// migrateコマンドはgoディレクトリで実行する必要があるため、実行ディレクトリを変更
	migrateCmd.Dir = ".."
	if output, err := migrateCmd.CombinedOutput(); err != nil {
		log.Fatalf("Could not run migrations: %v\nOutput: %s", err, string(output))
	}

	// シードデータのロード
	log.Println("Loading seed data...")
	if err := loadSeedData(testDB); err != nil {
		log.Fatalf("Could not load seed data: %v", err)
	}

	// --- テストの実行 ---
	code := m.Run()

	// --- 終了処理 ---
	os.Exit(code)
}

// setupTestRouterはテスト用のDB接続を受け取るように変更
func setupTestRouter(dbConn *sql.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)

	// mainのdbではなく、引数で渡されたテスト用DB接続を使う
	repo := NewTodoRepository(dbConn)
	todoHandler := NewTodoHandler(repo)
	authHandler := NewAuthHandler(repo)
	adminHandler := NewAdminHandler(repo)

	router := gin.New()
	router.Use(cors.Default())

	router.POST("/signup", errorHandler(authHandler.signup))
	router.POST("/login", errorHandler(authHandler.login))

	v1 := router.Group("/api/v1")
	v1.Use(authMiddleware())
	{
		v1.GET("/todos", errorHandler(todoHandler.getTodos))
		v1.POST("/todos", errorHandler(todoHandler.createTodo))

		adminRoutes := v1.Group("/admin")
		adminRoutes.Use(adminMiddleware())
		{
			adminRoutes.GET("/users", errorHandler(adminHandler.getAllUsers))
		}
	}
	return router
}

// TestUserFlowは、TestMainで準備されたテスト用DBを使って実行される
func TestUserFlow(t *testing.T) {
	// グローバルなtestDB接続を使ってルーターをセットアップ
	router := setupTestRouter(testDB)

	// --- 1. ログイン ---
	// TestMainでseed.sqlがロードされているため、既存のユーザーを使用
	loginBody := `{"email": "user-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]
	assert.NotEmpty(t, token)

	// --- 3. TODO作成 ---
	todoBody := `{"name": "Isolated Test Todo"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
}

// loadSeedDataはseed.sqlを読み込み、テストDBに適用します。
func loadSeedData(db *sql.DB) error {
	seedSQL, err := os.ReadFile("../../go/testdata/seed.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(seedSQL))
	if err != nil {
		return err
	}
	return nil
}