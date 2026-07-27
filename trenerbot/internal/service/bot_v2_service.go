package service

import (
	"database/sql"
	"time"

	"trenerbot/internal/domain"
)

// ---------- Students ----------

func (s *Services) CreateStudent(st domain.Student) (int64, error) {
	return s.Store.CreateStudentFull(st)
}

func (s *Services) StudentByID(id int64) (*domain.Student, error) {
	return s.Store.StudentByID(id)
}

func (s *Services) StudentsByUserID(userID int64) ([]domain.Student, error) {
	return s.Store.StudentsByUserID(userID)
}

// ---------- Relationships ----------

func (s *Services) CreateRelationship(rel domain.Relationship) (int64, error) {
	return s.Store.CreateRelationship(rel)
}

// ---------- Leads ----------

func (s *Services) CreateLead(l domain.Lead) (int64, error) {
	return s.Store.CreateLead(l)
}

func (s *Services) PendingLeads() ([]domain.Lead, error) {
	return s.Store.PendingLeads()
}

func (s *Services) LeadByID(id int64) (*domain.Lead, error) {
	return s.Store.LeadByID(id)
}

func (s *Services) ApproveLead(leadID int64, reviewedBy int64) (*domain.Lead, error) {
	lead, err := s.Store.LeadByID(leadID)
	if err != nil {
		return nil, err
	}
	if lead.Status != domain.LeadPending {
		return lead, nil
	}

	tx, err := s.Store.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var userID int64
	err = tx.QueryRow(`SELECT id FROM users WHERE telegram_id = ?`, lead.TelegramID).Scan(&userID)
	if err == sql.ErrNoRows {
		res, err := tx.Exec(`INSERT INTO users(telegram_id, role, updated_at) VALUES (?, ?, datetime('now'))`, lead.TelegramID, string(domain.RoleClient))
		if err != nil {
			return nil, err
		}
		uid, err := res.LastInsertId()
		if err != nil {
			return nil, err
		}
		userID = uid
		if lead.Phone != nil && *lead.Phone != "" {
			_, err = tx.Exec(`UPDATE users SET phone = ?, first_name = ?, updated_at = datetime('now') WHERE id = ?`, lead.Phone, lead.FullName, uid)
			if err != nil {
				return nil, err
			}
		}
	} else if err != nil {
		return nil, err
	}

	studentName := lead.FullName
	studentAge := lead.TargetAge
	if lead.RegType == "child" && lead.TargetName != nil {
		studentName = *lead.TargetName
	}
	res, err := tx.Exec(`INSERT INTO students(user_id, full_name, age, level, phone, status) VALUES (?, ?, ?, ?, ?, 'active')`,
		userID, studentName, studentAge, lead.TargetLevel, lead.Phone)
	if err != nil {
		return nil, err
	}
	studentID, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}
	// FK compatibility
	_, _ = tx.Exec("INSERT OR IGNORE INTO clients(id) VALUES (?)", studentID)

	relation := domain.RelSelf
	if lead.RegType == "child" {
		relation = domain.RelParent
	}
	_, err = tx.Exec(`INSERT OR IGNORE INTO relationships(user_id, student_id, relation) VALUES (?, ?, ?)`,
		userID, studentID, string(relation))
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(`UPDATE leads SET status = 'approved', reviewed_at = datetime('now'), reviewed_by = ?, created_user_id = ?, created_student_id = ? WHERE id = ?`,
		reviewedBy, userID, studentID, leadID)
	if err != nil {
		return nil, err
	}

	// Link student to coach
	var coachID int64
	err = tx.QueryRow(`SELECT id FROM coaches WHERE user_id = ?`, reviewedBy).Scan(&coachID)
	if err == nil && coachID > 0 {
		_, err = tx.Exec(`UPDATE students SET coach_id = ? WHERE id = ?`, coachID, studentID)
		if err != nil {
			return nil, err
		}
	} else if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	go s.Notify(userID, "lead_approved", map[string]any{
		"student_name": studentName,
	}, time.Now())

	return s.Store.LeadByID(leadID)
}

func (s *Services) RejectLead(leadID int64, reviewedBy int64) error {
	return s.Store.RejectLead(leadID, reviewedBy)
}

// ---------- Trainings ----------

func (s *Services) StudentTrainings(studentID int64, from, to string) ([]domain.Training, error) {
	return s.Store.TrainingsByDateRange(from, to, studentID)
}

// ---------- Absences ----------

func (s *Services) ReportAbsence(trainingID, studentID int64, reason string) error {
	return s.Store.CreateAbsence(trainingID, studentID, reason)
}

// ---------- Notification Prefs ----------

func (s *Services) GetNotifPrefs(userID int64, studentID int64) (*domain.NotificationPref, error) {
	return s.Store.GetNotifPrefs(userID, studentID)
}

func (s *Services) SaveNotifPrefs(p domain.NotificationPref) error {
	return s.Store.UpsertNotifPrefs(p)
}

// ---------- Client Messages ----------

func (s *Services) SaveClientMessage(userID int64, studentID *int64, text string) error {
	return s.Store.SaveClientMessage(userID, studentID, text)
}
