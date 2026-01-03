package main

import (
	"context"
	"database/sql"
	"fmt"
)

type TodoRepository struct {
	db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

func (r *TodoRepository) FindAll(userID int) ([]Todo, error) {
	rows, err := r.db.Query("SELECT id, name, user_id FROM todos WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Name, &t.UserID); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

func (r *TodoRepository) execTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit()
}

func (r *TodoRepository) createTodoInTx(tx *sql.Tx, todo Todo) (Todo, error) {
	var id int
	err := tx.QueryRow("INSERT INTO todos (name, user_id) VALUES ($1, $2) RETURNING id", todo.Name, todo.UserID).Scan(&id)
	if err != nil {
		return todo, err
	}
	todo.ID = id

	_, err = tx.Exec("INSERT INTO todo_audit_logs (todo_id, operation) VALUES ($1, $2)", id, "create")
	if err != nil {
		return todo, err
	}

	return todo, nil
}

func (r *TodoRepository) CreateTodoWithAudit(ctx context.Context, todo Todo) (Todo, error) {
	var createdTodo Todo
	err := r.execTx(ctx, func(tx *sql.Tx) error {
		var err error
		createdTodo, err = r.createTodoInTx(tx, todo)
		return err
	})

	return createdTodo, err
}

// --- User methods ---

func (r *TodoRepository) CreateUser(user User) (User, error) {
	// roleはDBのデフォルト値('user')が使われる
	err := r.db.QueryRow("INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, created_at, role", user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt, &user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r *TodoRepository) FindUserByEmail(email string) (User, error) {
	var user User
	// roleも取得するように変更
	err := r.db.QueryRow("SELECT id, email, password_hash, created_at, role FROM users WHERE email = $1", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt, &user.Role)
	if err != nil {
		return user, err
	}
	return user, nil
}

// FindAllUsersは管理者用に全てのユーザーを取得します。
func (r *TodoRepository) FindAllUsers() ([]User, error) {
	rows, err := r.db.Query("SELECT id, email, created_at, role FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Email, &u.CreatedAt, &u.Role); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}
