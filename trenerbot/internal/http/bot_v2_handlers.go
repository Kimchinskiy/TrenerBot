package http

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"trenerbot/internal/domain"
	"trenerbot/internal/service"
)

// ---------- Leads ----------

func createLeadHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	var body struct {
		TelegramID  string  `json:"telegram_id"`
		FullName    string  `json:"full_name"`
		Phone       *string `json:"phone,omitempty"`
		TargetName  *string `json:"target_name,omitempty"`
		TargetAge   *int    `json:"target_age,omitempty"`
		TargetLevel string  `json:"target_level"`
		RegType     string  `json:"reg_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if body.TelegramID == "" || body.FullName == "" {
		writeError(w, http.StatusBadRequest, "telegram_id and full_name required")
		return
	}
	if body.RegType == "" {
		body.RegType = "self"
	}
	if body.TargetLevel == "" {
		body.TargetLevel = "beginner"
	}

	lead := domain.Lead{
		TelegramID:  body.TelegramID,
		FullName:    body.FullName,
		Phone:       body.Phone,
		TargetName:  body.TargetName,
		TargetAge:   body.TargetAge,
		TargetLevel: body.TargetLevel,
		RegType:     body.RegType,
	}

	id, err := svc.CreateLead(lead)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create lead: "+err.Error())
		return
	}

	go func() {
		coaches, _ := svc.Store.ListCoaches()
		for _, c := range coaches {
			if c.UserID != nil {
				_ = svc.Notify(*c.UserID, "new_lead", map[string]any{
					"lead_id":    id,
					"full_name":  body.FullName,
					"target":     body.TargetName,
					"age":        body.TargetAge,
					"level":      body.TargetLevel,
					"reg_type":   body.RegType,
					"created_at": time.Now().Format(time.RFC3339),
				}, time.Now())
			}
		}
	}()

	writeJSON(w, http.StatusCreated, map[string]any{"id": id})
}

func listLeadsHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil || (u.Role != domain.RoleAdmin && u.Role != domain.RoleCoach) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	leads, err := svc.PendingLeads()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, leads)
}

func reviewLeadHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil || (u.Role != domain.RoleAdmin && u.Role != domain.RoleCoach) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	leadID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var body struct {
		Action string `json:"action"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}

	switch body.Action {
	case "approve":
		lead, err := svc.ApproveLead(leadID, u.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, lead)
	case "reject":
		if err := svc.RejectLead(leadID, u.ID); err != nil {
			writeError(w, http.StatusInternalServerError, "failed: "+err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
	default:
		writeError(w, http.StatusBadRequest, "action must be approve or reject")
	}
}

// ---------- Students ----------

func myStudentsHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	students, err := svc.StudentsByUserID(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, students)
}

// ---------- Student Trainings ----------

func studentTrainingsHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	studentID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	if from == "" {
		from = time.Now().Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	}
	trainings, err := svc.StudentTrainings(studentID, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, trainings)
}

// ---------- Training Absence ----------

func reportAbsenceHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		TrainingID int64  `json:"training_id"`
		StudentID  int64  `json:"student_id"`
		Reason     string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TrainingID == 0 {
		writeError(w, http.StatusBadRequest, "training_id required")
		return
	}
	if err := svc.ReportAbsence(body.TrainingID, body.StudentID, body.Reason); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}

	go func() {
		coaches, _ := svc.Store.ListCoaches()
		for _, c := range coaches {
			if c.UserID != nil {
				_ = svc.Notify(*c.UserID, "absence_report", map[string]any{
					"training_id": body.TrainingID,
					"student_id":  body.StudentID,
					"reason":      body.Reason,
				}, time.Now())
			}
		}
	}()

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- Student Subscription ----------

func studentSubscriptionHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	studentID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	student, err := svc.StudentByID(studentID)
	if err != nil || student == nil {
		writeError(w, http.StatusNotFound, "student not found")
		return
	}
	if student.ClientID != nil {
		subs, _ := svc.ClientSubscriptions(*student.ClientID)
		if len(subs) > 0 {
			sub := subs[0]
			writeJSON(w, http.StatusOK, map[string]any{
				"type":         sub.Type,
				"price":        sub.Price,
				"lessons_left": sub.LessonsLeft,
				"ends_at":      sub.EndsAt,
			})
			return
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"type":         "",
		"price":        0,
		"lessons_left": 0,
		"ends_at":      "",
	})
}

// ---------- Group Students (for coach broadcast) ----------

func groupStudentsHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil || (u.Role != domain.RoleAdmin && u.Role != domain.RoleCoach) {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	groupID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	members, err := svc.GetGroupClients(groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, members)
}

// ---------- Notification Prefs ----------

func getNotifPrefsHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	studentID, _ := strconv.ParseInt(r.URL.Query().Get("student_id"), 10, 64)
	prefs, err := svc.GetNotifPrefs(u.ID, studentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	if prefs == nil {
		prefs = &domain.NotificationPref{
			UserID:       u.ID,
			ReminderDay:   1,
			ReminderHours: 2,
			LessonsLow:    1,
			SubExpiring:   3,
			News:          1,
		}
	}
	writeJSON(w, http.StatusOK, prefs)
}

func saveNotifPrefsHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var p domain.NotificationPref
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	p.UserID = u.ID
	if err := svc.SaveNotifPrefs(p); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
