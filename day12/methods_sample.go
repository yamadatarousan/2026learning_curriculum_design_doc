package main

import "fmt"

type User struct {
  Name string
  Age int
}

func (u User) celebrateBirthday_Value() {
  u.Age++
  fmt.Printf("  [In Value Method] %s is now %d\n", u.Name, u.Age)
}

func (u *User) celebrateBirthday_Pointer() {
  u.Age++
  fmt.Printf("  [In Pointer Method] %s is now %d\n", u.Name, u.Age)
}

func main() {
  user_v := User{Name: "Hanako", Age: 25}
  fmt.Printf("Before: %s is %d\n", user_v.Name, user_v.Age)
  user_v.celebrateBirthday_Value()
  fmt.Printf("After: %s is %d\n\n", user_v.Name, user_v.Age)

  user_p := User{Name: "Jiro", Age: 30}
  fmt.Printf("Before: %s is %d\n", user_p.Name, user_p.Age)
  user_p.celebrateBirthday_Pointer()
  fmt.Printf("After: %s is %d\n", user_p.Name, user_p.Age)
}
