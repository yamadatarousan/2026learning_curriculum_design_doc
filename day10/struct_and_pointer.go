package main

import "fmt"

type User struct {
	ID	   int
	Name   string
	IsActive bool
}

func deactivate(user *User) {
	fmt.Printf("\n関数内でディアクティベートします: %s\n", user.Name)
	user.IsActive = false
}

func main() {
	user1 := User{
		ID:       1,
		Name:     "Yamada",
		IsActive: true,
	}
	fmt.Println("変更前のユーザー状態:", user1)

	deactivate(&user1)

	fmt.Println("変更後のユーザー状態:", user1)
}