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

    // 【変更点①】ループの中でまとめて処理するので、個別の呼び出しは不要になります
    // PrintArea(rect1)
    // ...

    shapes := []Shape{rect1, circ1, rect2, circ2}

    // 【変更点②】合計面積を保存する変数をループの前に用意
    var totalArea float64

    fmt.Println("--- Shape Calculation Results ---")

    // ループ処理
    for _, s := range shapes {
        // 【変更点③】面積と周囲長の両方を表示
        // fmt.Printf("%T\n", s) を使うと具体的な型（RectangleかCircleか）も表示できます
        fmt.Printf("Shape type: %T, Area: %f, Perimeter: %f\n", s, s.Area(), s.Perimeter())

        // 【変更点④】合計面積に現在の図形の面積を加算
        totalArea += s.Area()
    }

    fmt.Println("---------------------------------")

    // 【変更点⑤】ループの後に、計算した合計面積を表示
    fmt.Printf("Total area of all shapes: %f\n", totalArea)
  }