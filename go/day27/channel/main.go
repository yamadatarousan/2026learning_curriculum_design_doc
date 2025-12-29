package main

import (
  "fmt"
  "time"
)

func worker(ch chan string) {
  fmt.Println("...worker is doing something...")
  time.Sleep(2 * time.Second)
  fmt.Println("...worker finished!")

  ch <- "Job finished"
}

func main() {
  ch := make(chan string)

  go worker(ch)

  fmt.Println("Waiting for worker...")

  result := <- ch

  fmt.Println("Result from worker:", result)
  fmt.Println("Main function finished.")
}
