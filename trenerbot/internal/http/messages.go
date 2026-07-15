package http

import (
	"encoding/json"
	"net/http"

	"trenerbot/internal/service"
)

func messageCoach(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	var body struct {
		From string `json:"from"`
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Text == "" {
		writeError(w, http.StatusBadRequest, "text required")
		return
	}
	if err := svc.MessageCoaches(body.From, body.Text); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
