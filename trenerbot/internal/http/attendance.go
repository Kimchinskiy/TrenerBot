package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"trenerbot/internal/domain"
	"trenerbot/internal/service"
)

// GET /api/attendance/date/{date} — returns clients with lessons on that date + attendance status.
func getDateAttendance(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	date := chi.URLParam(r, "date")
	if date == "" {
		writeError(w, http.StatusBadRequest, "date required")
		return
	}
	clients, err := svc.GetDateAttendance(date)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"date":    date,
		"clients": clients,
	})
}

// POST /api/attendance/date — batch save attendance for a specific date.
func saveDateAttendance(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	var body struct {
		Date    string                   `json:"date"`
		Entries []domain.DailyAttendance `json:"entries"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if body.Date == "" {
		writeError(w, http.StatusBadRequest, "date required")
		return
	}
	for i := range body.Entries {
		body.Entries[i].Date = body.Date
		if u != nil {
			body.Entries[i].MarkedBy = &u.ID
		}
	}
	if err := svc.SaveDateAttendance(body.Date, body.Entries); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
