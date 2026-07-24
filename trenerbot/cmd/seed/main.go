package main

import (
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
				slog.Info("admin created", "id", id)
			}
		} else {
			slog.Info("admin already exists", "id", id)
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
	slog.Info("Seeding test data")

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

	// Create lesson entries for today to end of current week (per athlete)
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	} else {
		weekday = time.Monday - weekday + 7
	}
	endOfWeek := now.AddDate(0, 0, int(7-weekday))
	for d := 0; d <= int(endOfWeek.Sub(now).Hours()/24); d++ {
		date := now.AddDate(0, 0, d).Format("2006-01-02")

		for j, clientID := range clientIDs {
			timeStr := "09:00"
			if j%2 == 1 {
				timeStr = "10:00"
			}
			entry := domain.LessonEntry{
				Date:     date,
				Time:     timeStr,
				ClientID: clientID,
				CoachID:  &coach.ID,
				Duration: 60,
				Status:   domain.LessonPlanned,
			}
			id, _ := s.InsertLessonEntry(entry)
			fmt.Printf("Created lesson entry: %s %s client=%d entry#%d\n", date, timeStr, clientID, id)
		}
	}

	// Create historical (completed) entries for past days
	pastDate := now.AddDate(0, 0, -3).Format("2006-01-02")
	pastTimes := []string{"09:00", "10:00", "11:00"}
	for i, clientID := range clientIDs[:3] {
		entry := domain.LessonEntry{
			Date:     pastDate,
			Time:     pastTimes[i],
			ClientID: clientID,
			CoachID:  &coach.ID,
			Duration: 60,
			Status:   domain.LessonDone,
		}
		id, _ := s.InsertLessonEntry(entry)
		fmt.Printf("Created past lesson entry: %s %s client=%d entry#%d\n", pastDate, pastTimes[i], clientID, id)
	}

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