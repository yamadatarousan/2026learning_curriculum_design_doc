package main

import (
  "encoding/json"
  "net/http"
  "net/http/httptest"
  "reflect"
  "testing"
)

func setupTestDB() {
  initDB()

  db.Exec("DROP TABLE IF EXISTS todos")
  
  initDB()

  db.Exec(`INSERT INTO todos (id, name) VALUES (1, "Test Todo 1")`)
}

func TestGetTodosHandler(t *testing.T) {
  setupTestDB()

  req := httptest.NewRequest("GET", "/todos", nil)
  rr := httptest.NewRecorder()

  getTodosHandler(rr, req)

  if status := rr.Code; status != http.StatusOK {
    t.Errorf("handler returned wrong status code: got %v want %v",
      status, http.StatusOK)
  }

  expected := []Todo{{ID: 1, Name: "Test Todo 1"}}

  var actual []Todo
  if err := json.NewDecoder(rr.Body).Decode(&actual); err != nil {
    t.Fatalf("Could not decode response body: %v", err)
  }

  if !reflect.DeepEqual(actual, expected) {
    t.Errorf("handler returned unexpected body: got %v want %v",
      actual, expected)
  }
}
