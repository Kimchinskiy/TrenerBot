package store

import (
	"database/sql"
	"log/slog"
	"time"

	"trenerbot/internal/domain"
)

type Store struct {
	DB *sql.DB
}

func New(db *sql.DB) *Store { return &Store{DB: db} }

// ---------- Users ----------

func (s *Store) UserByTelegram(tgID string) (*domain.User, error) {
	u := &domain.User{}
	var created string
	err := s.DB.QueryRow(`SELECT id, telegram_id, role, created_at FROM users WHERE telegram_id = ?`, tgID).
		Scan(&u.ID, &u.TelegramID, &u.Role, &created)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	return u, nil
}

func (s *Store) UserByID(id int64) (*domain.User, error) {
	u := &domain.User{}
	var created string
	err := s.DB.QueryRow(`SELECT id, telegram_id, role, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.TelegramID, &u.Role, &created)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	return u, nil
}

func (s *Store) CreateUser(tgID *string, role domain.Role) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO users(telegram_id, role) VALUES (?, ?)`, tgID, string(role))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ---------- Clients ----------

func (s *Store) ClientByUserID(userID int64) (*domain.Client, error) {
	c := &domain.Client{}
	var birth, phone, tg, wa, email, med, note, reg, src, botAccess, subEnds sql.NullString
	var age, parentID, secondParentID sql.NullInt64
	var photo sql.NullString
	err := s.DB.QueryRow(`SELECT id, user_id, full_name, photo, birth_date, age, phone, telegram, whatsapp,
		parent_id, second_parent_id, email, medical_limits, note, status, registered_at, source, bot_access, subscription_ends_at
		FROM clients WHERE user_id = ?`, userID).
		Scan(&c.ID, &c.UserID, &c.FullName, &photo, &birth, &age, &phone, &tg, &wa,
			&parentID, &secondParentID, &email, &med, &note, &c.Status, &reg, &src, &botAccess, &subEnds)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	assignNulls(c, photo, birth, age, phone, tg, wa, parentID, secondParentID, email, med, note, reg, src, botAccess, subEnds)
	return c, nil
}

func (s *Store) ClientByID(id int64) (*domain.Client, error) {
	c := &domain.Client{}
	var birth, phone, tg, wa, email, med, note, reg, src, botAccess, subEnds sql.NullString
	var age, parentID, secondParentID sql.NullInt64
	var photo sql.NullString
	err := s.DB.QueryRow(`SELECT id, user_id, full_name, photo, birth_date, age, phone, telegram, whatsapp,
		parent_id, second_parent_id, email, medical_limits, note, status, registered_at, source, bot_access, subscription_ends_at
		FROM clients WHERE id = ?`, id).
		Scan(&c.ID, &c.UserID, &c.FullName, &photo, &birth, &age, &phone, &tg, &wa,
			&parentID, &secondParentID, &email, &med, &note, &c.Status, &reg, &src, &botAccess, &subEnds)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	assignNulls(c, photo, birth, age, phone, tg, wa, parentID, secondParentID, email, med, note, reg, src, botAccess, subEnds)
	return c, nil
}

func assignNulls(c *domain.Client, photo, birth sql.NullString, age sql.NullInt64, phone, tg, wa sql.NullString,
	parentID, secondParentID sql.NullInt64, email, med, note, reg, src, botAccess, subEnds sql.NullString) {
	if photo.Valid {
		v := photo.String
		c.Photo = &v
	}
	if birth.Valid {
		v := birth.String
		c.BirthDate = &v
	}
	if age.Valid {
		v := int(age.Int64)
		c.Age = &v
	}
	if phone.Valid {
		v := phone.String
		c.Phone = &v
	}
	if tg.Valid {
		v := tg.String
		c.Telegram = &v
	}
	if wa.Valid {
		v := wa.String
		c.WhatsApp = &v
	}
	if parentID.Valid {
		v := parentID.Int64
		c.ParentID = &v
	}
	if secondParentID.Valid {
		v := secondParentID.Int64
		c.SecondParentID = &v
	}
	if email.Valid {
		v := email.String
		c.Email = &v
	}
	if med.Valid {
		v := med.String
		c.MedicalLimits = &v
	}
	if note.Valid {
		v := note.String
		c.Note = &v
	}
	if reg.Valid {
		v := reg.String
		c.RegisteredAt = &v
	}
	if src.Valid {
		v := src.String
		c.Source = &v
	}
	if botAccess.Valid {
		c.BotAccess = botAccess.String == "1"
	}
	if subEnds.Valid {
		v := subEnds.String
		c.SubscriptionEndsAt = &v
	}
}

func (s *Store) CreateClient(c domain.Client) (int64, error) {
	botAccess := 0
	if c.BotAccess {
		botAccess = 1
	}
	var subEnds *string
	if c.SubscriptionEndsAt != nil {
		subEnds = c.SubscriptionEndsAt
	}
	res, err := s.DB.Exec(`INSERT INTO clients(user_id, full_name, photo, birth_date, age, phone, telegram,
		whatsapp, parent_id, second_parent_id, email, medical_limits, note, status, registered_at, source, bot_access, subscription_ends_at)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		c.UserID, c.FullName, c.Photo, c.BirthDate, c.Age, c.Phone, c.Telegram, c.WhatsApp,
		c.ParentID, c.SecondParentID, c.Email, c.MedicalLimits, c.Note, c.Status, c.RegisteredAt, c.Source, botAccess, subEnds)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateClient(c domain.Client) error {
	botAccess := 0
	if c.BotAccess {
		botAccess = 1
	}
	var subEnds *string
	if c.SubscriptionEndsAt != nil {
		subEnds = c.SubscriptionEndsAt
	}
	slog.Debug("UpdateClient", "id", c.ID, "botAccess", botAccess, "subEnds", subEnds)
	_, err := s.DB.Exec(`UPDATE clients SET full_name=?, photo=?, birth_date=?, age=?, phone=?, telegram=?,
		whatsapp=?, parent_id=?, second_parent_id=?, email=?, medical_limits=?, note=?, status=?, source=?, bot_access=?, subscription_ends_at=?
		WHERE id=?`,
		c.FullName, c.Photo, c.BirthDate, c.Age, c.Phone, c.Telegram, c.WhatsApp,
		c.ParentID, c.SecondParentID, c.Email, c.MedicalLimits, c.Note, c.Status, c.Source, botAccess, subEnds, c.ID)
	return err
}

func (s *Store) ListClients() ([]domain.Client, error) {
	rows, err := s.DB.Query(`SELECT id, user_id, full_name, status, bot_access, subscription_ends_at FROM clients ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Client
	for rows.Next() {
		var c domain.Client
		var uid sql.NullInt64
		var botAccess, subEnds sql.NullString
		if err := rows.Scan(&c.ID, &uid, &c.FullName, &c.Status, &botAccess, &subEnds); err != nil {
			return nil, err
		}
		if uid.Valid {
			v := uid.Int64
			c.UserID = &v
		}
		if botAccess.Valid {
			c.BotAccess = botAccess.String == "1"
		}
		if subEnds.Valid {
			c.SubscriptionEndsAt = &subEnds.String
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Children of a parent user (by parent_id linking to users.id).
func (s *Store) ChildrenOfParent(parentUserID int64) ([]domain.Client, error) {
	rows, err := s.DB.Query(`SELECT id, full_name, status FROM clients WHERE parent_id = ? OR second_parent_id = ? ORDER BY full_name`, parentUserID, parentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Client
	for rows.Next() {
		var c domain.Client
		if err := rows.Scan(&c.ID, &c.FullName, &c.Status); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// ---------- Coaches ----------

func (s *Store) CoachByUserID(userID int64) (*domain.Coach, error) {
	co := &domain.Coach{}
	var photo, contacts, position, sport, schedule, groupIDs sql.NullString
	err := s.DB.QueryRow(`SELECT id, user_id, full_name, photo, contacts, position, sport, schedule, group_ids
		FROM coaches WHERE user_id = ?`, userID).
		Scan(&co.ID, &co.UserID, &co.FullName, &photo, &contacts, &position, &sport, &schedule, &groupIDs)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	assignCoachNulls(co, photo, contacts, position, sport, schedule, groupIDs)
	return co, nil
}

func assignCoachNulls(co *domain.Coach, photo, contacts, position, sport, schedule, groupIDs sql.NullString) {
	if photo.Valid {
		v := photo.String
		co.Photo = &v
	}
	if contacts.Valid {
		v := contacts.String
		co.Contacts = &v
	}
	if position.Valid {
		v := position.String
		co.Position = &v
	}
	if sport.Valid {
		v := sport.String
		co.Sport = &v
	}
	if schedule.Valid {
		v := schedule.String
		co.Schedule = &v
	}
	if groupIDs.Valid {
		v := groupIDs.String
		co.GroupIDs = &v
	}
}

func (s *Store) CreateCoach(co domain.Coach) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO coaches(user_id, full_name, photo, contacts, position, sport, schedule, group_ids)
		VALUES (?,?,?,?,?,?,?,?)`,
		co.UserID, co.FullName, co.Photo, co.Contacts, co.Position, co.Sport, co.Schedule, co.GroupIDs)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListCoaches() ([]domain.Coach, error) {
	rows, err := s.DB.Query(`SELECT id, user_id, full_name FROM coaches ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Coach
	for rows.Next() {
		var co domain.Coach
		var uid sql.NullInt64
		if err := rows.Scan(&co.ID, &uid, &co.FullName); err != nil {
			return nil, err
		}
		if uid.Valid {
			v := uid.Int64
			co.UserID = &v
		}
		out = append(out, co)
	}
	return out, rows.Err()
}

// ---------- Lessons ----------

func (s *Store) CreateLesson(l domain.Lesson) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO lessons(date, time, coach_id, duration, status, location, comment, group_id)
		VALUES (?,?,?,?,?,?,?,?)`,
		l.Date, l.Time, l.CoachID, l.Duration, string(l.Status), l.Location, l.Comment, l.GroupID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) LessonByID(id int64) (*domain.Lesson, error) {
	l := &domain.Lesson{}
	var coachID, groupID sql.NullInt64
	var loc, comment sql.NullString
	err := s.DB.QueryRow(`SELECT id, date, time, coach_id, duration, status, location, comment, group_id
		FROM lessons WHERE id = ?`, id).
		Scan(&l.ID, &l.Date, &l.Time, &coachID, &l.Duration, &l.Status, &loc, &comment, &groupID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
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
	return l, nil
}

// ListWeek returns lessons for a coach within [from, to]; if coachID<=0 all coaches.
func (s *Store) ListLessons(from, to string, coachID int64) ([]domain.Lesson, error) {
	q := `SELECT id, date, time, coach_id, duration, status, location, comment, group_id FROM lessons WHERE date >= ? AND date <= ?`
	args := []interface{}{from, to}
	if coachID > 0 {
		q += ` AND coach_id = ?`
		args = append(args, coachID)
	}
	q += ` ORDER BY date, time`
	rows, err := s.DB.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
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

func (s *Store) UpdateLessonStatus(id int64, status domain.LessonStatus) error {
	_, err := s.DB.Exec(`UPDATE lessons SET status = ? WHERE id = ?`, string(status), id)
	return err
}

// ---------- Attendance ----------

func (s *Store) SetAttendance(lessonID, clientID int64, present bool, markedBy *int64) error {
	now := time.Now().Format("2006-01-02 15:04:05")
	_, err := s.DB.Exec(`INSERT INTO attendance(lesson_id, client_id, present, marked_at, marked_by)
		VALUES (?,?,?,?,?)
		ON CONFLICT(lesson_id, client_id) DO UPDATE SET present=excluded.present, marked_at=excluded.marked_at, marked_by=excluded.marked_by`,
		lessonID, clientID, boolToInt(present), now, markedBy)
	return err
}

func (s *Store) ListAttendanceByLesson(lessonID int64) ([]domain.Attendance, error) {
	rows, err := s.DB.Query(`SELECT lesson_id, client_id, present, marked_at, marked_by FROM attendance WHERE lesson_id = ?`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Attendance
	for rows.Next() {
		a := domain.Attendance{}
		var present int
		var markedAt sql.NullString
		var markedBy sql.NullInt64
		if err := rows.Scan(&a.LessonID, &a.ClientID, &present, &markedAt, &markedBy); err != nil {
			return nil, err
		}
		a.Present = present == 1
		if markedAt.Valid {
			v := markedAt.String
			a.MarkedAt = &v
		}
		if markedBy.Valid {
			v := markedBy.Int64
			a.MarkedBy = &v
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (s *Store) ListAttendanceByClient(clientID int64) ([]domain.Attendance, error) {
	rows, err := s.DB.Query(`SELECT lesson_id, client_id, present, marked_at, marked_by FROM attendance WHERE client_id = ?`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Attendance
	for rows.Next() {
		a := domain.Attendance{}
		var present int
		var markedAt sql.NullString
		var markedBy sql.NullInt64
		if err := rows.Scan(&a.LessonID, &a.ClientID, &present, &markedAt, &markedBy); err != nil {
			return nil, err
		}
		a.Present = present == 1
		if markedAt.Valid {
			v := markedAt.String
			a.MarkedAt = &v
		}
		if markedBy.Valid {
			v := markedBy.Int64
			a.MarkedBy = &v
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// ---------- Notifications (outbox with DB-level claim) ----------

func (s *Store) InsertNotification(n domain.Notification) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO notifications(channel, recipient_user_id, type, payload, send_at, status)
		VALUES (?,?,?,?,?, 'pending')`,
		n.Channel, n.RecipientUserID, n.Type, n.Payload, n.SendAt.Format("2006-01-02 15:04:05"))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ClaimDue atomically marks up to `limit` pending due notifications as 'claimed' using a unique
// token, then returns them. Only one instance can claim a given row -> no duplicate sends (ТЗ §15).
func (s *Store) ClaimDue(now time.Time, limit int, token string) ([]domain.Notification, error) {
	tx, err := s.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`UPDATE notifications SET status='claimed', claim_token=?
		WHERE id IN (SELECT id FROM notifications WHERE status='pending' AND send_at <= ? LIMIT ?)`,
		token, now.Format("2006-01-02 15:04:05"), limit)
	if err != nil {
		return nil, err
	}
	rows, err := tx.Query(`SELECT n.id, n.channel, n.recipient_user_id, n.type, n.payload, u.telegram_id
		FROM notifications n JOIN users u ON u.id = n.recipient_user_id
		WHERE n.status='claimed' AND n.claim_token=?`, token)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Notification
	for rows.Next() {
		n := domain.Notification{}
		var tgID sql.NullString
		if err := rows.Scan(&n.ID, &n.Channel, &n.RecipientUserID, &n.Type, &n.Payload, &tgID); err != nil {
			return nil, err
		}
		if tgID.Valid {
			v := tgID.String
			n.TelegramID = &v
		}
		out = append(out, n)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) MarkNotificationSent(id int64) error {
	_, err := s.DB.Exec(`UPDATE notifications SET status='sent', sent_at=datetime('now') WHERE id=?`, id)
	return err
}

func (s *Store) MarkNotificationFailed(id int64) error {
	_, err := s.DB.Exec(`UPDATE notifications SET status='failed' WHERE id=?`, id)
	return err
}

// LessonParticipants returns client IDs registered for a lesson.
func (s *Store) LessonParticipants(lessonID int64) ([]int64, error) {
	rows, err := s.DB.Query(`SELECT client_id FROM attendance WHERE lesson_id = ?`, lessonID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var cid int64
		if err := rows.Scan(&cid); err != nil {
			return nil, err
		}
		out = append(out, cid)
	}
	return out, rows.Err()
}

// NotificationExists reports whether an equivalent notification is already queued/delivered.
func (s *Store) NotificationExists(typ string, userID int64, payload string) (bool, error) {
	var n int
	err := s.DB.QueryRow(`SELECT 1 FROM notifications WHERE type = ? AND recipient_user_id = ? AND payload = ?
		AND status IN ('pending','claimed','sent') LIMIT 1`, typ, userID, payload).Scan(&n)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// ---------- Files ----------

func (s *Store) InsertFile(f domain.File) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO files(owner_type, owner_id, path, kind) VALUES (?,?,?,?)`,
		f.OwnerType, f.OwnerID, f.Path, f.Kind)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ---------- Activity log ----------

func (s *Store) LogActivity(clientID int64, typ string, refID int64, note string, createdBy *int64) error {
	_, err := s.DB.Exec(`INSERT INTO activity_log(client_id, type, ref_id, note, created_by) VALUES (?,?,?,?,?)`,
		clientID, typ, refID, note, createdBy)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
