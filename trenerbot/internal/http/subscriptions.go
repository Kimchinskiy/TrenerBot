package http

import (
	"encoding/json"
	"net/http"
	"time"

	"trenerbot/internal/domain"
	"trenerbot/internal/service"
)

// POST /api/coach/upgrade — upgrade current user to coach role.
func upgradeToCoach(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
		Sport    string `json:"sport"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if body.FullName == "" {
		writeError(w, http.StatusBadRequest, "full_name required")
		return
	}
	// Update user role
	if err := svc.Store.UpdateUserRole(u.ID, domain.RoleCoach); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	// Create coach profile
	co := domain.Coach{
		UserID:   &u.ID,
		FullName: body.FullName,
		Sport:    strPtr(body.Sport),
	}
	coachID, err := svc.CreateCoach(co)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	// Create trial subscription
	_, err = svc.GetOrCreateCoachSubscription(coachID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"role":      "coach",
		"coach_id":  coachID,
	})
}

// GET /api/coach/subscription — get coach subscription status.
func getCoachSubscription(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	co, err := svc.CoachByUser(u.ID)
	if err != nil || co == nil {
		writeError(w, http.StatusNotFound, "coach not found")
		return
	}
	sub, err := svc.GetOrCreateCoachSubscription(co.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	active := svc.IsCoachSubscriptionActive(co.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"subscription": sub,
		"active":       active,
	})
}

// POST /api/coach/subscription/activate — activate subscription (payment).
func activateCoachSubscription(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	co, err := svc.CoachByUser(u.ID)
	if err != nil || co == nil {
		writeError(w, http.StatusNotFound, "coach not found")
		return
	}
	var body struct {
		Days int `json:"days"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Days <= 0 {
		body.Days = 30
	}
	if err := svc.Store.ExtendCoachSubscription(co.ID, body.Days); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	sub, _ := svc.GetCoachSubscription(co.ID)
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"subscription": sub,
	})
}

// ---------- Parent handlers ----------

// POST /api/parent/upgrade — upgrade current user to parent role.
func upgradeToParent(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := svc.Store.UpdateUserRole(u.ID, domain.RoleParent); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
		"role":   "parent",
	})
}

// POST /api/parent/link — link parent to child by birth date verification.
func linkChild(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	// Ensure user is a parent
	if u.Role != domain.RoleParent {
		if err := svc.Store.UpdateUserRole(u.ID, domain.RoleParent); err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
	}
	var body struct {
		FullName  string `json:"full_name"`
		BirthDate string `json:"birth_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	if body.FullName == "" || body.BirthDate == "" {
		writeError(w, http.StatusBadRequest, "full_name and birth_date required")
		return
	}
	client, err := svc.FindChildByBirthDate(body.FullName, body.BirthDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	if client == nil {
		writeError(w, http.StatusNotFound, "Ребёнок с такими данными не найден")
		return
	}
	if err := svc.LinkParentToChild(u.ID, client.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"child_id":  client.ID,
		"child_name": client.FullName,
	})
}

// GET /api/parent/children/status — get lesson status for all children.
func getChildrenStatus(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	statuses, err := svc.GetChildrenLessonStatuses(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, statuses)
}

// GET /api/parent/notif-prefs — get notification preferences.
func getParentNotifPrefs(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	prefs, err := svc.GetParentNotifPrefs(u.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

// POST /api/parent/notif-prefs — save notification preferences.
func saveParentNotifPrefs(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	var body struct {
		ChildID     int64 `json:"child_id"`
		LessonStart *bool `json:"lesson_start,omitempty"`
		LessonEnd15 *bool `json:"lesson_end_15,omitempty"`
		LessonMissed *bool `json:"lesson_missed,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "bad json")
		return
	}
	pref := domain.ParentNotifPref{
		ParentUserID: u.ID,
		ChildID:      body.ChildID,
	}
	if body.LessonStart != nil {
		pref.LessonStart = *body.LessonStart
	} else {
		pref.LessonStart = true
	}
	if body.LessonEnd15 != nil {
		pref.LessonEnd15 = *body.LessonEnd15
	} else {
		pref.LessonEnd15 = true
	}
	if body.LessonMissed != nil {
		pref.LessonMissed = *body.LessonMissed
	} else {
		pref.LessonMissed = true
	}
	if err := svc.SaveParentNotifPrefs(pref); err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /api/coach/onboarding — get onboard state for coach.
func getCoachOnboarding(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	co, err := svc.CoachByUser(u.ID)
	if err != nil || co == nil {
		// Not a coach yet - return onboarding info
		writeJSON(w, http.StatusOK, map[string]any{
			"is_coach":  false,
			"message":   "До сих пор ведете учет в заметках или Excel? Платформа Плавли создана от тренеров для тренеров.",
		})
		return
	}
	sub, err := svc.GetCoachSubscription(co.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	active := svc.IsCoachSubscriptionActive(co.ID)
	resp := map[string]any{
		"is_coach": true,
		"coach":    co,
		"active":   active,
	}
	if sub != nil {
		resp["subscription"] = sub
		daysLeft := 0
		if sub.Status == domain.SubTrial && sub.TrialEnd != nil {
			if t, err := time.Parse("2006-01-02 15:04:05", *sub.TrialEnd); err == nil {
				daysLeft = int(time.Until(t).Hours() / 24)
			}
		} else if sub.Status == domain.SubActive && sub.PaidUntil != nil {
			if t, err := time.Parse("2006-01-02 15:04:05", *sub.PaidUntil); err == nil {
				daysLeft = int(time.Until(t).Hours() / 24)
			}
		}
		resp["days_left"] = daysLeft
	}
	writeJSON(w, http.StatusOK, resp)
}

// POST /api/coach/subscription/trial — start trial (idempotent).
func startCoachTrial(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	co, err := svc.CoachByUser(u.ID)
	if err != nil || co == nil {
		writeError(w, http.StatusNotFound, "coach not found")
		return
	}
	sub, err := svc.GetOrCreateCoachSubscription(co.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":       "ok",
		"subscription": sub,
	})
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
