package store

import (
	"database/sql"
	"time"

	"trenerbot/internal/domain"
)

func (s *Store) TrainingCount(from, to string, coachID int64) (int, error) {
	var count int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM lesson_entries WHERE date >= ? AND date <= ? AND status = 'done' AND coach_id = ?`,
		from, to, coachID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) NewClientsCount(from, to string) (int, error) {
	var count int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM clients WHERE registered_at >= ? AND registered_at <= ? AND full_name != ''`,
		from, to).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) ActiveClientsCount() (int, error) {
	var count int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM (SELECT id FROM clients WHERE status = 'active' AND full_name != '' GROUP BY CASE WHEN user_id IS NULL THEN full_name ELSE CAST(user_id AS TEXT) END)`).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) SubscriptionIncome(from, to string) (float64, error) {
	var total sql.NullFloat64
	err := s.DB.QueryRow(`SELECT COALESCE(SUM(price), 0) FROM subscriptions WHERE bought_at >= ? AND bought_at <= ?`,
		from, to).Scan(&total)
	if err != nil {
		return 0, err
	}
	return total.Float64, nil
}

func (s *Store) AttendanceRate(from, to string) (float64, error) {
	var rate sql.NullFloat64
	err := s.DB.QueryRow(`SELECT COALESCE(SUM(CASE WHEN present = 1 THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0), 0) FROM daily_attendance WHERE date >= ? AND date <= ?`,
		from, to).Scan(&rate)
	if err != nil {
		return 0, err
	}
	return rate.Float64, nil
}

func (s *Store) CanceledCount(from, to string, coachID int64) (int, error) {
	var count int
	err := s.DB.QueryRow(`SELECT COUNT(*) FROM lesson_entries WHERE date >= ? AND date <= ? AND status = 'canceled' AND coach_id = ?`,
		from, to, coachID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) IncomeChartData(from, to string) ([]domain.ChartPoint, error) {
	rows, err := s.DB.Query(`SELECT bought_at, COALESCE(SUM(price), 0) FROM subscriptions WHERE bought_at >= ? AND bought_at <= ? GROUP BY bought_at ORDER BY bought_at`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []domain.ChartPoint
	for rows.Next() {
		var p domain.ChartPoint
		if err := rows.Scan(&p.Label, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}
	return points, rows.Err()
}

func (s *Store) BusiestDay(from, to string, coachID int64) (string, error) {
	var day string
	var cnt int
	err := s.DB.QueryRow(`SELECT CASE CAST(strftime('%w', date) AS INTEGER) WHEN 0 THEN 'Вс' WHEN 1 THEN 'Пн' WHEN 2 THEN 'Вт' WHEN 3 THEN 'Ср' WHEN 4 THEN 'Чт' WHEN 5 THEN 'Пт' WHEN 6 THEN 'Сб' END, COUNT(*) FROM lesson_entries WHERE date >= ? AND date <= ? AND coach_id = ? GROUP BY strftime('%w', date) ORDER BY 2 DESC LIMIT 1`,
		from, to, coachID).Scan(&day, &cnt)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return day, nil
}

func (s *Store) PopularTime(from, to string, coachID int64) (string, error) {
	var t string
	var cnt int
	err := s.DB.QueryRow(`SELECT time, COUNT(*) FROM lesson_entries WHERE date >= ? AND date <= ? AND coach_id = ? GROUP BY time ORDER BY 2 DESC LIMIT 1`,
		from, to, coachID).Scan(&t, &cnt)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return t, nil
}

func (s *Store) AvgGroupSize() (float64, error) {
	var avg sql.NullFloat64
	err := s.DB.QueryRow(`SELECT COALESCE(AVG(cnt), 0) FROM (SELECT group_id, COUNT(*) as cnt FROM group_members GROUP BY group_id)`).Scan(&avg)
	if err != nil {
		return 0, err
	}
	return avg.Float64, nil
}

func (s *Store) DebtorClients() ([]struct {
	ClientID int64
	FullName string
	Phone    sql.NullString
	Debt     float64
	EndsAt   sql.NullString
}, error) {
	rows, err := s.DB.Query(`
		SELECT c.id, c.full_name, c.phone,
			COALESCE(s.price, 0),
			s.ends_at
		FROM clients c
		LEFT JOIN (
			SELECT client_id, id, price, ends_at
			FROM subscriptions
			WHERE id IN (SELECT MAX(id) FROM subscriptions GROUP BY client_id)
		) s ON s.client_id = c.id
		WHERE c.status = 'active' AND c.full_name != ''
			AND (s.ends_at IS NULL OR s.ends_at < date('now'))
		ORDER BY c.full_name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []struct {
		ClientID int64
		FullName string
		Phone    sql.NullString
		Debt     float64
		EndsAt   sql.NullString
	}
	for rows.Next() {
		var it struct {
			ClientID int64
			FullName string
			Phone    sql.NullString
			Debt     float64
			EndsAt   sql.NullString
		}
		if err := rows.Scan(&it.ClientID, &it.FullName, &it.Phone, &it.Debt, &it.EndsAt); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *Store) AvgAttendanceRate() (float64, error) {
	var rate sql.NullFloat64
	err := s.DB.QueryRow(`SELECT COALESCE(AVG(rate), 0) FROM (SELECT date, SUM(CASE WHEN present = 1 THEN 1 ELSE 0 END) * 100.0 / NULLIF(COUNT(*), 0) as rate FROM daily_attendance GROUP BY date)`).Scan(&rate)
	if err != nil {
		return 0, err
	}
	return rate.Float64, nil
}

func parseDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return time.Now()
	}
	return t
}
