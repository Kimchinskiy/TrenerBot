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

func (s *Store) StudentByUserID(userID int64) (*domain.Student, error) {
	var st domain.Student
	var photo, birth, phone, tg, wa, email, med, note, reg, src, level, addInfo, updatedAt, coachID sql.NullString
	var age sql.NullInt64
	var botAccess sql.NullString
	err := s.DB.QueryRow(`SELECT id, user_id, full_name, photo, birth_date, age, phone, telegram, whatsapp,
		email, level, additional_info, medical_limits, note, status, registered_at, source, bot_access, coach_id, created_at, updated_at
		FROM students WHERE user_id = ?`, userID).
		Scan(&st.ID, &st.UserID, &st.FullName, &photo, &birth, &age, &phone, &tg, &wa,
			&email, &level, &addInfo, &med, &note, &st.Status, &reg, &src, &botAccess, &coachID, &st.CreatedAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	scanStudentFields(&st, photo, birth, phone, tg, wa, email, level, addInfo, med, note, reg, src, botAccess, coachID, updatedAt, age)
	return &st, nil
}

func (s *Store) StudentByID(id int64) (*domain.Student, error) {
	var st domain.Student
	var photo, birth, phone, tg, wa, email, med, note, reg, src, level, addInfo, updatedAt, coachID sql.NullString
	var age sql.NullInt64
	var botAccess sql.NullString
	err := s.DB.QueryRow(`SELECT id, user_id, full_name, photo, birth_date, age, phone, telegram, whatsapp,
		email, level, additional_info, medical_limits, note, status, registered_at, source, bot_access, coach_id, created_at, updated_at
		FROM students WHERE id = ?`, id).
		Scan(&st.ID, &st.UserID, &st.FullName, &photo, &birth, &age, &phone, &tg, &wa,
			&email, &level, &addInfo, &med, &note, &st.Status, &reg, &src, &botAccess, &coachID, &st.CreatedAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	scanStudentFields(&st, photo, birth, phone, tg, wa, email, level, addInfo, med, note, reg, src, botAccess, coachID, updatedAt, age)
	return &st, nil
}

func scanStudentFields(st *domain.Student, photo, birth, phone, tg, wa, email, level, addInfo, med, note, reg, src, botAccess, coachID, updatedAt sql.NullString, age sql.NullInt64) {
	if photo.Valid { st.Photo = &photo.String }
	if birth.Valid { st.BirthDate = &birth.String }
	if age.Valid { v := int(age.Int64); st.Age = &v }
	if phone.Valid { st.Phone = &phone.String }
	if tg.Valid { st.Telegram = &tg.String }
	if wa.Valid { st.WhatsApp = &wa.String }
	if email.Valid { st.Email = &email.String }
	if level.Valid { st.Level = level.String }
	if addInfo.Valid { st.AdditionalInfo = &addInfo.String }
	if med.Valid { st.MedicalLimits = &med.String }
	if note.Valid { st.Note = &note.String }
	if reg.Valid { st.RegisteredAt = &reg.String }
	if src.Valid { st.Source = &src.String }
	if botAccess.Valid { st.BotAccess = botAccess.String == "1" }
	if coachID.Valid { v, _ := strconv.ParseInt(coachID.String, 10, 64); st.CoachID = &v }
	if updatedAt.Valid { st.UpdatedAt = &updatedAt.String }
}

func (s *Store) CreateStudentFull(st domain.Student) (int64, error) {
	botAccess := 0
	if st.BotAccess {
		botAccess = 1
	}
	res, err := s.DB.Exec(`INSERT INTO students(user_id, full_name, photo, birth_date, age, phone, telegram,
		whatsapp, email, level, additional_info, medical_limits, note, status, registered_at, source, bot_access, coach_id)
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		st.UserID, st.FullName, st.Photo, st.BirthDate, st.Age, st.Phone, st.Telegram, st.WhatsApp,
		st.Email, st.Level, st.AdditionalInfo, st.MedicalLimits, st.Note, st.Status, st.RegisteredAt, st.Source, botAccess, st.CoachID)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	// FK compat: ensure stub row exists in clients table
	_, _ = s.DB.Exec("INSERT OR IGNORE INTO clients(id) VALUES (?)", id)
	return id, nil
}

func (s *Store) UpdateStudent(st domain.Student) error {
	botAccess := 0
	if st.BotAccess {
		botAccess = 1
	}
	slog.Debug("UpdateStudent", "id", st.ID, "botAccess", botAccess)
	_, err := s.DB.Exec(`UPDATE students SET full_name=?, photo=?, birth_date=?, age=?, phone=?, telegram=?,
		whatsapp=?, email=?, level=?, additional_info=?, medical_limits=?, note=?, status=?, source=?, bot_access=?, coach_id=?
		WHERE id=?`,
		st.FullName, st.Photo, st.BirthDate, st.Age, st.Phone, st.Telegram, st.WhatsApp,
		st.Email, st.Level, st.AdditionalInfo, st.MedicalLimits, st.Note, st.Status, st.Source, botAccess, st.CoachID, st.ID)
	return err
}

func (s *Store) SetStudentCoachID(studentID, coachID int64) error {
	_, err := s.DB.Exec(`UPDATE students SET coach_id = ? WHERE id = ?`, coachID, studentID)
	return err
}

func (s *Store) ListStudents() ([]domain.Student, error) {
	rows, err := s.DB.Query(`SELECT s.id, s.user_id, s.full_name, s.status, s.bot_access FROM students s WHERE s.full_name != '' AND s.full_name IS NOT NULL AND (s.user_id IS NULL OR s.user_id NOT IN (SELECT id FROM users WHERE role IN ('admin', 'coach'))) AND s.id IN (SELECT MIN(s2.id) FROM students s2 WHERE s2.full_name != '' AND s2.full_name IS NOT NULL GROUP BY CASE WHEN s2.user_id IS NULL THEN s2.full_name ELSE CAST(s2.user_id AS TEXT) END) ORDER BY s.full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Student
	for rows.Next() {
		var st domain.Student
		var uid sql.NullInt64
		var botAccess sql.NullString
		if err := rows.Scan(&st.ID, &uid, &st.FullName, &st.Status, &botAccess); err != nil {
			return nil, err
		}
		if uid.Valid {
			v := uid.Int64
			st.UserID = &v
		}
		if botAccess.Valid {
			st.BotAccess = botAccess.String == "1"		}
		out = append(out, st)
	}
	return out, rows.Err()
}

// Children of a parent user (by parent_id linking to users.id).
func (s *Store) ChildrenOfParent(parentUserID int64) ([]domain.Student, error) {
	rows, err := s.DB.Query(`SELECT s.id, s.full_name, s.status FROM students s JOIN relationships r ON r.student_id = s.id WHERE r.user_id = ? AND r.relation = 'parent' ORDER BY s.full_name`, parentUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Student
	for rows.Next() {
		var s domain.Student
		if err := rows.Scan(&s.ID, &s.FullName, &s.Status); err != nil {
			return nil, err
		}
		out = append(out, s)
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
		if err := rows.Scan(&a.LessonID, &a.StudentID, &present, &markedAt, &markedBy); err != nil {
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

func (s *Store) ListAttendanceByStudent(studentID int64) ([]domain.Attendance, error) {
	rows, err := s.DB.Query(`SELECT lesson_id, client_id, present, marked_at, marked_by FROM attendance WHERE client_id = ?`, studentID)
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
		if err := rows.Scan(&a.LessonID, &a.StudentID, &present, &markedAt, &markedBy); err != nil {
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
		VALUES (?,?,?,?,?,?,?,?)`, e.Date, e.Time, e.StudentID, e.CoachID, e.GroupID, e.Duration, string(e.Status), e.Comment)
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
		JOIN students c ON c.id = le.client_id
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
		if err := rows.Scan(&e.ID, &e.Date, &e.Time, &e.StudentID, &e.StudentName, &coachIDVal, &e.Duration, &e.Status, &groupIDVal, &groupName); err != nil {
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
	StudentID  int64
	UserID     int64
	LessonID   int64
	Date       string
	Time       string
	Location   *string
	CoachID    *int64
}

// ListLessonEntriesForReminders returns planned lesson entries within [from, to]
// joined with the student's user_id for notification delivery.
func (s *Store) ListLessonEntriesForReminders(from, to string) ([]ReminderEntry, error) {
	rows, err := s.DB.Query(`SELECT le.id, le.date, le.time, le.client_id, c.user_id, le.coach_id
		FROM lesson_entries le
		JOIN students c ON c.id = le.client_id
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
		if err := rows.Scan(&e.LessonID, &e.Date, &e.Time, &e.StudentID, &uid, &coachID); err != nil {
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
				FROM students c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.coach_id = ? AND le.date = ?`, coachID, date)
		} else {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM students c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.date = ?`, date)
		}
	} else if groupID > 0 {
		if coachID > 0 {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM students c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.coach_id = ? AND le.group_id = ?`, coachID, groupID)
		} else {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM students c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.group_id = ?`, groupID)
		}
	} else {
		if coachID > 0 {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM students c
				JOIN lesson_entries le ON le.client_id = c.id
				WHERE le.coach_id = ?`, coachID)
		} else {
			rows, err = s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.user_id
				FROM students c
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
		if err := rows.Scan(&r.StudentID, &r.FullName, &r.UserID); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ---------- Daily attendance ----------

func (s *Store) ListStudentsWithLessonsOnDate(date string) ([]domain.DateAttendanceClient, error) {
	rows, err := s.DB.Query(`SELECT DISTINCT c.id, c.full_name, c.photo, le.time
		FROM students c
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
		if err := rows.Scan(&r.StudentID, &r.FullName, &avatar, &r.Time); err != nil {
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
			e.Date, e.StudentID, boolToInt(e.Present), e.MarkedBy)
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
	rows, err := s.DB.Query(`SELECT id, full_name, user_id FROM students WHERE id IN(`+placeholders[1:]+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Recipient
	for rows.Next() {
		var r domain.Recipient
		if err := rows.Scan(&r.StudentID, &r.FullName, &r.UserID); err != nil {
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

func (s *Store) ClientSubscriptions(clientID int64) ([]domain.Subscription, error) {
	rows, err := s.DB.Query(`SELECT id, client_id, type, price, bought_at, ends_at, lessons_left, freeze, created_at FROM subscriptions WHERE client_id = ? ORDER BY created_at DESC`, clientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Subscription
	for rows.Next() {
		var sub domain.Subscription
		if err := rows.Scan(&sub.ID, &sub.StudentID, &sub.Type, &sub.Price, &sub.BoughtAt, &sub.EndsAt, &sub.LessonsLeft, &sub.Freeze, &sub.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (s *Store) CreateClientSubscription(sub domain.Subscription) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO subscriptions(client_id, type, price, bought_at, ends_at, lessons_left, freeze) VALUES(?,?,?,?,?,?,?)`,
		sub.StudentID, sub.Type, sub.Price, sub.BoughtAt, sub.EndsAt, sub.LessonsLeft, sub.Freeze)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdateClientSubscription(sub domain.Subscription) error {
	_, err := s.DB.Exec(`UPDATE subscriptions SET type=?, price=?, ends_at=?, lessons_left=?, freeze=? WHERE id=? AND client_id=?`,
		sub.Type, sub.Price, sub.EndsAt, sub.LessonsLeft, sub.Freeze, sub.ID, sub.StudentID)
	return err
}

func (s *Store) DeleteClientSubscription(id int64) error {
	_, err := s.DB.Exec(`DELETE FROM subscriptions WHERE id = ?`, id)
	return err
}

// ---------- Parent features ----------

func (s *Store) StudentByBirthDate(fullName, birthDate string) (*domain.Student, error) {
	var st domain.Student
	var photo, phone, tg, wa, email, med, note, reg, src, botAccess sql.NullString
	var userID, age sql.NullInt64
	var birth sql.NullString
	err := s.DB.QueryRow(`SELECT id, user_id, full_name, photo, birth_date, age, phone, telegram, whatsapp,
		email, medical_limits, note, status, registered_at, source, bot_access
		FROM students WHERE full_name = ? AND birth_date = ?`, fullName, birthDate).
		Scan(&st.ID, &userID, &st.FullName, &photo, &birth, &age, &phone, &tg, &wa,
			&email, &med, &note, &st.Status, &reg, &src, &botAccess)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if userID.Valid {
		st.UserID = &userID.Int64
	}
	if photo.Valid {
		st.Photo = &photo.String
	}
	if birth.Valid {
		st.BirthDate = &birth.String
	}
	if age.Valid {
		v := int(age.Int64)
		st.Age = &v
	}
	if phone.Valid {
		st.Phone = &phone.String
	}
	if tg.Valid {
		st.Telegram = &tg.String
	}
	if wa.Valid {
		st.WhatsApp = &wa.String
	}
	if email.Valid {
		st.Email = &email.String
	}
	if med.Valid {
		st.MedicalLimits = &med.String
	}
	if note.Valid {
		st.Note = &note.String
	}
	if reg.Valid {
		st.RegisteredAt = &reg.String
	}
	if src.Valid {
		st.Source = &src.String
	}
	if botAccess.Valid {
		st.BotAccess = botAccess.String == "1"
	}
	return &st, nil
}

func (s *Store) LinkParentToChild(parentUserID int64, studentID int64) error {
	_, err := s.DB.Exec(`INSERT OR IGNORE INTO relationships(user_id, student_id, relation) VALUES(?, ?, 'parent')`, parentUserID, studentID)
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
		pref.ParentUserID, pref.StudentID, boolToInt(pref.LessonStart), boolToInt(pref.LessonEnd15), boolToInt(pref.LessonMissed))
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
		if err := rows.Scan(&p.ID, &p.ParentUserID, &p.StudentID, &start, &end15, &missed); err != nil {
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

func (s *Store) ListParentUserIDsByChildID(studentID int64) ([]int64, error) {
	rows, err := s.DB.Query(`SELECT user_id FROM relationships WHERE student_id = ? AND relation = 'parent'`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var uid int64
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		out = append(out, uid)
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

func (s *Store) AddStudentToGroup(groupID, studentID int64, role string) error {
	_, err := s.DB.Exec(`INSERT OR IGNORE INTO group_members(group_id, client_id, role) VALUES(?, ?, ?)`, groupID, studentID, role)
	return err
}

func (s *Store) RemoveStudentFromGroup(groupID, studentID int64) error {
	_, err := s.DB.Exec(`DELETE FROM group_members WHERE group_id = ? AND client_id = ?`, groupID, studentID)
	return err
}

func (s *Store) GetGroupStudents(groupID int64) ([]domain.GroupMember, error) {
	rows, err := s.DB.Query(`SELECT gm.id, gm.group_id, gm.client_id, gm.role, gm.joined_at, c.full_name FROM group_members gm JOIN students c ON c.id = gm.client_id WHERE gm.group_id = ? ORDER BY c.full_name`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.GroupMember
	for rows.Next() {
		var m domain.GroupMember
		if err := rows.Scan(&m.ID, &m.GroupID, &m.StudentID, &m.Role, &m.JoinedAt, &m.StudentName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *Store) GetStudentGroups(studentID int64) ([]domain.Group, error) {
	rows, err := s.DB.Query(`SELECT g.id, g.name, g.coach_id, g.max_members, g.schedule, g.price, g.location, g.active FROM groups g JOIN group_members gm ON gm.group_id = g.id WHERE gm.client_id = ? AND g.active = 1 ORDER BY g.name`, studentID)
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

func (s *Store) StudentsNotInGroup(groupID int64) ([]domain.Student, error) {
	rows, err := s.DB.Query(`SELECT c.id, c.user_id, c.full_name, c.age, c.phone, c.status FROM students c WHERE c.id NOT IN (SELECT gm.client_id FROM group_members gm WHERE gm.group_id = ?) AND c.status = 'active' AND (c.user_id IS NULL OR c.user_id NOT IN (SELECT id FROM users WHERE role IN ('admin', 'coach'))) ORDER BY c.full_name`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Student
	for rows.Next() {
		var st domain.Student
		var userID sql.NullInt64
		var fullName, phone, status sql.NullString
		var age sql.NullInt64
		if err := rows.Scan(&st.ID, &userID, &fullName, &age, &phone, &status); err != nil {
			return nil, err
		}
		if userID.Valid {
			st.UserID = &userID.Int64
		}
		if fullName.Valid {
			st.FullName = fullName.String
		}
		if age.Valid {
			v := int(age.Int64)
			st.Age = &v
		}
		if phone.Valid {
			st.Phone = &phone.String
		}
		if status.Valid {
			st.Status = status.String
		}
		out = append(out, st)
	}
	return out, rows.Err()
}
