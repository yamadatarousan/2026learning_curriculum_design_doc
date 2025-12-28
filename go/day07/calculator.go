package main

import (
"bufio"
"fmt"
"os"
"strconv"
)

func main() {

  var numbers []float64

  scanner := bufio.NewScanner(os.Stdin)
  fmt.Println("数字を入力してください（1行に1つ）。入力が終わったらCtrl+Dを押してください。")

  for scanner.Scan() {
    text := scanner.Text()
    num, err := strconv.ParseFloat(text, 64)
    if err != nil {
      fmt.Printf("\"%s\" は有効な数字ではありません。\n", text)
      continue
    }
    numbers = append(numbers, num)
  }

  var sum float64 = 0.0
  for _, v := range numbers {
    sum =+ v
  }

  average := sum / float64(len(numbers))

  fmt.Println("\n---")
  fmt.Printf("合計: %.2f\n", sum)
  fmt.Printf("平均: %.2f\n", average)
}
