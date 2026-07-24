package store

import (
	"database/sql"
	"trenerbot/internal/domain"
)

// ---------- Students ----------

func (s *Store) CreateStudent(st domain.Student) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO students(full_name, birth_date, age, level, phone, additional_info, status, client_id) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		st.FullName, st.BirthDate, st.Age, st.Level, st.Phone, st.AdditionalInfo, st.Status, st.ClientID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) StudentByID(id int64) (*domain.Student, error) {
	row := s.DB.QueryRow(`SELECT id, full_name, birth_date, age, level, phone, additional_info, status, client_id, created_at, updated_at FROM students WHERE id = ?`, id)
	return scanStudent(row)
}

func (s *Store) StudentsByUserID(userID int64) ([]domain.Student, error) {
	rows, err := s.DB.Query(`SELECT s.id, s.full_name, s.birth_date, s.age, s.level, s.phone, s.additional_info, s.status, s.client_id, s.created_at, s.updated_at FROM students s JOIN relationships r ON r.student_id = s.id WHERE r.user_id = ? AND s.status = 'active' ORDER BY s.full_name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var students []domain.Student
	for rows.Next() {
		st, err := scanStudent(rows)
		if err != nil {
			return nil, err
		}
		students = append(students, *st)
	}
	return students, rows.Err()
}

func scanStudent(row interface{ Scan(...any) error }) (*domain.Student, error) {
	var st domain.Student
	var birthDate, phone, addInfo, updatedAt sql.NullString
	var age sql.NullInt64
	var clientID sql.NullInt64
	err := row.Scan(&st.ID, &st.FullName, &birthDate, &age, &st.Level, &phone, &addInfo, &st.Status, &clientID, &st.CreatedAt, &updatedAt)
	if err != nil {
		return nil, err
	}
	st.BirthDate = nullStr(birthDate)
	st.Age = nullInt(age)
	st.Phone = nullStr(phone)
	st.AdditionalInfo = nullStr(addInfo)
	st.ClientID = nullInt64(clientID)
	if updatedAt.Valid {
		str := updatedAt.String
		st.UpdatedAt = &str
	}
	return &st, nil
}

// ---------- Relationships ----------

func (s *Store) CreateRelationship(rel domain.Relationship) (int64, error) {
	res, err := s.DB.Exec(`INSERT OR IGNORE INTO relationships(user_id, student_id, relation) VALUES (?, ?, ?)`,
		rel.UserID, rel.StudentID, string(rel.Relation))
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) RelationshipsByStudent(studentID int64) ([]domain.Relationship, error) {
	rows, err := s.DB.Query(`SELECT id, user_id, student_id, relation, created_at FROM relationships WHERE student_id = ?`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rels []domain.Relationship
	for rows.Next() {
		var r domain.Relationship
		if err := rows.Scan(&r.ID, &r.UserID, &r.StudentID, &r.Relation, &r.CreatedAt); err != nil {
			return nil, err
		}
		rels = append(rels, r)
	}
	return rels, rows.Err()
}

// ---------- Leads ----------

func (s *Store) CreateLead(l domain.Lead) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO leads(telegram_id, full_name, phone, target_name, target_age, target_level, reg_type, status) VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')`,
		l.TelegramID, l.FullName, l.Phone, l.TargetName, l.TargetAge, l.TargetLevel, l.RegType)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) LeadByID(id int64) (*domain.Lead, error) {
	var l domain.Lead
	var phone, targetName sql.NullString
	var targetAge sql.NullInt64
	var reviewedAt sql.NullString
	var reviewedBy, createdUID, createdSID sql.NullInt64
	err := s.DB.QueryRow(`SELECT id, telegram_id, full_name, phone, target_name, target_age, target_level, reg_type, status, created_at, reviewed_at, reviewed_by, created_user_id, created_student_id FROM leads WHERE id = ?`, id).
		Scan(&l.ID, &l.TelegramID, &l.FullName, &phone, &targetName, &targetAge, &l.TargetLevel, &l.RegType, &l.Status, &l.CreatedAt, &reviewedAt, &reviewedBy, &createdUID, &createdSID)
	if err != nil {
		return nil, err
	}
	l.Phone = nullStr(phone)
	l.TargetName = nullStr(targetName)
	l.TargetAge = nullInt(targetAge)
	if reviewedAt.Valid {
		str := reviewedAt.String
		l.ReviewedAt = &str
	}
	l.ReviewedBy = nullInt64(reviewedBy)
	l.CreatedUserID = nullInt64(createdUID)
	l.CreatedStudentID = nullInt64(createdSID)
	return &l, nil
}

func (s *Store) PendingLeads() ([]domain.Lead, error) {
	rows, err := s.DB.Query(`SELECT id, telegram_id, full_name, phone, target_name, target_age, target_level, reg_type, status, created_at FROM leads WHERE status = 'pending' ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var leads []domain.Lead
	for rows.Next() {
		var l domain.Lead
		var phone, targetName sql.NullString
		var targetAge sql.NullInt64
		if err := rows.Scan(&l.ID, &l.TelegramID, &l.FullName, &phone, &targetName, &targetAge, &l.TargetLevel, &l.RegType, &l.Status, &l.CreatedAt); err != nil {
			return nil, err
		}
		l.Phone = nullStr(phone)
		l.TargetName = nullStr(targetName)
		l.TargetAge = nullInt(targetAge)
		leads = append(leads, l)
	}
	return leads, rows.Err()
}

func (s *Store) LeadByTelegram(telegramID string) (*domain.Lead, error) {
	var l domain.Lead
	var phone, targetName sql.NullString
	var targetAge sql.NullInt64
	err := s.DB.QueryRow(`SELECT id, telegram_id, full_name, phone, target_name, target_age, target_level, reg_type, status, created_at FROM leads WHERE telegram_id = ? AND status = 'pending' ORDER BY created_at DESC LIMIT 1`, telegramID).
		Scan(&l.ID, &l.TelegramID, &l.FullName, &phone, &targetName, &targetAge, &l.TargetLevel, &l.RegType, &l.Status, &l.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	l.Phone = nullStr(phone)
	l.TargetName = nullStr(targetName)
	l.TargetAge = nullInt(targetAge)
	return &l, nil
}

func (s *Store) ApproveLead(id int64, userID int64, studentID int64, reviewedBy int64) error {
	_, err := s.DB.Exec(`UPDATE leads SET status = 'approved', reviewed_at = datetime('now'), reviewed_by = ?, created_user_id = ?, created_student_id = ? WHERE id = ?`,
		reviewedBy, userID, studentID, id)
	return err
}

func (s *Store) RejectLead(id int64, reviewedBy int64) error {
	_, err := s.DB.Exec(`UPDATE leads SET status = 'rejected', reviewed_at = datetime('now'), reviewed_by = ? WHERE id = ?`, reviewedBy, id)
	return err
}

// ---------- Training Templates ----------

func (s *Store) TemplatesByGroup(groupID int64) ([]domain.TrainingTemplate, error) {
	rows, err := s.DB.Query(`SELECT id, group_id, weekday, time, duration FROM training_templates WHERE group_id = ? ORDER BY weekday, time`, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var templates []domain.TrainingTemplate
	for rows.Next() {
		var t domain.TrainingTemplate
		if err := rows.Scan(&t.ID, &t.GroupID, &t.Weekday, &t.Time, &t.Duration); err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

// ---------- Trainings ----------

func (s *Store) CreateTraining(t domain.Training) (int64, error) {
	res, err := s.DB.Exec(`INSERT INTO trainings(group_id, coach_id, date, time, duration, status, location, comment) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		t.GroupID, t.CoachID, t.Date, t.Time, t.Duration, t.Status, t.Location, t.Comment)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) TrainingsByDateRange(from, to string, studentID int64) ([]domain.Training, error) {
	rows, err := s.DB.Query(`
		SELECT t.id, t.group_id, t.coach_id, t.date, t.time, t.duration, t.status, t.location, t.comment
		FROM trainings t
		JOIN group_members gm ON gm.group_id = t.group_id
		JOIN relationships r ON r.student_id = gm.client_id
		WHERE t.date >= ? AND t.date <= ? AND r.student_id = ? AND t.status = 'planned'
		ORDER BY t.date, t.time`, from, to, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var trainings []domain.Training
	for rows.Next() {
		var t domain.Training
		var groupID, coachID sql.NullInt64
		var location, comment sql.NullString
		if err := rows.Scan(&t.ID, &groupID, &coachID, &t.Date, &t.Time, &t.Duration, &t.Status, &location, &comment); err != nil {
			return nil, err
		}
		t.GroupID = nullInt64(groupID)
		t.CoachID = nullInt64(coachID)
		t.Location = nullStr(location)
		t.Comment = nullStr(comment)
		trainings = append(trainings, t)
	}
	return trainings, rows.Err()
}

// ---------- Training Absences ----------

func (s *Store) CreateAbsence(trainingID, studentID int64, reason string) error {
	_, err := s.DB.Exec(`INSERT OR IGNORE INTO training_absences(training_id, student_id, reason) VALUES (?, ?, ?)`,
		trainingID, studentID, reason)
	return err
}

// ---------- Notification Prefs ----------

func (s *Store) GetNotifPrefs(userID int64, studentID int64) (*domain.NotificationPref, error) {
	var p domain.NotificationPref
	var sid sql.NullInt64
	err := s.DB.QueryRow(`SELECT user_id, student_id, reminder_day, reminder_hours, lessons_low, sub_expiring, news FROM notification_prefs WHERE user_id = ? AND (student_id = ? OR (student_id IS NULL AND ? = 0))`,
		userID, studentID, studentID).Scan(&p.UserID, &sid, &p.ReminderDay, &p.ReminderHours, &p.LessonsLow, &p.SubExpiring, &p.News)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.StudentID = nullInt64(sid)
	return &p, nil
}

func (s *Store) UpsertNotifPrefs(p domain.NotificationPref) error {
	_, err := s.DB.Exec(`INSERT INTO notification_prefs(user_id, student_id, reminder_day, reminder_hours, lessons_low, sub_expiring, news) VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, student_id) DO UPDATE SET reminder_day=excluded.reminder_day, reminder_hours=excluded.reminder_hours, lessons_low=excluded.lessons_low, sub_expiring=excluded.sub_expiring, news=excluded.news, updated_at=datetime('now')`,
		p.UserID, p.StudentID, p.ReminderDay, p.ReminderHours, p.LessonsLow, p.SubExpiring, p.News)
	return err
}

// ---------- Client Messages ----------

func (s *Store) SaveClientMessage(userID int64, studentID *int64, text string) error {
	_, err := s.DB.Exec(`INSERT INTO client_messages(user_id, student_id, text) VALUES (?, ?, ?)`, userID, studentID, text)
	return err
}
