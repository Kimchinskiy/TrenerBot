package service

import (
	"time"
	"trenerbot/internal/domain"
)

// ---------- Students ----------

func (s *Services) CreateStudent(st domain.Student) (int64, error) {
	return s.Store.CreateStudent(st)
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

	var userID int64
	u, err := s.Store.UserByTelegram(lead.TelegramID)
	if err != nil {
		return nil, err
	}
	if u != nil {
		userID = u.ID
	} else {
		uid, err := s.Store.CreateUser(&lead.TelegramID, domain.RoleClient)
		if err != nil {
			return nil, err
		}
		userID = uid
		if lead.Phone != nil && *lead.Phone != "" {
			s.Store.UpdateUserProfile(domain.User{ID: uid, Phone: lead.Phone, FirstName: &lead.FullName})
		}
	}

	studentName := lead.FullName
	studentAge := lead.TargetAge
	if lead.RegType == "child" && lead.TargetName != nil {
		studentName = *lead.TargetName
	}
	studentID, err := s.Store.CreateStudent(domain.Student{
		FullName: studentName,
		Age:      studentAge,
		Level:    lead.TargetLevel,
		Phone:    lead.Phone,
		Status:   "active",
	})
	if err != nil {
		return nil, err
	}

	relation := domain.RelSelf
	if lead.RegType == "child" {
		relation = domain.RelParent
	}
	_, err = s.Store.CreateRelationship(domain.Relationship{
		UserID:    userID,
		StudentID: studentID,
		Relation:  relation,
	})
	if err != nil {
		return nil, err
	}

	if err := s.Store.ApproveLead(leadID, userID, studentID, reviewedBy); err != nil {
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
