package main

import (
	"fmt"

	"log"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Hashed password for '%s': %s\n", password, hashedPassword)
}
