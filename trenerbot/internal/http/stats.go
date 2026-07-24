package http

import (
	"net/http"
	"strconv"

	"trenerbot/internal/domain"
	"trenerbot/internal/service"
)

func statisticsHandler(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if u.Role != domain.RoleAdmin && u.Role != domain.RoleCoach {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	period := r.URL.Query().Get("period")
	if period == "" {
		period = "week"
	}
	if period != "week" && period != "month" && period != "year" {
		period = "week"
	}

	var coachID int64
	if u.Role == domain.RoleCoach {
		coach, err := svc.CoachByUser(u.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to get coach")
			return
		}
		if coach != nil {
			coachID = coach.ID
		}
	} else {
		queryCoach := r.URL.Query().Get("coach_id")
		if queryCoach != "" {
			coachID, _ = strconv.ParseInt(queryCoach, 10, 64)
		}
	}

	stats, err := svc.GetStatistics(service.StatisticsRequest{
		Period:  period,
		CoachID: coachID,
		IsAdmin: u.Role == domain.RoleAdmin,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get statistics: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, stats)
}
