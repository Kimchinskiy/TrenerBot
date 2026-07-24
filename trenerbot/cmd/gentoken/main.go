package main

import (
	"fmt"
	"time"

	"trenerbot/internal/auth"
	"trenerbot/internal/domain"
)

func main() {
	ts := auth.NewTokenService("change-me-in-prod", 24*time.Hour)
	tok, err := ts.GenerateAccess(1, domain.RoleCoach)
	if err != nil {
		panic(err)
	}
	fmt.Print(tok)
}
