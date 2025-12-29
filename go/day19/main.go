package main

import (
  "encoding/json"
  "fmt"
  "log"
  "net/http"
)

type User struct {
  ID    int     `json:"id"`
  Name  string  `json:"name"`
}

func main() {
  mux := http.NewServeMux()

  mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    fmt.Fprint(w, "Welcome to our API server!")
  })

  mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {

    users := []User{
      {ID:1, Name: "Taro Yamada"},
      {ID:2, Name: "Hanako Sato"},
      {ID:3, Name: "Jiro Suzuki"},
    }

    w.Header().Set("Content-Type", "application/json")

    jsonBytes, err := json.Marshal(users)
    if err != nil {
      log.Printf("Error marshalling users: %v", err)
      http.Error(w, "Internal Server Error", http.StatusInternalServerError)
      return
    }
    
    w.Write(jsonBytes)
  })

  fmt.Println("Starting server at port 8080")
  fmt.Println("Access http://localhost:8080/ or http://localhost:8080/users")

  err := http.ListenAndServe(":8080", mux)
  if err != nil {
    log.Fatal(err)
  }
}
