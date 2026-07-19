package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"trenerbot/internal/domain"
	"trenerbot/internal/service"
)

func notificationsDue(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	ns, err := svc.ClaimDue(limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, ns)
}

func notificationsResult(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var body struct {
		OK bool `json:"ok"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if err := svc.MarkResult(id, body.OK); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func coachOrAdmin(svc *service.Services, u *domain.User) (int64, error) {
	if u.Role == domain.RoleAdmin {
		return 0, nil
	}
	coach, err := svc.CoachByUser(u.ID)
	if err != nil {
		return 0, err
	}
	if coach == nil {
		return 0, nil
	}
	return coach.ID, nil
}

// POST /api/notifications/preview — preview recipients for a coach broadcast.
func notificationsPreview(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	coachID, err := coachOrAdmin(svc, u)
	if err != nil || coachID == 0 && u.Role != domain.RoleAdmin {
		writeError(w, http.StatusForbidden, "coach only")
		return
	}
	var body struct {
		Filter    string  `json:"filter"`
		GroupID   int64   `json:"group_id"`
		ClientIDs []int64 `json:"client_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	recipients, err := svc.ListRecipients(coachID, body.Filter, body.GroupID, body.ClientIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total":      len(recipients),
		"recipients": recipients,
	})
}

// POST /api/notifications/send — send a coach broadcast notification.
func notificationsSend(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	coachID, err := coachOrAdmin(svc, u)
	if err != nil || coachID == 0 && u.Role != domain.RoleAdmin {
		writeError(w, http.StatusForbidden, "coach only")
		return
	}
	var body struct {
		Filter    string  `json:"filter"`
		GroupID   int64   `json:"group_id"`
		ClientIDs []int64 `json:"client_ids"`
		Title     string  `json:"title"`
		Text      string  `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if body.Text == "" {
		writeError(w, http.StatusBadRequest, "text required")
		return
	}
	result, err := svc.SendCoachNotification(coachID, body.Filter, body.GroupID, body.ClientIDs, body.Title, body.Text)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, result)
}
