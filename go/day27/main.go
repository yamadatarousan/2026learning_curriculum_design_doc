package main

import (
  "fmt"
  "math/rand"
  "time"
)

func fetchFromAPI(apiName string, ch chan string) {
  sleepTime := time.Duration(100 + rand.Intn(900)) * time.Millisecond
  time.Sleep(sleepTime)

  result := fmt.Sprintf("Result from %s (took %v)", apiName, sleepTime)
  ch <- result
}

func main() {
  rand.Seed(time.Now().UnixNano())

  ch := make(chan string)

  apis := []string{"Google API", "Facebook API", "Twitter API", "GitHub API"}

  for _, api := range apis {
    fmt.Printf("Fetching from %s...\n", api)
    go fetchFromAPI(api, ch)
  }

  fmt.Println("\nWaiting for results...")
  
  for i := 0; i < len(apis); i++ {
    result := <-ch
    fmt.Println(result)
  }

  fmt.Println("\nAll APIs have responded.")
}
