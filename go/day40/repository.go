package main

import (
  "database/sql"
)

type TodoRepository struct {
  db *sql.DB
}

func NewTodoRepository(db *sql.DB) *TodoRepository {
  return &TodoRepository{db: db}
}

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

func (r *TodoRepository) Create(todo Todo) (Todo, error) {
  var id int
  err := r.db.QueryRow("INSERT INTO todos (name) VALUES ($1) RETURNING id", todo.Name).Scan(&id)
  if err != nil {
    return todo, err
  }
  todo.ID = id
  return todo, nil
}
