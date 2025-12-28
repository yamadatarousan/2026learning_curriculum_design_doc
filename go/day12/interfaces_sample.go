package main

import (
  "fmt"
  "math"
)

type Shape interface {
  Area() float64
}

type Rectangle struct {
  Width float64
  Height float64
}

func (r Rectangle) Area() float64 {
  return r.Width * r.Height
}

type Circle struct {
  Radius float64
}

func (c Circle) Area() float64 {
  return math.Pi * c.Radius * c.Radius
}

func PrintArea(s Shape) {
  fmt.Printf("This shape's area is %f\n", s.Area())
}

func main() {
  rect := Rectangle{Width: 10, Height: 5}
  circ := Circle{Radius: 4}

  PrintArea(rect)
  PrintArea(circ)

  shapes := []Shape{rect, circ}
  for _, s := range shapes{
    fmt.Printf("Area from slice: %f\n", s.Area())
  }
}
