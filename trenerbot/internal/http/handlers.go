package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"trenerbot/internal/domain"
	"trenerbot/internal/service"
)

func authTelegram(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var req service.TelegramLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if req.TelegramID == "" {
		writeError(w, http.StatusBadRequest, "telegram_id required")
		return
	}
	u, client, tok, err := svc.TelegramLogin(req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": u, "client": client, "token": tok})
}

func listStudents(svc *service.Services, w http.ResponseWriter, _ *http.Request) {
	cs, err := svc.ListStudents()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, cs)
}

func clientsMe(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	resp := map[string]any{"role": u.Role, "user_id": u.ID, "telegram_id": u.TelegramID}
	if u.Role == domain.RoleParent {
		children, err := svc.ChildrenOfParent(u.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		resp["children"] = children
		writeJSON(w, http.StatusOK, resp)
		return
	}
	st, err := svc.Store.StudentByUserID(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	if st != nil {
		resp["client"] = st
	}
	writeJSON(w, http.StatusOK, resp)
}

func getClient(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	st, err := svc.GetStudent(id)
	if err != nil || st == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, st)
}

func updateClient(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var st domain.Student
	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	st.ID = id
	if err := svc.UpdateStudent(st); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func listCoaches(svc *service.Services, w http.ResponseWriter, _ *http.Request) {
	cs, err := svc.ListCoaches()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, cs)
}

func createCoach(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var co domain.Coach
	if err := json.NewDecoder(r.Body).Decode(&co); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if co.FullName == "" {
		writeError(w, http.StatusBadRequest, "full_name required")
		return
	}
	uid, err := svc.Store.CreateUser(nil, domain.RoleCoach)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	co.UserID = &uid
	id, err := svc.CreateCoach(co)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	co.ID = id
	writeJSON(w, http.StatusCreated, co)
}

func bindCoachTelegram(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		TelegramID string `json:"telegram_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.TelegramID == "" {
		writeError(w, http.StatusBadRequest, "telegram_id required")
		return
	}
	// fetch coach to get its user_id using repository method
	coach, err := svc.Store.CoachByID(id)
	if err != nil || coach == nil {
		writeError(w, http.StatusNotFound, "coach not found")
		return
	}
	if err := svc.SetTelegramID(*coach.UserID, body.TelegramID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func listLessons(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	if from == "" {
		from = time.Now().Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	}
	u := UserFrom(r.Context())
	switch u.Role {
	case domain.RoleClient:
		st, err := svc.Store.StudentByUserID(u.ID)
		if err != nil || st == nil {
			writeError(w, http.StatusNotFound, "client not found")
			return
		}
		lessons, err := svc.ClientSchedule(st.ID, from)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		writeJSON(w, http.StatusOK, lessons)
	case domain.RoleParent:
		children, err := svc.ChildrenOfParent(u.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		merged := map[int64]domain.Lesson{}
		for _, ch := range children {
			ls, err := svc.ClientSchedule(ch.ID, from)
			if err != nil {
				continue
			}
			for _, l := range ls {
				merged[l.ID] = l
			}
		}
		out := make([]domain.Lesson, 0, len(merged))
		for _, l := range merged {
			out = append(out, l)
		}
		writeJSON(w, http.StatusOK, out)
	case domain.RoleCoach:
		co, err := svc.CoachByUser(u.ID)
		if err != nil || co == nil {
			writeError(w, http.StatusNotFound, "coach not found")
			return
		}
		lessons, err := svc.ListWeekLessons(co.ID, from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		writeJSON(w, http.StatusOK, lessons)
	default: // admin
		lessons, err := svc.ListWeekLessons(0, from, to)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		writeJSON(w, http.StatusOK, lessons)
	}
}

func createLesson(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var l domain.Lesson
	if err := json.NewDecoder(r.Body).Decode(&l); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if l.Date == "" || l.Time == "" {
		writeError(w, http.StatusBadRequest, "date and time required")
		return
	}
	id, err := svc.CreateLesson(l)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func getLesson(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	l, err := svc.GetLesson(id)
	if err != nil || l == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, l)
}

func setLessonStatus(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		Status domain.LessonStatus `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Status == "" {
		writeError(w, http.StatusBadRequest, "status required")
		return
	}
	if err := svc.SetLessonStatus(id, body.Status); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func listAttendance(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	a, err := svc.ListAttendance(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, a)
}

func markAttendance(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ClientID int64 `json:"client_id"`
		Present  bool  `json:"present"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	u := UserFrom(r.Context())
	var markedBy *int64
	if u != nil {
		markedBy = &u.ID
	}
	if err := svc.MarkAttendance(lessonID, body.ClientID, body.Present, markedBy); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func registerClient(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ClientID int64 `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ClientID == 0 {
		writeError(w, http.StatusBadRequest, "client_id required")
		return
	}
	if err := svc.RegisterClientToLesson(lessonID, body.ClientID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func uploadFile(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "file too large")
		return
	}
	f, fh, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file required")
		return
	}
	defer f.Close()
	ownerType := r.FormValue("owner_type")
	ownerID, _ := strconv.ParseInt(r.FormValue("owner_id"), 10, 64)
	kind := r.FormValue("kind")
	if kind == "" {
		kind = "photo"
	}
	if ownerType == "" || ownerID == 0 {
		writeError(w, http.StatusBadRequest, "owner_type and owner_id required")
		return
	}
	dir := "data/uploads"
	_ = os.MkdirAll(dir, 0o755)
	dst := filepath.Join(dir, time.Now().Format("20060102150405")+"_"+filepath.Base(fh.Filename))
	out, err := os.Create(dst)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, f); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	id, err := svc.SaveFile(domain.File{OwnerType: ownerType, OwnerID: ownerID, Path: dst, Kind: kind})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "path": dst})
}

// GET /api/schedule — returns lesson entries for the current week range.
func scheduleHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	if from == "" || to == "" {
		writeError(w, http.StatusBadRequest, "from and to required")
		return
	}
	u := UserFrom(r.Context())
	var coachID, clientID int64
	if u.Role == domain.RoleCoach {
		co, err := svc.CoachByUser(u.ID)
		if err == nil && co != nil {
			coachID = co.ID
		}
	} else if u.Role == domain.RoleClient {
		st, err := svc.Store.StudentByUserID(u.ID)
		if err == nil && st != nil {
			clientID = st.ID
		}
	}
	entries, err := svc.GetSchedule(from, to, u.Role, u.ID, coachID, clientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, entries)
}

func getReport(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	from := q.Get("from")
	to := q.Get("to")
	if from == "" {
		from = time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	}
	if to == "" {
		to = time.Now().Format("2006-01-02")
	}
	rep, err := svc.Report(from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, rep)
}

// ---- Schedule Entry CRUD ----

func createScheduleEntry(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	co, _ := svc.CoachByUser(u.ID)
	var coachID *int64
	if co != nil {
		coachID = &co.ID
	}
	var body struct {
		ClientID int64  `json:"client_id"`
		GroupID  int64  `json:"group_id"`
		Date     string `json:"date"`
		Time     string `json:"time"`
		Duration int    `json:"duration"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if body.ClientID == 0 && body.GroupID == 0 {
		writeError(w, http.StatusBadRequest, "client_id or group_id required")
		return
	}
	if body.Duration <= 0 {
		body.Duration = 60
	}

	var createdIDs []int64
	if body.GroupID > 0 {
		members, err := svc.GetGroupStudents(body.GroupID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		if len(members) == 0 {
			writeError(w, http.StatusBadRequest, "group is empty")
			return
		}
		for _, m := range members {
			id, err := svc.Store.InsertLessonEntry(domain.LessonEntry{
				Date:      body.Date,
				Time:      body.Time,
				StudentID: m.StudentID,
				CoachID:   coachID,
				Duration:  body.Duration,
				Status:    domain.LessonPlanned,
				GroupID:   &body.GroupID,
			})
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal")
				return
			}
			createdIDs = append(createdIDs, id)
		}
	} else {
		id, err := svc.Store.InsertLessonEntry(domain.LessonEntry{
			Date:      body.Date,
			Time:      body.Time,
			StudentID: body.ClientID,
			CoachID:  coachID,
			Duration: body.Duration,
			Status:   domain.LessonPlanned,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		createdIDs = append(createdIDs, id)
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ids": createdIDs})
}

// ---- Waiting List ----

func waitingList(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	groupIDStr := q.Get("group_id")
	var groupID *int64
	if groupIDStr != "" {
		id, _ := strconv.ParseInt(groupIDStr, 10, 64)
		groupID = &id
	}
	list, err := svc.WaitingList(groupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func addToWaitingList(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClientID int64  `json:"client_id"`
		GroupID  *int64 `json:"group_id,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ClientID == 0 {
		writeError(w, http.StatusBadRequest, "client_id required")
		return
	}
	id, err := svc.AddToWaitingList(body.ClientID, body.GroupID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func removeFromWaitingList(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := svc.RemoveFromWaitingList(id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---- Lesson Notifications ----

func notifyLessonChange(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	lessonID, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ChangeType string `json:"change_type"` // time|location|canceled|coach
		OldValue   string `json:"old_value"`
		NewValue   string `json:"new_value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if err := svc.NotifyLessonChange(lessonID, body.ChangeType, body.OldValue, body.NewValue); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---- Debtors Widget ----

func debtorsWidget(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	days, _ := strconv.Atoi(q.Get("days"))
	if days == 0 {
		days = 30
	}
	debtors, err := svc.DebtorsWidget(days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, debtors)
}

// ---- Social Media ----

func socialMediaLinks(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u != nil {
		co, err := svc.CoachByUser(u.ID)
		if err == nil && co != nil {
			links := svc.GetSocialLinksMap(co.ID)
			if len(links) > 0 {
				writeJSON(w, http.StatusOK, links)
				return
			}
		}
	}
	writeJSON(w, http.StatusOK, svc.SocialMediaLinks())
}

func saveSocialLinks(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	co, err := svc.CoachByUser(u.ID)
	if err != nil || co == nil {
		writeError(w, http.StatusForbidden, "coach only")
		return
	}
	coachID := co.ID
	var body struct {
		Links []struct {
			Platform string  `json:"platform"`
			URL      *string `json:"url"`
			Enabled  bool    `json:"enabled"`
		} `json:"links"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	domainLinks := make([]domain.SocialLink, len(body.Links))
	for i, l := range body.Links {
		domainLinks[i] = domain.SocialLink{
			CoachID:  coachID,
			Platform: l.Platform,
			URL:      l.URL,
			Enabled:  l.Enabled,
		}
	}
	if err := svc.SaveSocialLinks(coachID, domainLinks); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---- New Client FAQ ----

func newClientFAQ(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := q.Get("q")
	writeJSON(w, http.StatusOK, map[string]string{"answer": svc.NewClientFAQ(query)})
}

// ---- Wellbeing Feedback ----

func submitWellbeing(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u.Role != domain.RoleClient {
		writeError(w, http.StatusForbidden, "clients only")
		return
	}
	st, err := svc.Store.StudentByUserID(u.ID)
	if err != nil || st == nil {
		writeError(w, http.StatusNotFound, "student not found")
		return
	}
	var body struct {
		LessonID int64  `json:"lesson_id"`
		Wellbeing int   `json:"wellbeing"` // 1-5
		Note      string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if err := svc.ClientWellbeingFeedback(st.ID, body.LessonID, body.Wellbeing, body.Note); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func wellbeingHistory(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	clientID, _ := strconv.ParseInt(chi.URLParam(r, "client_id"), 10, 64)
	history, err := svc.GetStudentWellbeingHistory(clientID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, history)
}

// ---- Admin Panel ----

func adminListClients(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit := 20
	offset := (page - 1) * limit

	students, err := svc.Store.ListStudents()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}

	// Add pagination
	total := len(students)
	start := offset
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"clients": students[start:end],
		"page":    page,
		"total":   total,
		"limit":   limit,
	})
}

func adminGrantBotAccess(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClientID        int64  `json:"client_id"`
		TelegramID      string `json:"telegram_id"`
		SubscriptionDays int   `json:"subscription_days"`
	}
	bodyBytes, _ := io.ReadAll(r.Body)
	slog.Debug("adminGrantBotAccess body", "body", string(bodyBytes))
	r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error("decode error", "err", err)
		writeError(w, http.StatusBadRequest, "invalid json: "+err.Error())
		return
	}
	if body.ClientID == 0 {
		writeError(w, http.StatusBadRequest, "client_id required")
		return
	}

	if body.SubscriptionDays == 0 {
		body.SubscriptionDays = 30
	}

	// Get student
	slog.Debug("adminGrantBotAccess looking for student", "client_id", body.ClientID)
	student, err := svc.Store.StudentByID(body.ClientID)
	if err != nil {
		slog.Error("adminGrantBotAccess StudentByID error", "err", err)
		writeError(w, http.StatusNotFound, "student not found: "+err.Error())
		return
	}
	if student == nil {
		slog.Debug("adminGrantBotAccess student is nil")
		writeError(w, http.StatusNotFound, "student not found")
		return
	}
	slog.Debug("adminGrantBotAccess found student", "id", student.ID, "name", student.FullName)

	// Update student with bot access
	student.BotAccess = true

	// Link telegram ID to user if exists
	if student.UserID != nil && body.TelegramID != "" {
		_, err := svc.Store.DB.Exec(`UPDATE users SET telegram_id = ? WHERE id = ?`, body.TelegramID, *student.UserID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to link telegram")
			return
		}
	}

	if err := svc.Store.UpdateStudent(*student); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update student")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func adminRevokeBotAccess(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var body struct {
		ClientID int64 `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ClientID == 0 {
		writeError(w, http.StatusBadRequest, "client_id required")
		return
	}

	student, err := svc.Store.StudentByID(body.ClientID)
	if err != nil || student == nil {
		writeError(w, http.StatusNotFound, "student not found")
		return
	}

	student.BotAccess = false

	if err := svc.Store.UpdateStudent(*student); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update student")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- Groups ----------

func listGroups(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	groups, err := svc.ListGroups()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

func getGroup(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	g, err := svc.GetGroup(id)
	if err != nil || g == nil {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	writeJSON(w, http.StatusOK, g)
}

func createGroup(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var g domain.Group
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if g.Name == nil || strings.TrimSpace(*g.Name) == "" {
		writeError(w, http.StatusBadRequest, "name required")
		return
	}
	if g.Active == 0 {
		g.Active = 1
	}
	id, err := svc.CreateGroup(g)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func updateGroup(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var g domain.Group
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	g.ID = id
	if err := svc.UpdateGroup(g); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func deleteGroup(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := svc.DeleteGroup(id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func groupClients(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	members, err := svc.GetGroupStudents(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, members)
}

func addStudentToGroup(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ClientID int64  `json:"client_id"`
		Role     string `json:"role"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ClientID == 0 {
		writeError(w, http.StatusBadRequest, "client_id required")
		return
	}
	if body.Role == "" {
		body.Role = "member"
	}
	if err := svc.AddStudentToGroup(id, body.ClientID, body.Role); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ---------- Student Subscriptions ----------

func listClientSubscriptions(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	studentID, _ := strconv.ParseInt(chi.URLParam(r, "client_id"), 10, 64)
	subs, err := svc.ClientSubscriptions(studentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, subs)
}

func createClientSubscription(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var sub domain.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if sub.StudentID == 0 || sub.Type == "" {
		writeError(w, http.StatusBadRequest, "client_id and type required")
		return
	}
	id, err := svc.CreateClientSubscription(sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]int64{"id": id})
}

func updateClientSubscription(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var sub domain.Subscription
	if err := json.NewDecoder(r.Body).Decode(&sub); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if sub.ID == 0 {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	if err := svc.UpdateClientSubscription(sub); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func deleteClientSubscription(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err := svc.DeleteClientSubscription(id); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func groupAvailableStudents(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	students, err := svc.StudentsNotInGroup(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, students)
}

func removeStudentFromGroup(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		ClientID int64 `json:"client_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ClientID == 0 {
		writeError(w, http.StatusBadRequest, "client_id required")
		return
	}
	if err := svc.RemoveStudentFromGroup(id, body.ClientID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
