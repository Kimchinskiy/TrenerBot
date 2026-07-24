package http

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"trenerbot/internal/config"
	"trenerbot/internal/domain"
	"trenerbot/internal/service"
)

// AuthMiddleware resolves the caller from either:
//   - X-Service-Token + X-Telegram-Id  (bot channel, ТЗ §2/§15)
//   - Authorization: Bearer <JWT>       (future web panel)
func AuthMiddleware(svc *service.Services, cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Add security headers
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:;")

			svcToken := r.Header.Get("X-Service-Token")
			if svcToken != "" && svcToken == cfg.ServiceToken {
				tgID := r.Header.Get("X-Telegram-Id")
				if tgID == "" {
					// system/bot account (e.g. polling the notification outbox)
					sys := &domain.User{ID: 0, Role: domain.RoleAdmin}
					next.ServeHTTP(w, r.WithContext(withUser(r.Context(), sys)))
					return
				}
				// Validate telegram ID format (basic validation)
				if len(tgID) < 5 || len(tgID) > 15 {
					writeError(w, http.StatusBadRequest, "invalid telegram_id")
					return
				}
				u, err := svc.Store.UserByTelegram(tgID)
				if err != nil {
					writeError(w, http.StatusInternalServerError, "internal")
					return
				}
				if u == nil {
					writeError(w, http.StatusUnauthorized, "unknown telegram user")
					return
				}
				next.ServeHTTP(w, r.WithContext(withUser(r.Context(), u)))
				return
			}

			authz := r.Header.Get("Authorization")
			if strings.HasPrefix(authz, "Bearer ") {
				tok := strings.TrimPrefix(authz, "Bearer ")
				claims, err := svc.Tokens.Parse(tok)
				if err != nil {
					writeError(w, http.StatusUnauthorized, "invalid token")
					return
				}
				u, err := svc.Store.UserByID(claims.UserID)
				if err != nil || u == nil {
					writeError(w, http.StatusUnauthorized, "unknown user")
					return
				}
				next.ServeHTTP(w, r.WithContext(withUser(r.Context(), u)))
				return
			}

			writeError(w, http.StatusUnauthorized, "unauthorized")
		})
	}
}

// Rate limiting middleware to prevent abuse
func RateLimiter(maxRequests int, window time.Duration) func(http.Handler) http.Handler {
	// Simple in-memory rate limiter per IP
	type client struct {
		count     int
		resetTime  time.Time
	}
	clients := make(map[string]*client)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr
			now := time.Now()

			// Clean old clients
			for ip, c := range clients {
				if now.After(c.resetTime) {
					delete(clients, ip)
				}
			}

			// Get or create client
			c, exists := clients[ip]
			if !exists {
				clients[ip] = &client{count: 1, resetTime: now.Add(window)}
				c = clients[ip]
			}

			// Check if rate limit exceeded
			if c.count >= maxRequests {
				slog.Warn("rate limit exceeded", "ip", ip, "count", c.count)
				writeError(w, http.StatusTooManyRequests, "too many requests")
				return
			}

			// Increment count
			c.count++
			next.ServeHTTP(w, r)
		})
	}
}

// serviceTokenOnly requires a valid X-Service-Token and sets a system caller. Used for open
// endpoints like registration where the user does not exist yet.
func serviceTokenOnly(cfg *config.Config, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Service-Token") != cfg.ServiceToken {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r.WithContext(withUser(r.Context(), &domain.User{ID: 0, Role: domain.RoleClient})))
	}
}
func guard(svc *service.Services, roles []domain.Role, h handlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := UserFrom(r.Context())
		if u == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		ok := false
		for _, role := range roles {
			if u.Role == role {
				ok = true
				break
			}
		}
		if !ok {
			writeError(w, http.StatusForbidden, "forbidden")
			return
		}
		h(svc, w, r)
	}
}

type handlerFunc func(svc *service.Services, w http.ResponseWriter, r *http.Request)

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)
		slog.Info("http",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"dur", time.Since(start).String(),
		)
	})
}

// Router builds the chi router with all MVP endpoints.
func Router(svc *service.Services, cfg *config.Config) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(Logger)
	r.Use(corsAllow)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Registration is open: only the service token is required (the user is created here).
	r.Post("/api/auth/telegram", serviceTokenOnly(cfg, func(w http.ResponseWriter, r *http.Request) {
		authTelegram(svc, w, r)
	}))

	// ---- Website authentication (primary product). Public endpoints. ----
	// optionalAuth lets provider logins also act as "link to my account" when a token is sent.
	r.Group(func(r chi.Router) {
		r.Use(optionalAuth(svc))
		r.Post("/api/auth/register", registerByPhone(svc))
		r.Post("/api/auth/login", loginByPhone(svc))
		r.Post("/api/auth/refresh", refreshTokens(svc))
		r.Post("/api/auth/logout", logout(svc))
		r.Post("/api/auth/telegram-widget", telegramWidgetLogin(svc, cfg))
		r.Post("/api/auth/max", maxLogin(svc, cfg))
		// Telegram Mini App login: additional method, validated via initData signature.
		r.Post("/api/auth/telegram-webapp", webAppLogin(svc, cfg))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(AuthMiddleware(svc, cfg))

		// Auth
		r.Post("/auth/telegram", func(w http.ResponseWriter, r *http.Request) { authTelegram(svc, w, r) })

		// Current account (single User entity) + login-method management.
		r.Get("/me", func(w http.ResponseWriter, r *http.Request) { getMe(svc, w, r) })
		r.Post("/auth/link/telegram", linkTelegramWidget(svc, cfg))
		r.Post("/auth/set-password", setPassword(svc))

		// Clients
		r.Get("/clients", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, listClients))
		r.Get("/clients/me", func(w http.ResponseWriter, r *http.Request) { clientsMe(svc, w, r) })
		r.Get("/clients/{id}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, getClient))
		r.Put("/clients/{id}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach, domain.RoleClient}, updateClient))

		// Coaches
		r.Get("/coaches", func(w http.ResponseWriter, r *http.Request) { listCoaches(svc, w, r) })
		r.Post("/coaches", guard(svc, []domain.Role{domain.RoleAdmin}, createCoach))
		r.Patch("/coaches/{id}/telegram", guard(svc, []domain.Role{domain.RoleAdmin}, bindCoachTelegram))

		// Lessons
		r.Get("/lessons", func(w http.ResponseWriter, r *http.Request) { listLessons(svc, w, r) })
		r.Post("/lessons", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, createLesson))
		r.Get("/lessons/{id}", func(w http.ResponseWriter, r *http.Request) { getLesson(svc, w, r) })
		r.Patch("/lessons/{id}/status", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, setLessonStatus))
		r.Get("/lessons/{id}/attendance", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, listAttendance))
		r.Post("/lessons/{id}/attendance", guard(svc, []domain.Role{domain.RoleCoach}, markAttendance))
		r.Post("/lessons/{id}/register", func(w http.ResponseWriter, r *http.Request) { registerClient(svc, w, r) })

		// Schedule (new model — lesson entries per athlete)
		r.Get("/schedule", func(w http.ResponseWriter, r *http.Request) { scheduleHandler(svc, w, r) })
		r.Post("/schedule", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, createScheduleEntry))

		// Files
		r.Post("/files", func(w http.ResponseWriter, r *http.Request) { uploadFile(svc, w, r) })

		// Notifications dispatch (bot polls these via service-token)
		r.Get("/notifications/due", func(w http.ResponseWriter, r *http.Request) { notificationsDue(svc, w, r) })
		r.Post("/notifications/{id}/result", func(w http.ResponseWriter, r *http.Request) { notificationsResult(svc, w, r) })

		// Coach broadcast notifications
		r.Post("/notifications/preview", func(w http.ResponseWriter, r *http.Request) { notificationsPreview(svc, w, r) })
		r.Post("/notifications/send", func(w http.ResponseWriter, r *http.Request) { notificationsSend(svc, w, r) })

		// Client -> coach contact (ТЗ §4)
		r.Post("/messages/coach", func(w http.ResponseWriter, r *http.Request) { messageCoach(svc, w, r) })

		// Reports
		r.Get("/reports", guard(svc, []domain.Role{domain.RoleAdmin}, getReport))

		// Admin panel
		r.Get("/admin/clients", guard(svc, []domain.Role{domain.RoleAdmin}, adminListClients))
		r.Post("/admin/clients/grant", guard(svc, []domain.Role{domain.RoleAdmin}, adminGrantBotAccess))
		r.Post("/admin/clients/revoke", guard(svc, []domain.Role{domain.RoleAdmin}, adminRevokeBotAccess))

		// Waiting List
		r.Get("/waiting-list", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, waitingList))
		r.Post("/waiting-list", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, addToWaitingList))
		r.Delete("/waiting-list/{id}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, removeFromWaitingList))

		// Invite codes for parent binding
		r.Post("/coach/invite-code", guard(svc, []domain.Role{domain.RoleCoach}, createInviteCode))

		// Lesson notifications
		r.Post("/lessons/{id}/notify", guard(svc, []domain.Role{domain.RoleCoach}, notifyLessonChange))

		// Debtors widget
		r.Get("/debtors", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, debtorsWidget))

		// Social media links
		r.Get("/social-media", func(w http.ResponseWriter, r *http.Request) { socialMediaLinks(svc, w, r) })
		r.Post("/social-media", func(w http.ResponseWriter, r *http.Request) { saveSocialLinks(svc, w, r) })

		// New client FAQ
		r.Get("/faq", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach, domain.RoleClient}, newClientFAQ))

		// Daily attendance (date-based)
		r.Get("/attendance/date/{date}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, getDateAttendance))
		r.Post("/attendance/date", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, saveDateAttendance))

		// Wellbeing feedback
		r.Post("/wellbeing", guard(svc, []domain.Role{domain.RoleClient}, submitWellbeing))
		r.Get("/wellbeing/{client_id}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach, domain.RoleClient}, wellbeingHistory))

		// Coach subscription & onboarding
		r.Get("/coach/onboarding", func(w http.ResponseWriter, r *http.Request) { getCoachOnboarding(svc, w, r) })
		r.Post("/coach/upgrade", func(w http.ResponseWriter, r *http.Request) { upgradeToCoach(svc, w, r) })
		r.Get("/coach/subscription", func(w http.ResponseWriter, r *http.Request) { getCoachSubscription(svc, w, r) })
		r.Post("/coach/subscription/trial", func(w http.ResponseWriter, r *http.Request) { startCoachTrial(svc, w, r) })
		r.Post("/coach/subscription/activate", func(w http.ResponseWriter, r *http.Request) { activateCoachSubscription(svc, w, r) })

		// Parent features
		r.Post("/parent/upgrade", func(w http.ResponseWriter, r *http.Request) { upgradeToParent(svc, w, r) })
		r.Post("/parent/link", func(w http.ResponseWriter, r *http.Request) { linkChild(svc, w, r) })
		r.Get("/parent/children/status", func(w http.ResponseWriter, r *http.Request) { getChildrenStatus(svc, w, r) })
		r.Get("/parent/notif-prefs", func(w http.ResponseWriter, r *http.Request) { getParentNotifPrefs(svc, w, r) })
		r.Post("/parent/notif-prefs", func(w http.ResponseWriter, r *http.Request) { saveParentNotifPrefs(svc, w, r) })

		// Groups
		r.Get("/groups", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, listGroups))
		r.Get("/groups/{id}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, getGroup))
		r.Post("/groups", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, createGroup))
		r.Put("/groups/{id}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, updateGroup))
		r.Delete("/groups/{id}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, deleteGroup))
		r.Get("/groups/{id}/clients", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, groupClients))
		r.Post("/groups/{id}/clients", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, addClientToGroup))
		r.Delete("/groups/{id}/clients", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, removeClientFromGroup))
	})

	// Client subscriptions
	r.Route("/clients/{client_id}/subscriptions", func(r chi.Router) {
		r.Get("/", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, listClientSubscriptions))
		r.Post("/", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, createClientSubscription))
		r.Put("/", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, updateClientSubscription))
		r.Delete("/{id}", guard(svc, []domain.Role{domain.RoleAdmin, domain.RoleCoach}, deleteClientSubscription))
	})

	return r
}

func corsAllow(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Service-Token, X-Telegram-Id")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}
