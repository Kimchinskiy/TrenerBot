package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"trenerbot/internal/db"
)

func main() {
	database, err := db.Open("data/crm.db")
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Check existing data
	var lessonCount int
	database.QueryRow("SELECT COUNT(*) FROM lesson_entries").Scan(&lessonCount)

	if lessonCount > 10 {
		fmt.Printf("Data already exists (%d lessons), re-seeding...\n", lessonCount)
		// Clean and re-seed
		database.Exec("DELETE FROM daily_attendance")
		database.Exec("DELETE FROM lesson_entries")
		database.Exec("DELETE FROM subscriptions")
		database.Exec("DELETE FROM group_members")
		database.Exec("DELETE FROM groups WHERE id > 0")
		database.Exec("DELETE FROM clients WHERE id >= 2")
	} else {
		fmt.Println("Seeding test data...")
	}

	// Also ensure clean state for coaches/users
	database.Exec("DELETE FROM coaches")
	database.Exec("DELETE FROM users")

	// 1. User (coach)
	database.Exec(`INSERT OR IGNORE INTO users(id, role, first_name, created_at) VALUES (1, 'coach', 'Анна', datetime('now'))`)

	// 2. Coach
	database.Exec(`INSERT OR IGNORE INTO coaches(id, user_id, full_name) VALUES (1, 1, 'Анна Тренер')`)
	database.Exec(`UPDATE users SET updated_at = datetime('now') WHERE id = 1`)

	// 3. Clients
	clients := []struct {
		id       int
		name     string
		status   string
		regDate  string
		phone    string
		hasSub   bool
		subType  string
		subPrice float64
		subEnds  string
		lessons  int
	}{
		{2, "Иван Петров", "active", "2026-06-01", "+7 (999) 111-22-33", true, "count", 5000, "2026-08-15", 12},
		{3, "Мария Иванова", "active", "2026-06-05", "+7 (999) 222-33-44", true, "period", 8000, "2026-09-01", 0},
		{4, "Алексей Смирнов", "active", "2026-06-10", "+7 (999) 333-44-55", true, "count", 4500, "2026-07-20", 5},
		{5, "Елена Козлова", "active", "2026-06-12", "+7 (999) 444-55-66", false, "", 0, "", 0},
		{6, "Дмитрий Новиков", "active", "2026-06-15", "+7 (999) 555-66-77", true, "count", 6000, "2026-08-30", 8},
		{7, "Ольга Фёдорова", "active", "2026-06-20", "+7 (999) 666-77-88", true, "period", 10000, "2026-10-01", 0},
		{8, "Сергей Морозов", "active", "2026-07-01", "+7 (999) 777-88-99", false, "", 0, "", 0},
		{9, "Наталья Кузнецова", "active", "2026-07-05", "+7 (999) 888-99-00", true, "count", 5500, "2026-07-25", 3},
		{10, "Павел Захаров", "active", "2026-07-10", "+7 (999) 000-11-22", true, "count", 7000, "2026-09-15", 10},
		{11, "Анна Белова", "active", "2026-07-12", "+7 (999) 111-22-44", false, "", 0, "", 0},
		{12, "Михаил Соколов", "paused", "2026-06-18", "+7 (999) 222-55-77", false, "", 0, "", 0},
		{13, "Татьяна Орлова", "active", "2026-07-15", "+7 (999) 333-66-88", true, "count", 5000, "2026-07-10", 0},
	}

	for _, c := range clients {
		_, err := database.Exec(`INSERT OR IGNORE INTO clients(id, full_name, status, registered_at, phone) VALUES (?, ?, ?, ?, ?)`,
			c.id, c.name, c.status, c.regDate, c.phone)
		if err != nil {
			log.Printf("client %d: %v", c.id, err)
		}
	}

	// 4. Subscriptions
	subID := 1
	for _, c := range clients {
		if !c.hasSub {
			continue
		}
		boughtAt := c.regDate
		database.Exec(`INSERT OR IGNORE INTO subscriptions(id, client_id, type, price, bought_at, ends_at, lessons_left) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			subID, c.id, c.subType, c.subPrice, boughtAt, c.subEnds, c.lessons)
		subID++
	}

	// 5. Groups
	groups := []struct{ id int; name string }{
		{1, "Утренняя группа"},
		{2, "Вечерняя группа"},
		{3, "Выходного дня"},
	}
	for _, g := range groups {
		database.Exec(`INSERT OR IGNORE INTO groups(id, name, coach_id, active) VALUES (?, ?, 1, 1)`, g.id, g.name)
	}

	// 6. Group members
	members := []struct{ gID, cID int }{
		{1, 2}, {1, 3}, {1, 4}, {1, 6}, {1, 10},
		{2, 5}, {2, 7}, {2, 8}, {2, 9}, {2, 11},
		{3, 2}, {3, 7}, {3, 13},
	}
	for _, m := range members {
		database.Exec(`INSERT OR IGNORE INTO group_members(group_id, client_id, role) VALUES (?, ?, 'member')`, m.gID, m.cID)
	}

	// 7. Lesson entries (last 60 days)
	clientIDs := []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 13}
	now := time.Now()
	rng := rand.New(rand.NewSource(42))
	lessonID := 1

	for dayOffset := 60; dayOffset >= 0; dayOffset-- {
		date := now.AddDate(0, 0, -dayOffset)
		dateStr := date.Format("2006-01-02")
		weekday := date.Weekday()
		if weekday == time.Sunday {
			continue
		}
		sessions := []string{"09:00", "10:00", "18:00", "19:00"}
		for _, session := range sessions {
			if rng.Float64() > 0.6 {
				continue
			}
			n := 2 + rng.Intn(4)
			rng.Shuffle(len(clientIDs), func(i, j int) { clientIDs[i], clientIDs[j] = clientIDs[i], clientIDs[j] })
			for i := 0; i < n && i < len(clientIDs); i++ {
				status := "done"
				if rng.Float64() < 0.1 {
					status = "canceled"
				}
				_, err := database.Exec(`INSERT OR IGNORE INTO lesson_entries(id, date, time, client_id, coach_id, duration, status) VALUES (?, ?, ?, ?, 1, 60, ?)`,
					lessonID, dateStr, session, clientIDs[i], status)
				if err != nil {
					log.Printf("lesson %d: %v", lessonID, err)
				}
				lessonID++
			}
		}
	}

	// 8. Daily attendance (last 30 days)
	rows, err := database.Query(`SELECT id, date, client_id FROM lesson_entries WHERE status = 'done' AND date >= date('now', '-30 days')`)
	if err != nil {
		log.Fatal(err)
	}
	var attrs []struct{ id, cid int; date string }
	for rows.Next() {
		var a struct{ id, cid int; date string }
		rows.Scan(&a.id, &a.date, &a.cid)
		attrs = append(attrs, a)
	}
	rows.Close()

	for _, a := range attrs {
		present := 1
		if rng.Float64() < 0.08 {
			present = 0
		}
		database.Exec(`INSERT OR IGNORE INTO daily_attendance(date, client_id, present, marked_by) VALUES (?, ?, ?, 1)`,
			a.date, a.cid, present)
	}

	// 9. Extra subscriptions for richer income chart
	extraSubs := []struct {
		clientID int
		price    float64
		daysAgo  int
	}{
		{2, 5000, 45}, {3, 8000, 40}, {4, 4500, 35}, {6, 6000, 30},
		{7, 10000, 25}, {9, 5500, 20}, {10, 7000, 15}, {13, 5000, 10},
		{2, 5000, 50}, {5, 4000, 48}, {8, 3500, 42}, {11, 6000, 38},
		{3, 8000, 28}, {4, 4500, 22}, {6, 6000, 12}, {7, 10000, 5},
		{10, 7000, 55}, {13, 5000, 52},
	}
	for _, s := range extraSubs {
		date := now.AddDate(0, 0, -s.daysAgo).Format("2006-01-02")
		endsAt := now.AddDate(0, 0, 30-s.daysAgo).Format("2006-01-02")
		database.Exec(`INSERT OR IGNORE INTO subscriptions(id, client_id, type, price, bought_at, ends_at, lessons_left) VALUES (?, ?, 'count', ?, ?, ?, ?)`,
			subID, s.clientID, s.price, date, endsAt, 8+rng.Intn(8))
		subID++
	}

	fmt.Println("=== Seed data inserted ===")
	fmt.Printf("  Clients: %d\n", len(clients))
	fmt.Printf("  Lesson entries: %d\n", lessonID-1)
	fmt.Printf("  Attendance records: %d\n", len(attrs))
	fmt.Printf("  Groups: %d\n", len(groups))
}
