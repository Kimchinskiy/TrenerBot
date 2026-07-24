package main

import (
	"encoding/json"
	"fmt"
	"log"

	"trenerbot/internal/db"
	"trenerbot/internal/service"
	"trenerbot/internal/store"
)

func main() {
	database, err := db.Open("data/crm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	s := store.New(database)
	svc := service.New(s, nil)

	for _, period := range []string{"week", "month", "year"} {
		stats, err := svc.GetStatistics(service.StatisticsRequest{
			Period:  period,
			CoachID: 1,
		})
		if err != nil {
			log.Printf("Error for %s: %v", period, err)
			continue
		}
		b, _ := json.MarshalIndent(stats, "", "  ")
		fmt.Printf("\n=== Period: %s ===\n", period)
		fmt.Println(string(b))
	}
}
