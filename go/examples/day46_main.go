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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"golang.org/x/crypto/bcrypt"
)

// --- Models ---

type Todo struct {
	ID   int    `json:"id"`
	Name string `json:"name" binding:"required"`
}

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never return password hash
	CreatedAt    time.Time `json:"created_at"`
}

// --- Global Variables ---

var db *sql.DB
// NOTE: In a real application, this secret should be loaded from a secure configuration, not hardcoded.
var jwtSecret = []byte("a-very-secret-key")

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
				if pgErr.Code == "23505" { // unique_violation
					c.JSON(http.StatusConflict, gin.H{"error": "Conflict", "message": "This resource already exists."})
					return
				}
			}
			
			// Specific error for login failure
			if errors.Is(err, sql.ErrNoRows) || errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized", "message": "Invalid email or password"})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
		}
	}
}

// --- TODO Handler ---

type TodoHandler struct {
	repo *TodoRepository
}

func NewTodoHandler(repo *TodoRepository) *TodoHandler {
	return &TodoHandler{repo: repo}
}

func (h *TodoHandler) getTodos(c *gin.Context) error {
	todos, err := h.repo.FindAll()
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

	createdTodo, err := h.repo.CreateTodoWithAudit(c.Request.Context(), newTodo)
	if err != nil {
		return err
	}

	c.JSON(http.StatusCreated, createdTodo)
	return nil
}

// --- Auth Handler ---

type AuthHandler struct {
	repo *TodoRepository
}

func NewAuthHandler(repo *TodoRepository) *AuthHandler {
	return &AuthHandler{repo: repo}
}

type SignupInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

func (h *AuthHandler) signup(c *gin.Context) error {
	var input SignupInput
	if err := c.ShouldBindJSON(&input); err != nil {
		return err // Will be handled by errorHandler as a validation error
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user := User{
		Email:        input.Email,
		PasswordHash: string(hashedPassword),
	}

	createdUser, err := h.repo.CreateUser(user)
	if err != nil {
		return err // Will be handled by errorHandler (e.g., unique constraint)
	}

	c.JSON(http.StatusCreated, gin.H{"id": createdUser.ID, "email": createdUser.Email, "created_at": createdUser.CreatedAt})
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
		// Return a generic error to avoid telling attackers whether the email exists.
		// The errorHandler will catch sql.ErrNoRows and return 401.
		return err
	}

	// Compare the provided password with the stored hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password))
	if err != nil {
		// Password does not match. The errorHandler will catch this and return 401.
		return err
	}

	// Generate JWT
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   fmt.Sprint(user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)), // Token expires in 24 hours
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	tokenString, err := claims.SignedString(jwtSecret)
	if err != nil {
		return fmt.Errorf("failed to create token: %w", err)
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
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

	todoRepo := NewTodoRepository(db)
	todoHandler := NewTodoHandler(todoRepo)
	authHandler := NewAuthHandler(todoRepo) // Create AuthHandler

	router := gin.New()
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

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes
	router.POST("/signup", errorHandler(authHandler.signup))
	router.POST("/login", errorHandler(authHandler.login))

	// Todo routes (for now, they are public)
	router.GET("/todos", errorHandler(todoHandler.getTodos))
	router.POST("/todos", errorHandler(todoHandler.createTodo))

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