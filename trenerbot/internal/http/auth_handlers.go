package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"trenerbot/internal/config"
	"trenerbot/internal/domain"
	"trenerbot/internal/service"
)

// currentUserID returns the authenticated user's id from a Bearer token if present,
// or 0 when the request is anonymous. Used by public auth endpoints that also support
// "link this provider to my current account" flows.
func currentUserID(r *http.Request) int64 {
	u := UserFrom(r.Context())
	if u != nil {
		return u.ID
	}
	return 0
}

// optionalAuth resolves a Bearer token if provided, without rejecting anonymous
// requests and without an extra DB round-trip: it only populates the caller id/role
// from the verified JWT claims. Endpoints that need the full user record live in the
// AuthMiddleware-protected group instead.
func optionalAuth(svc *service.Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authz := r.Header.Get("Authorization")
			if strings.HasPrefix(authz, "Bearer ") {
				tok := strings.TrimPrefix(authz, "Bearer ")
				if claims, err := svc.Tokens.Parse(tok); err == nil && claims.UserID != 0 {
					r = r.WithContext(withUser(r.Context(), &domain.User{ID: claims.UserID, Role: claims.Role}))
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func writeAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrPhoneTaken):
		writeError(w, http.StatusConflict, "Этот номер уже зарегистрирован")
	case errors.Is(err, service.ErrInvalidCreds):
		writeError(w, http.StatusUnauthorized, "Неверный номер телефона или пароль")
	case errors.Is(err, service.ErrWeakPassword):
		writeError(w, http.StatusBadRequest, "Пароль слишком короткий (минимум 6 символов)")
	case errors.Is(err, service.ErrInvalidPhone):
		writeError(w, http.StatusBadRequest, "Некорректный номер телефона")
	case errors.Is(err, service.ErrProviderTaken):
		writeError(w, http.StatusConflict, "Этот аккаунт уже привязан к другому пользователю")
	case errors.Is(err, service.ErrProviderDisabled):
		writeError(w, http.StatusNotImplemented, "Способ входа временно недоступен")
	default:
		writeError(w, http.StatusInternalServerError, "internal")
	}
}

// POST /api/auth/register — phone + password.
func registerByPhone(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Phone     string `json:"phone"`
			Password  string `json:"password"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "bad json")
			return
		}
		tokens, err := svc.RegisterByPhone(body.Phone, body.Password, body.FirstName, body.LastName)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeJSON(w, http.StatusCreated, tokens)
	}
}

// POST /api/auth/login — phone + password.
func loginByPhone(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Phone    string `json:"phone"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "bad json")
			return
		}
		tokens, err := svc.LoginByPhone(body.Phone, body.Password)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tokens)
	}
}

// POST /api/auth/refresh — rotate refresh token, issue a new pair.
func refreshTokens(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
			writeError(w, http.StatusBadRequest, "refresh_token required")
			return
		}
		tokens, err := svc.Refresh(body.RefreshToken)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tokens)
	}
}

// POST /api/auth/logout — revoke a refresh token.
func logout(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			RefreshToken string `json:"refresh_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body.RefreshToken != "" {
			_ = svc.Logout(body.RefreshToken)
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// POST /api/auth/telegram — Telegram Login Widget (website login, not Mini App).
// Accepts the flat field map produced by the widget. If a Bearer token is present,
// the Telegram identity is linked to that account instead of creating a new one.
func telegramWidgetLogin(svc *service.Services, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var fields map[string]string
		if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
			writeError(w, http.StatusBadRequest, "bad json")
			return
		}
		tgUser, ok := validateWidgetData(fields, cfg.BotToken)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid telegram signature")
			return
		}
		tokens, err := svc.LoginWithProvider(service.ProviderProfile{
			Provider:   "telegram",
			ExternalID: tgUser.ID,
			FirstName:  tgUser.FirstName,
			LastName:   tgUser.LastName,
			AvatarURL:  tgUser.PhotoURL,
		}, currentUserID(r))
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tokens)
	}
}

// POST /api/auth/max — MAX OAuth login. Architecture placeholder: the provider
// pipeline is fully wired, only the token verification against MAX is pending.
// Once MAX auth is available, implement verifyMaxToken and remove the guard.
func maxLogin(svc *service.Services, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Token string `json:"token"` // MAX OAuth token / auth code
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "bad json")
			return
		}
		profile, err := verifyMaxToken(cfg, body.Token)
		if err != nil {
			writeAuthError(w, err)
			return
		}
		tokens, err := svc.LoginWithProvider(*profile, currentUserID(r))
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, tokens)
	}
}

// verifyMaxToken validates a MAX auth token and returns the normalized profile.
// TODO: implement against the official MAX OAuth API when credentials are available.
func verifyMaxToken(_ *config.Config, _ string) (*service.ProviderProfile, error) {
	return nil, service.ErrProviderDisabled
}

// GET /api/me — the authenticated account with its enabled login methods.
func getMe(svc *service.Services, w http.ResponseWriter, r *http.Request) {
	u := UserFrom(r.Context())
	if u == nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"user":         u,
		"auth_methods": u.AuthMethods(),
	})
}

// POST /api/auth/link/telegram — attach a Telegram identity to the current account.
func linkTelegramWidget(svc *service.Services, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := UserFrom(r.Context())
		if u == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var fields map[string]string
		if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
			writeError(w, http.StatusBadRequest, "bad json")
			return
		}
		tgUser, ok := validateWidgetData(fields, cfg.BotToken)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid telegram signature")
			return
		}
		err := svc.LinkProviderToUser(u.ID, service.ProviderProfile{
			Provider:   "telegram",
			ExternalID: tgUser.ID,
			FirstName:  tgUser.FirstName,
			LastName:   tgUser.LastName,
			AvatarURL:  tgUser.PhotoURL,
		})
		if err != nil {
			writeAuthError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

// POST /api/auth/set-password — add or change the password for the current account
// (e.g. a Telegram-first user enabling phone+password login).
func setPassword(svc *service.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := UserFrom(r.Context())
		if u == nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		var body struct {
			Phone    string `json:"phone"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "bad json")
			return
		}
		if err := svc.SetAccountPassword(u.ID, body.Phone, body.Password); err != nil {
			writeAuthError(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
