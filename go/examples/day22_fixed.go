package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Todo struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var todos = []Todo{
	{ID: 1, Name: "Taro Yamada todo"},
	{ID: 2, Name: "Hanako Sato todo"},
}
var nextID = 3

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Welcome to our API server!")
	})

	mux.HandleFunc("/todos", func(w http.ResponseWriter, r *http.Request) {

		switch r.Method {
		case http.MethodGet:
			w.Header().Set("Content-Type", "application/json")

			jsonBytes, err := json.Marshal(todos)
			if err != nil {
				log.Printf("Error marshalling users: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			w.Write(jsonBytes)
		case http.MethodPost:
			var newTodo Todo
			if err := json.NewDecoder(r.Body).Decode(&newTodo); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			newTodo.ID = nextID
			nextID++
			todos = append(todos, newTodo)

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(newTodo)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}

	})

	fmt.Println("Starting server at port 8080")
	fmt.Println("Access http://localhost:8080/ or http://localhost:8080/todos")

	err := http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal(err)
	}
}
