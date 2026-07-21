package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"trenerbot/internal/auth"
	"trenerbot/internal/domain"
	"trenerbot/internal/store"
)

type Services struct {
	Store  *store.Store
	Tokens *auth.TokenService
}

func New(store *store.Store, tokens *auth.TokenService) *Services {
	return &Services{Store: store, Tokens: tokens}
}

// ---------- Auth / registration ----------

type TelegramLoginRequest struct {
	TelegramID    string `json:"telegram_id"`
	FullName      string `json:"full_name"`
	Phone         string `json:"phone"`
	Age           int    `json:"age"`
	MedicalLimits string `json:"medical_limits"`
	Source        string `json:"source"`
}

// TelegramLogin resolves (or creates) a client account linked to a Telegram ID and returns a JWT.
func (s *Services) TelegramLogin(req TelegramLoginRequest) (*domain.User, *domain.Client, string, error) {
	u, err := s.Store.UserByTelegram(req.TelegramID)
	if err != nil {
		return nil, nil, "", err
	}
	if u == nil {
		uid, err := s.Store.CreateUser(&req.TelegramID, domain.RoleClient)
		if err != nil {
			return nil, nil, "", err
		}
		src := req.Source
		c := domain.Client{
			UserID:        &uid,
			FullName:      req.FullName,
			Phone:         nullable(req.Phone),
			Age:           nullableInt(req.Age),
			MedicalLimits: nullable(req.MedicalLimits),
			Status:        "active",
			Source:        nullable(src),
		}
		cid, err := s.Store.CreateClient(c)
		if err != nil {
			return nil, nil, "", err
		}
		u = &domain.User{ID: uid, TelegramID: &req.TelegramID, Role: domain.RoleClient}
		client, _ := s.Store.ClientByID(cid)
		tok, err := s.Tokens.Generate(uid, u.Role)
		if err != nil {
			return nil, nil, "", err
		}
		// notify all coaches about a new client
		_ = s.notifyCoaches("new_client", map[string]any{"user_id": uid, "name": req.FullName})
		return u, client, tok, nil
	}

	// existing user: refresh client profile if provided
	if u.Role == domain.RoleClient {
		c, err := s.Store.ClientByUserID(u.ID)
		if err != nil {
			return nil, nil, "", err
		}
		if c != nil && req.FullName != "" {
			c.FullName = req.FullName
			if req.Phone != "" {
				v := req.Phone
				c.Phone = &v
			}
			if req.Age > 0 {
				v := req.Age
				c.Age = &v
			}
			if req.MedicalLimits != "" {
				v := req.MedicalLimits
				c.MedicalLimits = &v
			}
			_ = s.Store.UpdateClient(*c)
		}
	}
	tok, err := s.Tokens.Generate(u.ID, u.Role)
	if err != nil {
		return nil, nil, "", err
	}
	client, _ := s.Store.ClientByUserID(u.ID)
	return u, client, tok, nil
}

func (s *Services) notifyCoaches(typ string, payload map[string]any) error {
	coaches, err := s.Store.ListCoaches()
	if err != nil {
		return err
	}
	b, _ := json.Marshal(payload)
	for _, co := range coaches {
		if co.UserID == nil {
			continue
		}
		_ = s.enqueue(*co.UserID, typ, string(b), time.Now())
	}
	return nil
}

func (s *Services) enqueue(userID int64, typ, payload string, sendAt time.Time) error {
	_, err := s.Store.InsertNotification(domain.Notification{
		Channel:         "telegram",
		RecipientUserID: userID,
		Type:            typ,
		Payload:         payload,
		SendAt:          sendAt,
	})
	return err
}

// SetTelegramID links a Telegram ID to an existing user (used by admin to enable coach/admin bot access).
func (s *Services) SetTelegramID(userID int64, telegramID string) error {
	_, err := s.Store.DB.Exec(`UPDATE users SET telegram_id = ? WHERE id = ?`, telegramID, userID)
	return err
}

// ---------- Clients ----------

func (s *Services) GetClient(id int64) (*domain.Client, error) { return s.Store.ClientByID(id) }

func (s *Services) ListClients() ([]domain.Client, error) { return s.Store.ListClients() }

func (s *Services) UpdateClient(c domain.Client) error { return s.Store.UpdateClient(c) }

func (s *Services) ChildrenOfParent(parentUserID int64) ([]domain.Client, error) {
	return s.Store.ChildrenOfParent(parentUserID)
}

// ---------- Coaches ----------

func (s *Services) CreateCoach(co domain.Coach) (int64, error) { return s.Store.CreateCoach(co) }

func (s *Services) ListCoaches() ([]domain.Coach, error) { return s.Store.ListCoaches() }

func (s *Services) CoachByUser(userID int64) (*domain.Coach, error) { return s.Store.CoachByUserID(userID) }

// ---------- Lessons ----------

func (s *Services) CreateLesson(l domain.Lesson) (int64, error) {
	if l.Status == "" {
		l.Status = domain.LessonPlanned
	}
	return s.Store.CreateLesson(l)
}

func (s *Services) GetLesson(id int64) (*domain.Lesson, error) { return s.Store.LessonByID(id) }

func (s *Services) ListWeekLessons(coachID int64, from, to string) ([]domain.Lesson, error) {
	return s.Store.ListLessons(from, to, coachID)
}

func (s *Services) SetLessonStatus(id int64, status domain.LessonStatus) error {
	return s.Store.UpdateLessonStatus(id, status)
}

// ClientSchedule returns lessons the client is registered for (via attendance rows).
func (s *Services) ClientSchedule(clientID int64, from string) ([]domain.Lesson, error) {
	rows, err := s.Store.DB.Query(`SELECT l.id, l.date, l.time, l.coach_id, l.duration, l.status, l.location, l.comment, l.group_id
		FROM lessons l JOIN attendance a ON a.lesson_id = l.id WHERE a.client_id = ? AND l.date >= ? ORDER BY l.date, l.time`, clientID, from)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLessons(rows)
}

// ---------- Attendance ----------

func (s *Services) MarkAttendance(lessonID, clientID int64, present bool, markedBy *int64) error {
	if err := s.Store.SetAttendance(lessonID, clientID, present, markedBy); err != nil {
		return err
	}
	typ := "visit"
	if !present {
		typ = "absence"
	}
	return s.Store.LogActivity(clientID, typ, lessonID, "", markedBy)
}

func (s *Services) ListAttendance(lessonID int64) ([]domain.Attendance, error) {
	return s.Store.ListAttendanceByLesson(lessonID)
}

// RegisterClientToLesson adds a client to a lesson (запись).
func (s *Services) RegisterClientToLesson(lessonID, clientID int64) error {
	return s.Store.SetAttendance(lessonID, clientID, false, nil)
}

// ---------- Lesson entries (new schedule model) ----------

func (s *Services) GetSchedule(from, to string, role domain.Role, userID, coachID, clientID int64) ([]domain.ScheduleEntry, error) {
	switch role {
	case domain.RoleCoach:
		return s.Store.ListScheduleEntries(from, to, coachID, 0)
	case domain.RoleClient:
		return s.Store.ListScheduleEntries(from, to, 0, clientID)
	default:
		return s.Store.ListScheduleEntries(from, to, 0, 0)
	}
}

// ---------- Lesson change notifications ----------

// NotifyLessonChange notifies all registered participants about a lesson change.
func (s *Services) NotifyLessonChange(lessonID int64, changeType string, oldVal, newVal string) error {
	parts, err := s.Store.LessonParticipants(lessonID)
	if err != nil {
		return err
	}
	lesson, err := s.Store.LessonByID(lessonID)
	if err != nil || lesson == nil {
		return err
	}
	for _, cid := range parts {
		c, err := s.Store.ClientByID(cid)
		if err != nil || c == nil || c.UserID == nil {
			continue
		}
		payload := map[string]any{
			"lesson_id":   lessonID,
			"date":        lesson.Date,
			"time":        lesson.Time,
			"location":    lesson.Location,
			"change_type": changeType,
			"old_value":   oldVal,
			"new_value":   newVal,
		}
		b, _ := json.Marshal(payload)
		_ = s.enqueue(*c.UserID, "lesson_change", string(b), time.Now())
	}
	return nil
}

// NotifyLessonCanceled notifies participants when a lesson is canceled.
func (s *Services) NotifyLessonCanceled(lessonID int64, reason string) error {
	parts, err := s.Store.LessonParticipants(lessonID)
	if err != nil {
		return err
	}
	lesson, err := s.Store.LessonByID(lessonID)
	if err != nil || lesson == nil {
		return err
	}
	for _, cid := range parts {
		c, err := s.Store.ClientByID(cid)
		if err != nil || c == nil || c.UserID == nil {
			continue
		}
		payload := map[string]any{
			"lesson_id": lessonID,
			"date":      lesson.Date,
			"time":      lesson.Time,
			"location":  lesson.Location,
			"reason":    reason,
		}
		b, _ := json.Marshal(payload)
		_ = s.enqueue(*c.UserID, "lesson_canceled", string(b), time.Now())
	}
	return nil
}

// ReapStaleClaimed returns 'claimed' rows that were never acknowledged within timeout back to 'pending'.
func (s *Services) ReapStaleClaimed(timeout time.Duration) (int64, error) {
	cutoff := time.Now().Add(-timeout).Format("2006-01-02 15:04:05")
	res, err := s.Store.DB.Exec(`UPDATE notifications SET status='pending', claim_token=NULL
		WHERE status='claimed' AND sent_at < ?`, cutoff)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}

// EnsureReminders scans upcoming lessons and enqueues a day-of 08:00 reminder for each
// registered client (idempotent via NotificationExists). ТЗ §9.
func (s *Services) EnsureReminders() (int, error) {
	now := time.Now()
	from := now.Format("2006-01-02")
	to := now.AddDate(0, 0, 1).Format("2006-01-02")
	entries, err := s.Store.ListLessonEntriesForReminders(from, to)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, e := range entries {
		if e.UserID == 0 {
			continue
		}
		sendAt, err := time.Parse("2006-01-02 15:04:05", e.Date+" 08:00:00")
		if err != nil || sendAt.Before(now) {
			continue
		}
		payload, _ := json.Marshal(map[string]any{
			"lesson_id": e.LessonID,
			"date":      e.Date,
			"time":      e.Time,
		})
		exists, err := s.Store.NotificationExists("lesson_reminder", e.UserID, string(payload))
		if err != nil || exists {
			continue
		}
		if err := s.enqueue(e.UserID, "lesson_reminder", string(payload), sendAt); err == nil {
			count++
		}
	}
	return count, nil
}

// ---------- Waiting List ----------

// AddToWaitingList adds a client to the waiting list.
func (s *Services) AddToWaitingList(clientID int64, groupID *int64) (int64, error) {
	// Get next position
	var pos int
	err := s.Store.DB.QueryRow(`SELECT COALESCE(MAX(position), 0) + 1 FROM waiting_list WHERE group_id IS ?`, groupID).Scan(&pos)
	if err != nil {
		pos = 1
	}
	res, err := s.Store.DB.Exec(`INSERT INTO waiting_list(client_id, group_id, position, created_at) VALUES (?, ?, ?, datetime('now'))`, clientID, groupID, pos)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// WaitingList is an alias for GetWaitingList for API compatibility.
func (s *Services) WaitingList(groupID *int64) ([]map[string]any, error) {
	return s.GetWaitingList(groupID)
}

// GetWaitingList returns the waiting list for a group (or all if groupID is nil).
func (s *Services) GetWaitingList(groupID *int64) ([]map[string]any, error) {
	q := `SELECT wl.id, wl.client_id, wl.group_id, wl.position, wl.created_at, c.full_name, c.phone
		FROM waiting_list wl JOIN clients c ON c.id = wl.client_id`
	args := []any{}
	if groupID != nil {
		q += ` WHERE wl.group_id = ?`
		args = append(args, *groupID)
	}
	q += ` ORDER BY wl.position`
	rows, err := s.Store.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, clientID int64
		var groupID sql.NullInt64
		var pos int
		var created, name, phone string
		if err := rows.Scan(&id, &clientID, &groupID, &pos, &created, &name, &phone); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":         id,
			"client_id":  clientID,
			"group_id":   groupID,
			"position":   pos,
			"created_at": created,
			"name":       name,
			"phone":      phone,
		})
	}
	return out, rows.Err()
}

// RemoveFromWaitingList removes a client from the waiting list.
func (s *Services) RemoveFromWaitingList(id int64) error {
	_, err := s.Store.DB.Exec(`DELETE FROM waiting_list WHERE id = ?`, id)
	return err
}

// MoveFromWaitingList moves the first person from waiting list to a lesson.
func (s *Services) MoveFromWaitingList(lessonID int64) error {
	// Get first in waiting list
	var wlID, clientID int64
	err := s.Store.DB.QueryRow(`SELECT id, client_id FROM waiting_list WHERE group_id IS NULL ORDER BY position LIMIT 1`).Scan(&wlID, &clientID)
	if err != nil {
		return err // no one in waiting list
	}
	// Register to lesson
	if err := s.Store.SetAttendance(lessonID, clientID, false, nil); err != nil {
		return err
	}
	// Remove from waiting list
	_, _ = s.Store.DB.Exec(`DELETE FROM waiting_list WHERE id = ?`, wlID)
	// Reorder positions
	_, _ = s.Store.DB.Exec(`UPDATE waiting_list SET position = position - 1 WHERE position > (SELECT position FROM waiting_list WHERE id = ?)`, wlID)
	return nil
}

// ---------- Analytics / Debtors ----------

// DebtorsWidget returns clients with missed payments/attendance grouped by days.
func (s *Services) DebtorsWidget(days int) ([]map[string]any, error) {
	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	rows, err := s.Store.DB.Query(`
		SELECT c.id, c.full_name, c.phone, c.status,
		       COUNT(CASE WHEN a.present = 0 AND l.date >= ? THEN 1 END) as missed_count,
		       GROUP_CONCAT(l.date || ' ' || l.time) as missed_dates
		FROM clients c
		LEFT JOIN attendance a ON a.client_id = c.id
		LEFT JOIN lessons l ON l.id = a.lesson_id
		WHERE c.status = 'active' AND l.date >= ?
		GROUP BY c.id
		HAVING missed_count > 0
		ORDER BY missed_count DESC
	`, since, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id int64
		var name, phone, status, missedDates string
		var missedCount int
		if err := rows.Scan(&id, &name, &phone, &status, &missedCount, &missedDates); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"client_id":    id,
			"name":         name,
			"phone":        phone,
			"status":       status,
			"missed_count": missedCount,
			"missed_dates": missedDates,
		})
	}
	return out, rows.Err()
}

// ClientWellbeingFeedback records wellbeing feedback from client after training.
func (s *Services) ClientWellbeingFeedback(clientID int64, lessonID int64, wellbeing int, note string) error {
	_, err := s.Store.DB.Exec(`
		INSERT INTO activity_log(client_id, type, ref_id, note, created_by, created_at)
		VALUES (?, 'wellbeing', ?, ?, ?, datetime('now'))
	`, clientID, lessonID, note, clientID)
	return err
}

// GetClientWellbeingHistory returns wellbeing history for a client.
func (s *Services) GetClientWellbeingHistory(clientID int64) ([]map[string]any, error) {
	rows, err := s.Store.DB.Query(`
		SELECT al.id, al.ref_id, al.note, al.created_at, l.date, l.time
		FROM activity_log al
		LEFT JOIN lessons l ON l.id = al.ref_id
		WHERE al.client_id = ? AND al.type = 'wellbeing'
		ORDER BY al.created_at DESC
	`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []map[string]any
	for rows.Next() {
		var id, refID int64
		var note, created, date, time string
		if err := rows.Scan(&id, &refID, &note, &created, &date, &time); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"id":        id,
			"lesson_id": refID,
			"note":      note,
			"created_at": created,
			"date":      date,
			"time":      time,
		})
	}
	return out, rows.Err()
}

// NewClientFAQ returns FAQ variations for new clients based on their query.
func (s *Services) NewClientFAQ(query string) string {
	query = strings.ToLower(query)
	switch {
	case strings.Contains(query, "цена") || strings.Contains(query, "стоимост") || strings.Contains(query, "абонемент"):
		return "💰 Стоимость занятий:\n• Разовое — 1500₽\n• Абонемент на 8 — 10000₽ (1250₽/зан)\n• Абонемент на 12 — 13200₽ (1100₽/зан)\n\nЕсть скидка 10% на первый абонемент!"
	case strings.Contains(query, "распис") || strings.Contains(query, "когда") || strings.Contains(query, "время"):
		return "📅 Расписание:\n• Понедельник 09:00, 19:00\n• Среда 09:00, 19:00\n• Пятница 09:00, 18:00\n• Суббота 10:00\n\nТочное расписание на неделю — кнопка «Моё расписание»"
	case strings.Contains(query, "медицин") || strings.Contains(query, "здоров") || strings.Contains(query, "ограничен"):
		return "🏥 Медицинские ограничения:\nПри наличии хронических заболеваний или травм — обязательно справка от врача. Тренер скорректирует нагрузку под ваши особенности."
	case strings.Contains(query, "что взять") || strings.Contains(query, "экипировк") || strings.Contains(query, "одежда"):
		return "👟 Что взять:\n• Удобная спортивная одежда\n• Кроссовки с чистым подошвой\n• Вода\n• Полотенце (по желанию)\n\nДуш и раздевалки на месте."
	default:
		return "👋 Добро пожаловать! Частые вопросы:\n• /price — цены и абонементы\n• /schedule — расписание\n• /medical — мед. ограничения\n• /gear — что взять\n\nИли просто напишите свой вопрос — отвечу!"
	}
}

// ---------- Notification helpers (used by handlers) ----------

// Notify enqueues a notification for a specific user.
func (s *Services) Notify(userID int64, typ string, payload map[string]any, sendAt time.Time) error {
	b, _ := json.Marshal(payload)
	return s.enqueue(userID, typ, string(b), sendAt)
}

// MessageCoaches sends a message from a client to all coaches.
func (s *Services) MessageCoaches(fromName, text string) error {
	return s.notifyCoaches("client_message", map[string]any{"from": fromName, "text": text})
}

// ClaimDue claims due notifications for the bot to dispatch.
func (s *Services) ClaimDue(limit int) ([]domain.Notification, error) {
	token := randToken()
	return s.Store.ClaimDue(time.Now(), limit, token)
}

// MarkResult marks a notification as sent or failed.
func (s *Services) MarkResult(id int64, ok bool) error {
	if ok {
		return s.Store.MarkNotificationSent(id)
	}
	return s.Store.MarkNotificationFailed(id)
}

// SendResult is the result of a batch notification send.
type SendResult struct {
	Total    int `json:"total"`
	Enqueued int `json:"enqueued"`
	Skipped  int `json:"skipped"`
	Errors   int `json:"errors"`
}

// ListRecipients returns eligible recipients for a coach notification.
func (s *Services) ListRecipients(coachID int64, filter string, groupID int64, clientIDs []int64) ([]domain.Recipient, error) {
	switch filter {
	case "today":
		return s.Store.ListCoachRecipients(coachID, time.Now().Format("2006-01-02"), 0)
	case "tomorrow":
		return s.Store.ListCoachRecipients(coachID, time.Now().AddDate(0, 0, 1).Format("2006-01-02"), 0)
	case "group":
		return s.Store.ListCoachRecipients(coachID, "", groupID)
	case "manual":
		return s.Store.ListRecipientsByIDs(clientIDs)
	default:
		return s.Store.ListCoachRecipients(coachID, "", 0)
	}
}

// SendCoachNotification sends a notification to all selected recipients.
func (s *Services) SendCoachNotification(coachID int64, filter string, groupID int64, clientIDs []int64, title, text string) (*SendResult, error) {
	recipients, err := s.ListRecipients(coachID, filter, groupID, clientIDs)
	if err != nil {
		return nil, err
	}
	result := &SendResult{Total: len(recipients)}
	payload, _ := json.Marshal(map[string]string{
		"title": title,
		"text":  text,
	})
	for _, r := range recipients {
		if r.UserID == nil {
			result.Skipped++
			continue
		}
		if err := s.enqueue(*r.UserID, "coach_broadcast", string(payload), time.Now()); err != nil {
			result.Errors++
			continue
		}
		result.Enqueued++
	}
	return result, nil
}

// ---------- Daily attendance ----------

func (s *Services) GetDateAttendance(date string) ([]domain.DateAttendanceClient, error) {
	clients, err := s.Store.ListClientsWithLessonsOnDate(date)
	if err != nil {
		return nil, err
	}
	existing, err := s.Store.GetDailyAttendance(date)
	if err != nil {
		return nil, err
	}
	for i := range clients {
		if p, ok := existing[clients[i].ClientID]; ok {
			v := p
			clients[i].Present = &v
		}
	}
	return clients, nil
}

func (s *Services) SaveDateAttendance(date string, entries []domain.DailyAttendance) error {
	return s.Store.SaveDailyAttendance(date, entries)
}

func (s *Services) GetSocialLinks(coachID int64) ([]domain.SocialLink, error) {
	if err := s.Store.SeedDefaultSocialLinks(coachID); err != nil {
		return nil, err
	}
	return s.Store.ListSocialLinks(coachID)
}

func (s *Services) GetSocialLinksMap(coachID int64) map[string]string {
	links, err := s.Store.ListSocialLinks(coachID)
	if err != nil {
		return map[string]string{}
	}
	m := make(map[string]string)
	for _, l := range links {
		if l.Enabled && l.URL != nil {
			m[l.Platform] = *l.URL
		}
	}
	return m
}

func (s *Services) SaveSocialLinks(coachID int64, links []domain.SocialLink) error {
	for _, l := range links {
		if err := s.Store.UpsertSocialLink(coachID, l.Platform, l.URL, l.Enabled); err != nil {
			return err
		}
	}
	return nil
}

// SocialMediaLinks returns configured social media links for the club/coach.
func (s *Services) SocialMediaLinks() map[string]string {
	return map[string]string{
		"instagram": "https://instagram.com/yourclub",
		"telegram":  "https://t.me/yourclub",
		"vk":        "https://vk.com/yourclub",
		"youtube":   "https://youtube.com/@yourclub",
	}
}

// ---------- Files ----------

func (s *Services) SaveFile(f domain.File) (int64, error) { return s.Store.InsertFile(f) }

// ---------- Reports ----------

type Report struct {
	ClientsTotal   int `json:"clients_total"`
	CoachesTotal   int `json:"coaches_total"`
	LessonsWeek    int `json:"lessons_week"`
}

func (s *Services) Report(from, to string) (*Report, error) {
	r := &Report{}
	clients, err := s.Store.ListClients()
	if err != nil {
		return nil, err
	}
	r.ClientsTotal = len(clients)
	coaches, err := s.Store.ListCoaches()
	if err != nil {
		return nil, err
	}
	r.CoachesTotal = len(coaches)
	lessons, err := s.Store.ListLessons(from, to, 0)
	if err != nil {
		return nil, err
	}
	r.LessonsWeek = len(lessons)
	return r, nil
}

// ---------- helpers ----------

func scanLessons(rows *sql.Rows) ([]domain.Lesson, error) {
	var out []domain.Lesson
	for rows.Next() {
		l := &domain.Lesson{}
		var coachID, groupID sql.NullInt64
		var loc, comment sql.NullString
		if err := rows.Scan(&l.ID, &l.Date, &l.Time, &coachID, &l.Duration, &l.Status, &loc, &comment, &groupID); err != nil {
			return nil, err
		}
		if coachID.Valid {
			v := coachID.Int64
			l.CoachID = &v
		}
		if groupID.Valid {
			v := groupID.Int64
			l.GroupID = &v
		}
		if loc.Valid {
			v := loc.String
			l.Location = &v
		}
		if comment.Valid {
			v := comment.String
			l.Comment = &v
		}
		out = append(out, *l)
	}
	return out, rows.Err()
}

func nullable(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableInt(n int) *int {
	if n == 0 {
		return nil
	}
	return &n
}

var ErrNotFound = errors.New("not found")

func (s *Services) MustClient(id int64) (*domain.Client, error) {
	c, err := s.Store.ClientByID(id)
	if err != nil {
		return nil, err
	}
	if c == nil {
		return nil, fmt.Errorf("%w: client %d", ErrNotFound, id)
	}
	return c, nil
}

func randToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// ---------- Coach Subscriptions ----------

const TrialDays = 7

func (s *Services) GetOrCreateCoachSubscription(coachID int64) (*domain.CoachSubscription, error) {
	sub, err := s.Store.CoachSubscriptionByCoachID(coachID)
	if err != nil {
		return nil, err
	}
	if sub == nil {
		return s.Store.CreateCoachSubscription(coachID, TrialDays)
	}
	return sub, nil
}

func (s *Services) GetCoachSubscription(coachID int64) (*domain.CoachSubscription, error) {
	return s.Store.CoachSubscriptionByCoachID(coachID)
}

func (s *Services) IsCoachSubscriptionActive(coachID int64) bool {
	sub, err := s.Store.CoachSubscriptionByCoachID(coachID)
	if err != nil || sub == nil {
		return false
	}
	if sub.Status == domain.SubActive {
		if sub.PaidUntil != nil {
			paidUntil, err := time.Parse("2006-01-02 15:04:05", *sub.PaidUntil)
			if err == nil && paidUntil.Before(time.Now()) {
				return false
			}
		}
		return true
	}
	if sub.Status == domain.SubTrial {
		if sub.TrialEnd != nil {
			trialEnd, err := time.Parse("2006-01-02 15:04:05", *sub.TrialEnd)
			if err == nil && trialEnd.Before(time.Now()) {
				_ = s.Store.UpdateCoachSubscriptionStatus(coachID, domain.SubExpired)
				return false
			}
		}
		return true
	}
	return false
}

// ---------- Parent features ----------

func (s *Services) FindChildByBirthDate(fullName, birthDate string) (*domain.Client, error) {
	return s.Store.ClientByBirthDate(fullName, birthDate)
}

func (s *Services) LinkParentToChild(parentUserID int64, childClientID int64) error {
	return s.Store.LinkParentToChild(parentUserID, childClientID)
}

func (s *Services) CreateInviteCode(clientID, createdBy int64) (string, error) {
	expiresAt := time.Now().Add(72 * time.Hour).Format("2006-01-02 15:04:05")
	return s.Store.CreateInviteCode(clientID, createdBy, expiresAt)
}

func (s *Services) UseInviteCode(code string) (int64, error) {
	return s.Store.UseInviteCode(code)
}

func (s *Services) GetParentNotifPrefs(parentUserID int64) ([]domain.ParentNotifPref, error) {
	return s.Store.GetParentNotifPrefs(parentUserID)
}

func (s *Services) SaveParentNotifPrefs(pref domain.ParentNotifPref) error {
	return s.Store.UpsertParentNotifPref(pref)
}

// GetChildLessonStatus returns lesson status info for a child for the parent dashboard.
func (s *Services) GetChildLessonStatus(childID int64) (*domain.ChildLessonStatus, error) {
	now := time.Now()
	today := now.Format("2006-01-02")
	currentTime := now.Format("15:04")

	entries, err := s.Store.ListChildLessonEntries(childID, today, today)
	if err != nil {
		return nil, err
	}

	client, err := s.Store.ClientByID(childID)
	if err != nil || client == nil {
		return nil, fmt.Errorf("%w: child %d", ErrNotFound, childID)
	}

	status := &domain.ChildLessonStatus{
		ClientID:       childID,
		FullName:       client.FullName,
		HasLessonToday: len(entries) > 0,
	}

	for _, e := range entries {
		if status.Date == "" || e.Date > status.Date || (e.Date == status.Date && e.Time > status.Time) {
			status.Date = e.Date
			status.Time = e.Time
			status.Duration = e.Duration
			status.Status = string(e.Status)
		}

		if e.Date == today {
			status.HasLessonToday = true
			startTime, err := time.Parse("15:04", e.Time)
			if err != nil {
				continue
			}
			currentParsed, err := time.Parse("15:04", currentTime)
			if err != nil {
				continue
			}
			startMinutes := startTime.Hour()*60 + startTime.Minute()
			currentMinutes := currentParsed.Hour()*60 + currentParsed.Minute()
			endMinutes := startMinutes + e.Duration

			if currentMinutes >= startMinutes && currentMinutes <= endMinutes {
				left := endMinutes - currentMinutes
				status.IsOngoing = true
				status.MinutesLeft = &left
			} else if currentMinutes < startMinutes {
				until := startMinutes - currentMinutes
				status.MinutesUntil = &until
			}
		}
	}

	status.IsToday = status.Date == today
	return status, nil
}

// ---------- Lesson status for parent dashboard ----------

func (s *Services) GetChildrenLessonStatuses(parentUserID int64) ([]domain.ChildLessonStatus, error) {
	children, err := s.ChildrenOfParent(parentUserID)
	if err != nil {
		return nil, err
	}
	var out []domain.ChildLessonStatus
	for _, ch := range children {
		status, err := s.GetChildLessonStatus(ch.ID)
		if err != nil {
			continue
		}
		out = append(out, *status)
	}
	return out, nil
}

// ---------- Groups ----------

func (s *Services) ListGroups() ([]domain.Group, error) {
	return s.Store.ListGroups()
}

func (s *Services) GetGroup(id int64) (*domain.Group, error) {
	return s.Store.GetGroup(id)
}

func (s *Services) CreateGroup(g domain.Group) (int64, error) {
	return s.Store.CreateGroup(g)
}

func (s *Services) UpdateGroup(g domain.Group) error {
	return s.Store.UpdateGroup(g)
}

func (s *Services) DeleteGroup(id int64) error {
	return s.Store.DeleteGroup(id)
}

func (s *Services) AddClientToGroup(groupID, clientID int64, role string) error {
	return s.Store.AddClientToGroup(groupID, clientID, role)
}

func (s *Services) RemoveClientFromGroup(groupID, clientID int64) error {
	return s.Store.RemoveClientFromGroup(groupID, clientID)
}

func (s *Services) GetGroupClients(groupID int64) ([]domain.GroupMember, error) {
	return s.Store.GetGroupClients(groupID)
}

func (s *Services) GetClientGroups(clientID int64) ([]domain.Group, error) {
	return s.Store.GetClientGroups(clientID)
}
