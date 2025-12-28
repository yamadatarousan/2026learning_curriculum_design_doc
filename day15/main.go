package main

import (
  "fmt"
  "math"
)

type Shape interface {
  Area() float64
  Perimeter() float64
}

type Rectangle struct {
  Width float64
  Height float64
}

func (r Rectangle) Area() float64 {
  return r.Width * r.Height
}

func (r Rectangle) Perimeter() float64 {
  return 2 * (r.Width + r.Height)
} 

type Circle struct {
  Radius float64
}

func (c Circle) Area() float64 {
  return math.Pi * c.Radius * c.Radius
}

func (c Circle) Perimeter() float64 {
  return 2 * math.Pi * c.Radius
}

func PrintArea(s Shape) {
  fmt.Printf("This shape's area is %f\n", s.Area())
}

func main() {
  rect1 := Rectangle{Width: 10, Height: 5}
  rect2 := Rectangle{Width: 20, Height: 5}
  circ1 := Circle{Radius: 4}
  circ2 := Circle{Radius: 8}

  PrintArea(rect1)
  PrintArea(rect2)
  PrintArea(circ1)
  PrintArea(circ2)

  shapes := []Shape{rect1, circ1, rect2, circ2}
  var totalArea float64
  for _, s := range shapes {
    fmt.Printf("Area from slice: %f\n", s.Area())
    fmt.Printf("Perimeter from slice: %f\n", s.Perimeter())
    totalArea += s.Area()
  }

  fmt.Printf("totalArea: %f\n", totalArea)
}
