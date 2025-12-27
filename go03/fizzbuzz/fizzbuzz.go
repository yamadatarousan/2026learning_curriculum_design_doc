package main

import "fmt"

func main() {
    for i := 1; i <= 100; i++ {
        // ここにロジックを書く
        // 1. 15で割り切れるか？ (if i % 15 == 0)
        if i % 15 == 0 {
            fmt.Println("FizzBuzz")
        // 2. 3で割り切れるか？ (else if i % 3 == 0)
        } else if i % 3 == 0 {
            fmt.Println("Fizz")
        // 3. 5で割り切れるか？ (else if i % 5 == 0)
        } else if i % 5 == 0 {
            fmt.Println("Buzz")
        // 4. それ以外 (else)
        } else {
            fmt.Println(i) // この行をif-elseブロックの中に移動・変更していくことになります
        }
    }
}
