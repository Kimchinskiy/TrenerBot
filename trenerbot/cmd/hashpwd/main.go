package main

import (
	"fmt"
	"os"

	"trenerbot/internal/auth"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run cmd/hashpwd/main.go <password>")
		return
	}
	h, err := auth.HashPassword(os.Args[1])
	if err != nil {
		panic(err)
	}
	fmt.Print(h)
}
