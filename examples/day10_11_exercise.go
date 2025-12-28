package main

import "fmt"

// User構造体を定義
type User struct {
	ID       int
	Name     string
	IsActive bool
}

// deactivate関数
// 引数に *User と書くことで、「User型のポインタ」を受け取ることを示す
func deactivate(user *User) {
	// ポインタを通じて、元のuserオブジェクトのIsActiveフィールドを直接変更する
	fmt.Printf("\n関数内でディアクティベートします: %s\n", user.Name)
	user.IsActive = false
	// 注意: Goでは、C言語などと違い、ポインタ経由のフィールドアクセスでも
	// (*user).IsActive のように書く必要はなく、 user.IsActive とシンプルに書けます。
	// コンパイラが自動で解釈してくれます。
}

func main() {
	// ユーザーを作成
	user1 := User{
		ID:       1,
		Name:     "Yamada",
		IsActive: true,
	}
	fmt.Println("変更前のユーザー状態:", user1)

	// deactivate関数を呼び出す
	// この時、user1そのものではなく、そのメモリアドレス(&user1)を渡す
	deactivate(&user1)

	// deactivate関数は何も返していないが、user1の状態が変更されていることを確認
	fmt.Println("変更後のユーザー状態:", user1)
}
