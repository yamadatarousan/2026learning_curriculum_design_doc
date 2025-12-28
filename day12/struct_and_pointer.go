package main

import "fmt"

type User struct {
	ID	   int
	Name   string
	IsActive bool
  Age   int
}

func deactivate(user *User) {
	fmt.Printf("\n関数内でディアクティベートします: %s\n", user.Name)
	user.IsActive = false
}

func (u *User) celebrateBirthday_Pointer() {
  u.Age++
  fmt.Printf("  [In Pointer Method] %s is now %d\n", u.Name, u.Age)
}

func main() {
	user1 := User{
		ID:       1,
		Name:     "Yamada",
		IsActive: true,
    Age:  20,
	}
	fmt.Println("変更前のユーザー状態:", user1)

	deactivate(&user1)

	fmt.Println("変更後のユーザー状態:", user1)

  // ポインタレシーバーの場合
  fmt.Printf("Before: %s is %d\n", user1.Name, user1.Age)
  user1.celebrateBirthday_Pointer()
  fmt.Printf("After: %s is %d\n", user1.Name, user1.Age) // 年齢が変わる！
}
