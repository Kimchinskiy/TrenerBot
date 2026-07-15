package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"trenerbot/internal/config"
	"trenerbot/internal/db"
	"trenerbot/internal/domain"
	"trenerbot/internal/store"
)

// seed creates the initial admin and coach accounts from env (TELEGRAM_ID values).
func main() {
	_, _ = config.Load()
	_ = godotenvLoad()

	database, err := db.Open(os.Getenv("DB_PATH"))
	if err != nil {
		slog.Error("db", "err", err)
		os.Exit(1)
	}
	defer database.Close()
	s := store.New(database)

	// Admin
	if id := os.Getenv("ADMIN_TELEGRAM_ID"); id != "" {
		if u, _ := s.UserByTelegram(id); u == nil {
			if _, err := s.CreateUser(&id, domain.RoleAdmin); err != nil {
				slog.Error("create admin", "err", err)
			} else {
				fmt.Println("admin created:", id)
			}
		} else {
			fmt.Println("admin already exists:", id)
		}
	}

	// Coach
	var coachUserID int64
	if id := os.Getenv("COACH_TELEGRAM_ID"); id != "" {
		if u, _ := s.UserByTelegram(id); u == nil {
			uid, err := s.CreateUser(&id, domain.RoleCoach)
			if err != nil {
				slog.Error("create coach user", "err", err)
				os.Exit(1)
			}
			co := domain.Coach{UserID: &uid, FullName: os.Getenv("COACH_NAME")}
			if co.FullName == "" {
				co.FullName = "Тренер"
			}
			if _, err := s.CreateCoach(co); err != nil {
				slog.Error("create coach", "err", err)
				os.Exit(1)
			}
			coachUserID = uid
			fmt.Println("coach created:", id)
		} else {
			fmt.Println("coach already exists:", id)
			coachUserID = u.ID
		}
	}

	// Create test clients and lessons if requested
	if os.Getenv("SEED_TEST_DATA") == "1" {
		seedTestData(s, coachUserID)
	}
}

func seedTestData(s *store.Store, coachUserID int64) {
	fmt.Println("Seeding test data...")

	// Get coach ID
	coach, err := s.CoachByUserID(coachUserID)
	if err != nil || coach == nil {
		slog.Error("get coach", "err", err)
		return
	}

	// Create test clients
	clients := []domain.Client{
		{UserID: nil, FullName: "Анна Петрова", Phone: strPtr("+7 900 111-22-33"), Age: intPtr(28), Status: "active", Source: strPtr("telegram_bot")},
		{UserID: nil, FullName: "Иван Сидоров", Phone: strPtr("+7 900 222-33-44"), Age: intPtr(35), Status: "active", Source: strPtr("telegram_bot")},
		{UserID: nil, FullName: "Мария Козлова", Phone: strPtr("+7 900 333-44-55"), Age: intPtr(24), Status: "active", Source: strPtr("telegram_bot")},
		{UserID: nil, FullName: "Дмитрий Волков", Phone: strPtr("+7 900 444-55-66"), Age: intPtr(30), Status: "active", Source: strPtr("telegram_bot")},
		{UserID: nil, FullName: "Елена Смирнова", Phone: strPtr("+7 900 555-66-77"), Age: intPtr(27), Status: "active", Source: strPtr("telegram_bot")},
	}

	var clientIDs []int64
	for _, c := range clients {
		id, err := s.CreateClient(c)
		if err != nil {
			slog.Error("create client", "err", err)
			continue
		}
		clientIDs = append(clientIDs, id)
		fmt.Printf("Created client: %s (ID: %d)\n", c.FullName, id)
	}

	// Create lessons for today and next 7 days
	today := time.Now()
	for i := 0; i < 7; i++ {
		date := today.AddDate(0, 0, i).Format("2006-01-02")

		// Morning lesson
		lesson1 := domain.Lesson{
			Date:     date,
			Time:     "09:00",
			CoachID:  &coach.ID,
			Duration: 60,
			Status:   domain.LessonPlanned,
			Location: strPtr("Зал А"),
			Comment:  strPtr("Утренняя группа"),
		}
		id1, _ := s.CreateLesson(lesson1)

		// Evening lesson
		lesson2 := domain.Lesson{
			Date:     date,
			Time:     "19:00",
			CoachID:  &coach.ID,
			Duration: 90,
			Status:   domain.LessonPlanned,
			Location: strPtr("Зал Б"),
			Comment:  strPtr("Вечерняя группа"),
		}
		id2, _ := s.CreateLesson(lesson2)

		// Register clients for lessons (alternating)
		for j, clientID := range clientIDs {
			if (i+j)%2 == 0 {
				s.SetAttendance(id1, clientID, false, nil)
			}
			if (i+j+1)%3 == 0 {
				s.SetAttendance(id2, clientID, false, nil)
			}
		}

		fmt.Printf("Created lessons for %s: #%d (09:00), #%d (19:00)\n", date, id1, id2)
	}

	// Mark some attendance for past days (simulate history)
	pastDate := today.AddDate(0, 0, -3).Format("2006-01-02")
	pastLesson := domain.Lesson{
		Date:     pastDate,
		Time:     "10:00",
		CoachID:  &coach.ID,
		Duration: 60,
		Status:   domain.LessonDone,
		Location: strPtr("Зал А"),
	}
	pastLessonID, _ := s.CreateLesson(pastLesson)
	for _, clientID := range clientIDs[:3] {
		s.SetAttendance(pastLessonID, clientID, true, &coachUserID)
	}
	fmt.Printf("Created past lesson #%d with attendance\n", pastLessonID)

	// Create waiting list entries
	for i, clientID := range clientIDs[:2] {
		s.DB.Exec(`INSERT INTO waiting_list(client_id, group_id, position, created_at) VALUES (?, NULL, ?, datetime('now'))`, clientID, i+1)
	}
	fmt.Println("Created waiting list entries")

	fmt.Println("Test data seeding complete!")
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int { return &i }

func godotenvLoad() error {
	if _, err := os.Stat(".env"); err != nil {
		return nil
	}
	data, err := os.ReadFile(".env")
	if err != nil {
		return err
	}
	for _, line := range splitLines(string(data)) {
		if line == "" || line[0] == '#' {
			continue
		}
		var k, v string
		for i := 0; i < len(line); i++ {
			if line[i] == '=' {
				k, v = line[:i], line[i+1:]
				break
			}
		}
		if k != "" {
			_ = os.Setenv(k, v)
		}
	}
	return nil
}

func splitLines(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == '\n' {
			out = append(out, cur)
			cur = ""
		} else {
			cur += string(r)
		}
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}