package main

import (
  "errors"
  "fmt"
)

func divide(a, b float64) (float64, error) {
  if b == 0 {
    return 0, errors.New("ゼロで割ることはできません")
  }
  return a / b, nil
}

func main() {
  fmt.Println("--- 成功ケース ---")
  result, err := divide(10.0, 2.0)
  if err != nil {
    fmt.Println("エラー発生:", err)
  } else {
    fmt.Println("結果:", result)
  }
  
  fmt.Println("\n--- 失敗ケース ---")
  result, err = divide(10.0, 0.0)
  if err != nil {
    fmt.Println("エラー発生:", err)
  } else {
    fmt.Println("結果:", result)
  }
}
