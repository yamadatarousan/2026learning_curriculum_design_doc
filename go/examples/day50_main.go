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
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors" // CORSミドルウェアをインポート
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

// --- Models ---

type Todo struct {
	ID     int    `json:"id"`
	Name   string `json:"name" binding:"required"`
	UserID int    `json:"user_id"`
}

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

// --- Global Variables & JWT Claims ---

var db *sql.DB
var jwtSecret = []byte("a-very-secret-key")

// AppClaimsはJWTに含めるカスタムクレームです
type AppClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// --- Middleware ---

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is malformed"})
			return
		}
		tokenString := parts[1]

		token, err := jwt.ParseWithClaims(tokenString, &AppClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token", "details": err.Error()})
			return
		}

		if claims, ok := token.Claims.(*AppClaims); ok && token.Valid {
			c.Set("claims", claims)
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
		}
	}
}

func adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Not an admin"})
			return
		}

		appClaims, ok := claims.(*AppClaims)
		if !ok || appClaims.Role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden", "message": "Not an admin"})
			return
		}

		c.Next()
	}
}

// --- Error Handling ---
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
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				if pgErr.Code == "23505" {
					c.JSON(http.StatusConflict, gin.H{"error": "Conflict", "message": "This resource already exists."})
					return
				}
			}
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "message": "Invalid email or password"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
	}
}

// --- Handlers ---

type TodoHandler struct{ repo *TodoRepository }

func NewTodoHandler(repo *TodoRepository) *TodoHandler { return &TodoHandler{repo: repo} }

func (h *TodoHandler) getTodos(c *gin.Context) error {
	claims := c.MustGet("claims").(*AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)
	todos, err := h.repo.FindAll(userID)
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, todos)
	return nil
}

func (h *TodoHandler) createTodo(c *gin.Context) error {
	var newTodo Todo
	if err := c.BindJSON(&newTodo); err != nil {
		return err
	}
	claims := c.MustGet("claims").(*AppClaims)
	userID, _ := strconv.Atoi(claims.Subject)
	newTodo.UserID = userID
	createdTodo, err := h.repo.CreateTodoWithAudit(c.Request.Context(), newTodo)
	if err != nil {
		return err
	}
	c.JSON(http.StatusCreated, createdTodo)
	return nil
}

type AuthHandler struct{ repo *TodoRepository }

func NewAuthHandler(repo *TodoRepository) *AuthHandler { return &AuthHandler{repo: repo} }

type SignupInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) signup(c *gin.Context) error {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}
	user := User{Email: input.Email, PasswordHash: string(hashedPassword)}
	createdUser, err := h.repo.CreateUser(user)
	if err != nil {
		return err
	}
	c.JSON(http.StatusCreated, gin.H{"id": createdUser.ID, "email": createdUser.Email, "created_at": createdUser.CreatedAt, "role": createdUser.Role})
	return nil
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) login(c *gin.Context) error {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err
	}
	user, err := h.repo.FindUserByEmail(input.Email)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		return err
	}
	claims := AppClaims{
		user.Role,
		jwt.RegisteredClaims{
			Subject:   fmt.Sprint(user.ID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}
	c.JSON(http.StatusOK, gin.H{"token": tokenString})
	return nil
}

type AdminHandler struct{ repo *TodoRepository }

func NewAdminHandler(repo *TodoRepository) *AdminHandler { return &AdminHandler{repo: repo} }

func (h *AdminHandler) getAllUsers(c *gin.Context) error {
	users, err := h.repo.FindAllUsers()
	if err != nil {
		return err
	}
	c.JSON(http.StatusOK, users)
	return nil
}

// --- Main Application ---

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

	repo := NewTodoRepository(db)
	todoHandler := NewTodoHandler(repo)
	authHandler := NewAuthHandler(repo)
	adminHandler := NewAdminHandler(repo)

	router := gin.New()

	// --- CORS Middleware ---
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:3000"} // 例: React開発サーバー
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	logFormatter := func(param gin.LogFormatterParams) string {
		requestID := param.Keys["RequestID"]
		return fmt.Sprintf("%s | %s | %s | %3d | %13v | %15s | %s\n",
			param.TimeStamp.Format(time.RFC3339),
			requestID,
			param.Method,
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Path,
		)
	}
	router.Use(gin.Recovery(), requestIDMiddleware(), gin.LoggerWithFormatter(logFormatter))

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
