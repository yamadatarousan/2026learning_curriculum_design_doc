package main

import (
"fmt"
"strings"
)

func main() {
  text :="apple banana apple cherry banana apple"

  words := strings.Split(text, " ")

  counts := make(map[string]int)

  for _, word := range words {
    counts[word]++
  }

  fmt.Println(counts)
}
