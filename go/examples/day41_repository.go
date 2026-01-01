package main

import (
	"context"
	"database/sql"
	"fmt"
)

// TodoRepositoryはデータベース操作を担当します。
type TodoRepository struct {
	db *sql.DB
}

// NewTodoRepositoryは新しいTodoRepositoryのインスタンスを作成します。
func NewTodoRepository(db *sql.DB) *TodoRepository {
	return &TodoRepository{db: db}
}

// FindAllはすべてのTODOアイテムを取得します。
func (r *TodoRepository) FindAll() ([]Todo, error) {
	rows, err := r.db.Query("SELECT id, name FROM todos")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		todos = append(todos, t)
	}
	return todos, nil
}

// execTxはトランザクションを実行するためのヘルパー関数です。
// トランザクションを開始し、渡された関数(fn)を実行します。
// fnがエラーを返した場合、トランザクションはロールバックされます。
// エラーがなければ、トランザクションはコミットされます。
func (r *TodoRepository) execTx(ctx context.Context, fn func(*sql.Tx) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		// エラーが発生した場合、ロールバックを試みる
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	// エラーがなければコミット
	return tx.Commit()
}

// createTodoInTxはトランザクション内でTODOと監査ログを作成します。
func (r *TodoRepository) createTodoInTx(tx *sql.Tx, todo Todo) (Todo, error) {
	// 1. todosテーブルに新しいTODOを挿入し、IDを取得
	var id int
	err := tx.QueryRow("INSERT INTO todos (name) VALUES ($1) RETURNING id", todo.Name).Scan(&id)
	if err != nil {
		return todo, err
	}
	todo.ID = id

	// 2. todo_audit_logsテーブルに監査ログを挿入
	_, err = tx.Exec("INSERT INTO todo_audit_logs (todo_id, operation) VALUES ($1, $2)", id, "create")
	if err != nil {
		return todo, err
	}

	return todo, nil
}

// CreateTodoWithAuditはトランザクションを使用してTODOと監査ログを作成します。
func (r *TodoRepository) CreateTodoWithAudit(ctx context.Context, todo Todo) (Todo, error) {
	var createdTodo Todo
	err := r.execTx(ctx, func(tx *sql.Tx) error {
		var err error
		createdTodo, err = r.createTodoInTx(tx, todo)
		return err
	})

	return createdTodo, err
}
