package main

import "testing"

func TestAdd(t *testing.T) {
  x, y := 2, 3
  expected := 5

  result := Add(x, y)

  if result != expected {
    t.Errorf("Add(%d, %d) = %d; want %d", x, y, result, expected)
  }
}
