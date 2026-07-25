package store

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"trenerbot/internal/domain"
)

type Store struct {
	DB *sql.DB
}

func New(db *sql.DB) *Store { return &Store{DB: db} }

// ---------- Users ----------

const userColumns = `id, phone, password_hash, telegram_id, max_id, first_name, last_name, avatar_url, role, created_at, updated_at`

func scanUser(row interface{ Scan(...any) error }) (*domain.User, error) {
	u := &domain.User{}
	var phone, pass, tgID, maxID, first, last, avatar sql.NullString
	var created string
	var updated sql.NullString
	err := row.Scan(&u.ID, &phone, &pass, &tgID, &maxID, &first, &last, &avatar, &u.Role, &created, &updated)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	u.Phone = nullStr(phone)
	u.PasswordHash = nullStr(pass)
	u.TelegramID = nullStr(tgID)
	u.MaxID = nullStr(maxID)
	u.FirstName = nullStr(first)
	u.LastName = nullStr(last)
	u.AvatarURL = nullStr(avatar)
	u.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	if updated.Valid {
		if t, err := time.Parse("2006-01-02 15:04:05", updated.String); err == nil {
			u.UpdatedAt = &t
		}
	}
	return u, nil
}

func nullStr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func nullInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
}

func nullInt64(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func nullStrPtr(v *string) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt64Ptr(v *int64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func nullIntPtr(v *int) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func nullFloat64Ptr(v *float64) interface{} {
	if v == nil {
		return nil
	}
	return *v
}

func (s *Store) UserByTelegram(tgID string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE telegram_id = ?`, tgID))
}

func (s *Store) UserByID(id int64) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE id = ?`, id))
}

func (s *Store) UserByPhone(phone string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE phone = ?`, phone))
}

func (s *Store) UserByMaxID(maxID string) (*domain.User, error) {
	return scanUser(s.DB.QueryRow(`SELECT `+userColumns+` FROM users WHERE max_id = ?`, maxID))
}

// CreateUser inserts a Telegram-only account (kept for the bot self-registration flow).
func (s *Store) CreateUser(tgID *string, role domain.Role) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO users(telegram_id, role, updated_at) VALUES (?, ?, datetime('now'))`, tgID, string(role))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// InsertUser creates a full user account (website registration / provider login).
func (s *Store) InsertUser(u domain.User) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO users(phone, password_hash, telegram_id, max_id, first_name, last_name, avatar_url, role, updated_at)
		VALUES (?,?,?,?,?,?,?,?, datetime('now'))`,
		u.Phone, u.PasswordHash, u.TelegramID, u.MaxID, u.FirstName, u.LastName, u.AvatarURL, string(u.Role))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateUserProfile updates the mutable profile/identity fields of a user.
func (s *Store) UpdateUserProfile(u domain.User) error {
	_, err := s.DB.Exec(`UPDATE users SET phone=?, password_hash=?, telegram_id=?, max_id=?,
		first_name=?, last_name=?, avatar_url=?, role=?, updated_at=datetime('now') WHERE id=?`,
		u.Phone, u.PasswordHash, u.TelegramID, u.MaxID, u.FirstName, u.LastName, u.AvatarURL, string(u.Role), u.ID)
	return err
}

// UpdateUserRole updates only the role field for a user.
func (s *Store) UpdateUserRole(id int64, role domain.Role) error {
	_, err := s.DB.Exec(`UPDATE users SET role=?, updated_at=datetime('now') WHERE id=?`, string(role), id)
	return err
}

// LinkTelegram attaches a telegram_id to an existing user (login-method linking).
func (s *Store) LinkTelegram(userID int64, tgID string) error {
	_, err := s.DB.Exec(`UPDATE users SET telegram_id=?, updated_at=datetime('now') WHERE id=?`, tgID, userID)
	return err
}

// LinkMax attaches a max_id to an existing user.
func (s *Store) LinkMax(userID int64, maxID string) error {
	_, err := s.DB.Exec(`UPDATE users SET max_id=?, updated_at=datetime('now') WHERE id=?`, maxID, userID)
	return err
}

// ---------- Refresh tokens ----------

func (s *Store) InsertRefreshToken(userID int64, tokenHash string, expiresAt time.Time) error {
	_, err := s.DB.Exec(`INSERT INTO refresh_tokens(user_id, token_hash, expires_at) VALUES (?,?,?)`,
		userID, tokenHash, expiresAt.Format("2006-01-02 15:04:05"))
	return err
}

// RefreshTokenUser returns the owning user id if the token hash is valid (not revoked, not expired).
func (s *Store) RefreshTokenUser(tokenHash string) (int64, error) {
	var userID int64
	var expires string
	var revoked int
	err := s.DB.QueryRow(`SELECT user_id, expires_at, revoked FROM refresh_tokens WHERE token_hash = ?`, tokenHash).
		Scan(&userID, &expires, &revoked)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	if revoked == 1 {
		return 0, nil
	}
	if t, err := time.Parse("2006-01-02 15:04:05", expires); err == nil && t.Before(time.Now()) {
		return 0, nil
	}
	return userID, nil
}

func (s *Store) RevokeRefreshToken(tokenHash string) error {
	_, err := s.DB.Exec(`UPDATE refresh_tokens SET revoked=1 WHERE token_hash=?`, tokenHash)
	return err
}

func (s *Store) RevokeAllRefreshTokens(userID int64) error {
	_, err := s.DB.Exec(`UPDATE refresh_tokens SET revoked=1 WHERE user_id=?`, userID)
	return err
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
	rows, err := s.DB.Query(`SELECT c.id, c.user_id, c.full_name, c.status, c.bot_access, c.subscription_ends_at FROM clients c WHERE c.full_name != '' AND c.full_name IS NOT NULL AND (c.user_id IS NULL OR c.user_id NOT IN (SELECT id FROM users WHERE role IN ('admin', 'coach'))) AND c.id IN (SELECT MIN(c2.id) FROM clients c2 WHERE c2.full_name != '' AND c2.full_name IS NOT NULL GROUP BY CASE WHEN c2.user_id IS NULL THEN c2.full_name ELSE CAST(c2.user_id AS TEXT) END) ORDER BY c.full_name`)
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

func (s *Store) CoachByID(id int64) (*domain.Coach, error) {
	co := &domain.Coach{}
	var photo, contacts, position, sport, schedule, groupIDs sql.NullString
	err := s.DB.QueryRow(`SELECT id, user_id, full_name, photo, contacts, position, sport, schedule, group_ids
		FROM coaches WHERE id = ?`, id).
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

// ---------- Lesson entries (new schedule model) ----------

func (s *Store) InsertLessonEntry(e domain.LessonEntry) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO lesson_entries(date, time, client_id, coach_id, group_id, duration, status, comment)
		VALUES (?,?,?,?,?,?,?,?)`, e.Date, e.Time, e.ClientID, e.CoachID, e.GroupID, e.Duration, string(e.Status), e.Comment)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// ListScheduleEntries returns entries in [from, to] joined with client names,
// ordered by date then time. If coachID > 0, filters by coach; if clientID > 0,
// filters by client; otherwise returns all.
func (s *Store) ListScheduleEntries(from, to string, coachID, clientID int64) ([]domain.ScheduleEntry, error) {
	rows, err := s.DB.Query(`SELECT le.id, le.date, le.time, le.client_id, c.full_name, le.coach_id, le.duration, le.status, le.group_id, g.name
		FROM lesson_entries le
		JOIN clients c ON c.id = le.client_id
		LEFT JOIN groups g ON g.id = le.group_id
		WHERE le.date >= ? AND le.date <= ?
		AND (? <= 0 OR le.coach_id = ?)
		AND (? <= 0 OR le.client_id = ?)
		ORDER BY le.date, le.time`, from, to, coachID, coachID, clientID, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ScheduleEntry
	for rows.Next() {
		var e domain.ScheduleEntry
		var coachIDVal, groupIDVal sql.NullInt64
		var groupName sql.NullString
		if err := rows.Scan(&e.ID, &e.Date, &e.Time, &e.ClientID, &e.ClientName, &coachIDVal, &e.Duration, &e.Status, &groupIDVal, &groupName); err != nil {
			return nil, err
		}
		if coachIDVal.Valid {
			v := coachIDVal.Int64
			e.CoachID = &v
		}
		if groupIDVal.Valid {
			v := groupIDVal.Int64
			e.GroupID = &v
		}
		if groupName.Valid {
			v := groupName.String
			e.GroupName = &v
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ReminderEntry is the data needed for sending lesson reminders.
type ReminderEntry struct {
	ClientID   int64
	UserID     int64
	LessonID   int64
	Date       string
	Time       string
	Location   *string
	CoachID    *int64
}

// ListLessonEntriesForReminders returns planned lesson entries within [from, to]
// joined with the client's user_id for notification delivery.
func (s *Store) ListLessonEntriesForReminders(from, to string) ([]ReminderEntry, error) {
	rows, err := s.DB.Query(`SELECT le.id, le.date, le.time, le.client_id, c.user_id, le.coach_id
		FROM lesson_entries le
		JOIN clients c ON c.id = le.client_id
		WHERE le.date >= ? AND le.date <= ? AND le.status = 'planned'
		ORDER BY le.date, le.time`, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReminderEntry
	for rows.Next() {
		var e ReminderEntry
		var uid, coachID sql.NullInt64
		if err := rows.Scan(&e.LessonID, &e.Date, &e.Time, &e.ClientID, &uid, &coachID); err != nil {
			return nil, err
		}
		if uid.Valid {
			e.UserID = uid.Int64
		}
		if coachID.Valid {
			e.CoachID = &coachID.Int64
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ---------- Coach notification recipients ----------

// ListCoachRecipients returns distinct clients who have lesson entries
// with the given coach, optionally filtered by date or group.
func (s *Store) ListCoachRecipients(coachID int64, date string, groupID int64) ([]domain.Recipient, error) {
	var rows *sql.Rows
	var err error
	if date != "" {
		if coachID > 0 {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM clients c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.coach_id = ? AND le.date = ?`, coachID, date)
		} else {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM clients c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.date = ?`, date)
		}
	} else if groupID > 0 {
		if coachID > 0 {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM clients c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.coach_id = ? AND le.group_id = ?`, coachID, groupID)
		} else {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM clients c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.group_id = ?`, groupID)
		}
	} else {
		if coachID > 0 {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM clients c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.coach_id = ?`, coachID)
		} else {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM clients c
				JOIN lesson_entries le ON le.client_id = c.id`)
		}
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Recipient
	for rows.Next() {
		var r domain.Recipient
		if err := rows.Scan(&r.ClientID, &r.FullName, &r.UserID); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---------- Daily attendance ----------

func (s *Store) ListClientsWithLessonsOnDate(date string) ([]domain.DateAttendanceClient, error) {
	rows, err := s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.photo, le.time
		FROM clients c
		JOIN lesson_entries le ON le.client_id = c.id
		WHERE le.date = ? ORDER BY le.time, c.full_name`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.DateAttendanceClient
	for rows.Next() {
		var r domain.DateAttendanceClient
		var avatar sql.NullString
		if err := rows.Scan(&r.ClientID, &r.FullName, &avatar, &r.Time); err != nil {
			return nil, err
		}
		if avatar.Valid {
			r.Photo = &avatar.String
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *Store) GetDailyAttendance(date string) (map[int64]bool, error) {
	rows, err := s.DB.Query(`SELECT client_id, present FROM daily_attendance WHERE date = ?`, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[int64]bool)
	for rows.Next() {
		var cid int64
		var present int
		if err := rows.Scan(&cid, &present); err != nil {
			return nil, err
		}
		m[cid] = present == 1
	}
	return m, rows.Err()
}

func (s *Store) SaveDailyAttendance(date string, entries []domain.DailyAttendance) error {
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	for _, e := range entries {
		_, err := tx.Exec(`INSERT INTO daily_attendance(date, client_id, present, marked_by, updated_at) VALUES (?,?,?,?,datetime('now'))
			ON CONFLICT(date, client_id) DO UPDATE SET present=excluded.present, marked_by=excluded.marked_by, updated_at=excluded.updated_at`,
			e.Date, e.ClientID, boolToInt(e.Present), e.MarkedBy)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// ListRecipientsByIDs returns clients matching the given IDs.
func (s *Store) ListRecipientsByIDs(ids []int64) ([]domain.Recipient, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	placeholders := strings.Repeat(",?", len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		args[i] = id
	}
	rows, err := s.DB.Query(`SELECT id, full_name, user_id FROM clients WHERE id IN(`+placeholders[1:]+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Recipient
	for rows.Next() {
		var r domain.Recipient
		if err := rows.Scan(&r.ClientID, &r.FullName, &r.UserID); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
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

// ---------- Coach Subscriptions ----------

func (s *Store) CreateCoachSubscription(coachID int64, trialDays int) (*domain.CoachSubscription, error) {
	trialEnd := time.Now().AddDate(0, 0, trialDays).Format("2006-01-02 15:04:05")
	_, err := s.DB.Exec(`INSERT INTO coach_subscriptions(coach_id, status, trial_end) VALUES (?, 'trial', ?)`, coachID, trialEnd)
	if err != nil {
		return nil, err
	}
	return s.CoachSubscriptionByCoachID(coachID)
}

func (s *Store) CoachSubscriptionByCoachID(coachID int64) (*domain.CoachSubscription, error) {
	sub := &domain.CoachSubscription{}
	var trialEnd, paidUntil, updated sql.NullString
	err := s.DB.QueryRow(`SELECT id, coach_id, status, trial_start, trial_end, paid_until, created_at, updated_at
		FROM coach_subscriptions WHERE coach_id = ?`, coachID).
		Scan(&sub.ID, &sub.CoachID, &sub.Status, &sub.TrialStart, &trialEnd, &paidUntil, &sub.CreatedAt, &updated)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if trialEnd.Valid {
		sub.TrialEnd = &trialEnd.String
	}
	if paidUntil.Valid {
		sub.PaidUntil = &paidUntil.String
	}
	if updated.Valid {
		sub.UpdatedAt = &updated.String
	}
	return sub, nil
}

func (s *Store) UpdateCoachSubscriptionStatus(coachID int64, status domain.SubscriptionStatus) error {
	_, err := s.DB.Exec(`UPDATE coach_subscriptions SET status=?, updated_at=datetime('now') WHERE coach_id=?`, string(status), coachID)
	return err
}

func (s *Store) ExtendCoachSubscription(coachID int64, days int) error {
	_, err := s.DB.Exec(`UPDATE coach_subscriptions SET paid_until=datetime('now', ?), status='active', updated_at=datetime('now') WHERE coach_id=?`,
		"+"+strconv.Itoa(days)+" days", coachID)
	return err
}

// ---------- Client Subscriptions ----------

func (s *Store) ClientSubscriptions(clientID int64) ([]domain.ClientSubscription, error) {
	rows, err := s.DB.Query(`SELECT id, client_id, type, price, bought_at, ends_at, lessons_left, freeze, created_at FROM subscriptions WHERE client_id = ? ORDER BY created_at DESC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ClientSubscription
	for rows.Next() {
		var sub domain.ClientSubscription
		if err := rows.Scan(&sub.ID, &sub.ClientID, &sub.Type, &sub.Price, &sub.BoughtAt, &sub.EndsAt, &sub.LessonsLeft, &sub.Freeze, &sub.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (s *Store) CreateClientSubscription(sub domain.ClientSubscription) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO subscriptions(client_id, type, price, bought_at, ends_at, lessons_left, freeze) VALUES(?,?,?,?,?,?,?)`,
		sub.ClientID, sub.Type, sub.Price, sub.BoughtAt, sub.EndsAt, sub.LessonsLeft, sub.Freeze)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateClientSubscription(sub domain.ClientSubscription) error {
	_, err := s.DB.Exec(`UPDATE subscriptions SET type=?, price=?, ends_at=?, lessons_left=?, freeze=? WHERE id=? AND client_id=?`,
		sub.Type, sub.Price, sub.EndsAt, sub.LessonsLeft, sub.Freeze, sub.ID, sub.ClientID)
	return err
}

func (s *Store) DeleteClientSubscription(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM subscriptions WHERE id = ?`, id)
	return err
}

// ---------- Parent features ----------

func (s *Store) ClientByBirthDate(fullName, birthDate string) (*domain.Client, error) {
	c := &domain.Client{}
	var birth, photo, phone, tg, wa, email, med, note, reg, src, botAccess, subEnds sql.NullString
	var userID, age, parentID, secondParentID sql.NullInt64
	err := s.DB.QueryRow(`SELECT id, user_id, full_name, photo, birth_date, age, phone, telegram, whatsapp,
		parent_id, second_parent_id, email, medical_limits, note, status, registered_at, source, bot_access, subscription_ends_at
		FROM clients WHERE full_name = ? AND birth_date = ?`, fullName, birthDate).
		Scan(&c.ID, &userID, &c.FullName, &photo, &birth, &age, &phone, &tg, &wa,
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

func (s *Store) LinkParentToChild(parentUserID int64, childClientID int64) error {
	var currentParentID, currentSecondParentID sql.NullInt64
	err := s.DB.QueryRow(`SELECT parent_id, second_parent_id FROM clients WHERE id = ?`, childClientID).Scan(&currentParentID, &currentSecondParentID)
	if err != nil {
		return err
	}
	if currentParentID.Valid && currentParentID.Int64 == parentUserID {
		return nil
	}
	if currentSecondParentID.Valid && currentSecondParentID.Int64 == parentUserID {
		return nil
	}
	if !currentParentID.Valid {
		_, err = s.DB.Exec(`UPDATE clients SET parent_id = ? WHERE id = ?`, parentUserID, childClientID)
	} else if !currentSecondParentID.Valid {
		_, err = s.DB.Exec(`UPDATE clients SET second_parent_id = ? WHERE id = ?`, parentUserID, childClientID)
	}
	return err
}

func (s *Store) ListChildLessonEntries(childID int64, from, to string) ([]domain.ScheduleEntry, error) {
	return s.ListScheduleEntries(from, to, 0, childID)
}

// ---------- Parent invite codes ----------

func (s *Store) CreateInviteCode(clientID, createdBy int64, expiresAt string) (string, error) {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	code := make([]byte, 6)
	for {
		for i := range code {
			var b [1]byte
			if _, err := rand.Read(b[:]); err != nil {
				return "", err
			}
			code[i] = charset[int(b[0])%len(charset)]
		}
		c := string(code)
		var exists int
		err := s.DB.QueryRow(`SELECT 1 FROM parent_invite_codes WHERE code = ? AND used_at IS NULL AND expires_at > datetime('now')`, c).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			return "", err
		}
		if exists == 0 {
			_, err = s.DB.Exec(`INSERT INTO parent_invite_codes(client_id, code, created_by, expires_at) VALUES (?,?,?,?)`, clientID, c, createdBy, expiresAt)
			if err != nil {
				return "", err
			}
			return c, nil
		}
	}
}

func (s *Store) UseInviteCode(code string) (int64, error) {
	var clientID int64
	var usedAt sql.NullString
	err := s.DB.QueryRow(`SELECT client_id, used_at FROM parent_invite_codes WHERE code = ? AND expires_at > datetime('now')`, code).Scan(&clientID, &usedAt)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("invite code not found or expired")
	}
	if err != nil {
		return 0, err
	}
	if usedAt.Valid {
		return 0, fmt.Errorf("invite code already used")
	}
	_, err = s.DB.Exec(`UPDATE parent_invite_codes SET used_at = datetime('now') WHERE code = ?`, code)
	if err != nil {
		return 0, err
	}
	return clientID, nil
}

func (s *Store) UpsertParentNotifPref(pref domain.ParentNotifPref) error {
	_, err := s.DB.Exec(`INSERT INTO parent_notification_prefs(parent_user_id, child_id, lesson_start, lesson_end_15, lesson_missed)
		VALUES (?,?,?,?,?)
		ON CONFLICT(parent_user_id, child_id) DO UPDATE SET
		lesson_start=excluded.lesson_start, lesson_end_15=excluded.lesson_end_15, lesson_missed=excluded.lesson_missed`,
		pref.ParentUserID, pref.ChildID, boolToInt(pref.LessonStart), boolToInt(pref.LessonEnd15), boolToInt(pref.LessonMissed))
	return err
}

func (s *Store) GetParentNotifPrefs(parentUserID int64) ([]domain.ParentNotifPref, error) {
	rows, err := s.DB.Query(`SELECT id, parent_user_id, child_id, lesson_start, lesson_end_15, lesson_missed
		FROM parent_notification_prefs WHERE parent_user_id = ?`, parentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.ParentNotifPref
	for rows.Next() {
		var p domain.ParentNotifPref
		var start, end15, missed int
		if err := rows.Scan(&p.ID, &p.ParentUserID, &p.ChildID, &start, &end15, &missed); err != nil {
			return nil, err
		}
		p.LessonStart = start == 1
		p.LessonEnd15 = end15 == 1
		p.LessonMissed = missed == 1
		out = append(out, p)
	}
	return out, rows.Err()
}

// ---------- Coach Social Links ----------

func (s *Store) ListSocialLinks(coachID int64) ([]domain.SocialLink, error) {
	rows, err := s.DB.Query(`SELECT id, coach_id, platform, url, enabled FROM coach_social_links WHERE coach_id = ? ORDER BY platform`, coachID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.SocialLink
	for rows.Next() {
		var l domain.SocialLink
		var url sql.NullString
		var enabled int
		if err := rows.Scan(&l.ID, &l.CoachID, &l.Platform, &url, &enabled); err != nil {
			return nil, err
		}
		if url.Valid {
			l.URL = &url.String
		}
		l.Enabled = enabled == 1
		out = append(out, l)
	}
	return out, nil
}

func (s *Store) UpsertSocialLink(coachID int64, platform string, url *string, enabled bool) error {
	_, err := s.DB.Exec(`INSERT INTO coach_social_links(coach_id, platform, url, enabled, updated_at) VALUES (?,?,?,?,datetime('now'))
		ON CONFLICT(coach_id, platform) DO UPDATE SET url=excluded.url, enabled=excluded.enabled, updated_at=excluded.updated_at`,
		coachID, platform, url, boolToInt(enabled))
	return err
}

func (s *Store) SeedDefaultSocialLinks(coachID int64) error {
	platforms := []string{"instagram", "telegram", "vk", "youtube", "whatsapp"}
	for _, p := range platforms {
		_, err := s.DB.Exec(`INSERT OR IGNORE INTO coach_social_links(coach_id, platform) VALUES (?,?)`, coachID, p)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListParentUserIDsByChildID(childID int64) ([]int64, error) {
	rows, err := s.DB.Query(`SELECT parent_id, second_parent_id FROM clients WHERE id = ? AND (parent_id IS NOT NULL OR second_parent_id IS NOT NULL)`, childID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var p1, p2 sql.NullInt64
		if err := rows.Scan(&p1, &p2); err != nil {
			return nil, err
		}
		if p1.Valid {
			out = append(out, p1.Int64)
		}
		if p2.Valid {
			out = append(out, p2.Int64)
		}
	}
	return out, nil
}

// ---------- Groups ----------

func (s *Store) ListGroups() ([]domain.Group, error) {
	rows, err := s.DB.Query(`SELECT g.id, g.name, g.coach_id, g.max_members, g.schedule, g.price, g.location, g.active, COALESCE((SELECT COUNT(*) FROM group_members gm WHERE gm.group_id = g.id), 0) FROM groups g WHERE g.active = 1 ORDER BY g.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Group
	for rows.Next() {
		var g domain.Group
		var name, schedule, location sql.NullString
		var coachID sql.NullInt64
		var maxMembers sql.NullInt64
		var price sql.NullFloat64
		if err := rows.Scan(&g.ID, &name, &coachID, &maxMembers, &schedule, &price, &location, &g.Active, &g.MemberCount); err != nil {
			return nil, err
		}
		g.Name = nullStr(name)
		g.CoachID = nullInt64(coachID)
		g.MaxMembers = nullInt(maxMembers)
		g.Schedule = nullStr(schedule)
		if price.Valid {
			v := price.Float64
			g.Price = &v
		}
		g.Location = nullStr(location)
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) GetGroup(id int64) (*domain.Group, error) {
	g := &domain.Group{}
	var name, schedule, location sql.NullString
	var coachID sql.NullInt64
	var maxMembers sql.NullInt64
	var price sql.NullFloat64
	err := s.DB.QueryRow(`SELECT g.id, g.name, g.coach_id, g.max_members, g.schedule, g.price, g.location, g.active, COALESCE((SELECT COUNT(*) FROM group_members gm WHERE gm.group_id = g.id), 0) FROM groups g WHERE g.id = ?`, id).
		Scan(&g.ID, &name, &coachID, &maxMembers, &schedule, &price, &location, &g.Active, &g.MemberCount)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	g.Name = nullStr(name)
	g.CoachID = nullInt64(coachID)
	g.MaxMembers = nullInt(maxMembers)
	g.Schedule = nullStr(schedule)
	if price.Valid {
		v := price.Float64
		g.Price = &v
	}
	g.Location = nullStr(location)
	return g, nil
}

func (s *Store) CreateGroup(g domain.Group) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO groups(name, coach_id, max_members, schedule, price, location, active) VALUES(?,?,?,?,?,?,?)`,
		nullStrPtr(g.Name), nullInt64Ptr(g.CoachID), nullIntPtr(g.MaxMembers), nullStrPtr(g.Schedule), nullFloat64Ptr(g.Price), nullStrPtr(g.Location), g.Active)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateGroup(g domain.Group) error {
	_, err := s.DB.Exec(`UPDATE groups SET name=?, coach_id=?, max_members=?, schedule=?, price=?, location=?, active=? WHERE id=?`,
		nullStrPtr(g.Name), nullInt64Ptr(g.CoachID), nullIntPtr(g.MaxMembers), nullStrPtr(g.Schedule), nullFloat64Ptr(g.Price), nullStrPtr(g.Location), g.Active, g.ID)
	return err
}

func (s *Store) DeleteGroup(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM groups WHERE id = ?`, id)
	return err
}

func (s *Store) AddClientToGroup(groupID, clientID int64, role string) error {
	_, err := s.DB.Exec(`INSERT OR IGNORE INTO group_members(group_id, client_id, role) VALUES(?,?,?)`, groupID, clientID, role)
	return err
}

func (s *Store) RemoveClientFromGroup(groupID, clientID int64) error {
	_, err := s.DB.Exec(`DELETE FROM group_members WHERE group_id = ? AND client_id = ?`, groupID, clientID)
	return err
}

func (s *Store) GetGroupClients(groupID int64) ([]domain.GroupMember, error) {
	rows, err := s.DB.Query(`SELECT gm.id, gm.group_id, gm.client_id, gm.role, gm.joined_at, c.full_name FROM group_members gm JOIN clients c ON c.id = gm.client_id WHERE gm.group_id = ? ORDER BY c.full_name`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.GroupMember
	for rows.Next() {
		var m domain.GroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.ClientID, &m.Role, &m.JoinedAt, &m.ClientName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) GetClientGroups(clientID int64) ([]domain.Group, error) {
	rows, err := s.DB.Query(`SELECT g.id, g.name, g.coach_id, g.max_members, g.schedule, g.price, g.location, g.active FROM groups g JOIN group_members gm ON gm.group_id = g.id WHERE gm.client_id = ? AND g.active = 1 ORDER BY g.name`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Group
	for rows.Next() {
		var g domain.Group
		var name, schedule, location sql.NullString
		var coachID sql.NullInt64
		var maxMembers sql.NullInt64
		var price sql.NullFloat64
		if err := rows.Scan(&g.ID, &name, &coachID, &maxMembers, &schedule, &price, &location, &g.Active); err != nil {
			return nil, err
		}
		g.Name = nullStr(name)
		g.CoachID = nullInt64(coachID)
		g.MaxMembers = nullInt(maxMembers)
		g.Schedule = nullStr(schedule)
		if price.Valid {
			v := price.Float64
			g.Price = &v
		}
		g.Location = nullStr(location)
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) ClientsNotInGroup(groupID int64) ([]domain.Client, error) {
	rows, err := s.DB.Query(`SELECT c.id, c.user_id, c.full_name, c.age, c.phone, c.status FROM clients c WHERE c.id NOT IN (SELECT gm.client_id FROM group_members gm WHERE gm.group_id = ?) AND c.status = 'active' AND (c.user_id IS NULL OR c.user_id NOT IN (SELECT id FROM users WHERE role IN ('admin', 'coach'))) ORDER BY c.full_name`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Client
	for rows.Next() {
		var c domain.Client
		var userID sql.NullInt64
		var fullName, phone, status sql.NullString
		var age sql.NullInt64
		if err := rows.Scan(&c.ID, &userID, &fullName, &age, &phone, &status); err != nil {
			return nil, err
		}
		if userID.Valid {
			c.UserID = &userID.Int64
		}
		if fullName.Valid {
			c.FullName = fullName.String
		}
		if age.Valid {
			v := int(age.Int64)
			c.Age = &v
		}
		if phone.Valid {
			c.Phone = &phone.String
		}
		if status.Valid {
			c.Status = status.String
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
