package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// setupTestRouterはテスト用のルーターを初期化して返します。
// 実際のmain関数とほぼ同じですが、DBの初期化などをテストの都度行えるように分離しています。
func setupTestRouter() *gin.Engine {
	// テスト中はデバッグメッセージを抑制
	gin.SetMode(gin.TestMode)

	// main関数と同様に依存関係を構築
	initDB()
	repo := NewTodoRepository(db)
	todoHandler := NewTodoHandler(repo)
	authHandler := NewAuthHandler(repo)
	adminHandler := NewAdminHandler(repo)

	router := gin.New()
	router.Use(cors.Default()) // テスト用にデフォルトのCORS設定

	// ルーティングもmain関数と同様に設定
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

// TestUserFlowは、ユーザーの一連の操作をシミュレートする結合テストです。
func TestUserFlow(t *testing.T) {
	router := setupTestRouter()
	
	// --- 1. 新規登録 ---
	signupBody := `{"email": "integration-test@example.com", "password": "password123"}`
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBufferString(signupBody))
	req.Header.Set("Content-Type", "application/json")
	
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// --- 2. ログイン ---
	loginBody := `{"email": "integration-test@example.com", "password": "password123"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// レスポンスからトークンを抜き出す
	var loginResponse map[string]string
	json.Unmarshal(w.Body.Bytes(), &loginResponse)
	token := loginResponse["token"]
	assert.NotEmpty(t, token)

	// --- 3. 認証が必要なAPI（TODO作成）をトークン付きで叩く ---
	todoBody := `{"name": "Test Todo from Integration Test"}`
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/todos", bytes.NewBufferString(todoBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token) // 取得したトークンをセット

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// レスポンスから作成されたTODOのIDなどを確認
	var todoResponse Todo
	json.Unmarshal(w.Body.Bytes(), &todoResponse)
	assert.Equal(t, "Test Todo from Integration Test", todoResponse.Name)
	assert.NotZero(t, todoResponse.ID)

	// --- 4. TODOリストを取得し、今作成したTODOが含まれているか確認 ---
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/todos", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var todosResponse []Todo
	json.Unmarshal(w.Body.Bytes(), &todosResponse)
	assert.NotEmpty(t, todosResponse)
	// 作成したTODOがリストの最初の要素であると仮定してチェック
	assert.Equal(t, "Test Todo from Integration Test", todosResponse[0].Name)
}
