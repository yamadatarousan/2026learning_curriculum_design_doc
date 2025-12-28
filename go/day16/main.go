package main

import (
  "fmt"
  "log"
  "net/http"
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintf(w, "Hello, Web!")
}

func main() {
  http.HandleFunc("/hello", helloHandler)

  fmt.Println("Starting server at port 8080")
  fmt.Println("Please access http://localhost:8080/hello")

  err := http.ListenAndServe(":8080", nil)
  if err != nil {
    log.Fatal("ListenAndServe: ", err)
  }
}
