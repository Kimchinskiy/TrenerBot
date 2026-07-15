package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"trenerbot/internal/config"
	"trenerbot/internal/service"
)

// webAppLogin validates the Telegram Mini App initData signature and issues a JWT.
// It is a public endpoint (like the bot login) and must NOT be placed inside the
// AuthMiddleware-protected /api group.
func webAppLogin(svc *service.Services, cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			InitData string `json:"init_data"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "bad json")
			return
		}
		if req.InitData == "" {
			writeError(w, http.StatusBadRequest, "init_data required")
			return
		}

		tgID, firstName, lastName, ok := validateInitData(req.InitData, cfg.BotToken)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid init_data")
			return
		}

		fullName := strings.TrimSpace(firstName + " " + lastName)
		u, client, tok, err := svc.TelegramLogin(service.TelegramLoginRequest{
			TelegramID: tgID,
			FullName:   fullName,
			Source:     "webapp",
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"user": u, "client": client, "token": tok})
	}
}

// validateInitData verifies the HMAC signature that Telegram attaches to Mini App launches.
// See https://core.telegram.org/bots/webapps#validating-data-received-via-the-mini-app
//
//   - secret_key = HMAC_SHA256("WebAppData", bot_token)
//   - data_check_string = key=value pairs sorted alphabetically by key, joined by "\n", excluding "hash"
//   - hash must equal HMAC_SHA256(secret_key, data_check_string) encoded as hex
func validateInitData(initData, botToken string) (telegramID, firstName, lastName string, ok bool) {
	params, err := url.ParseQuery(initData)
	if err != nil {
		return "", "", "", false
	}
	hash := params.Get("hash")
	if hash == "" {
		return "", "", "", false
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		if k == "hash" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var sb strings.Builder
	for i, k := range keys {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		sb.WriteString(params.Get(k))
	}
	dataCheckString := sb.String()

	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(botToken))
	secretKey := secret.Sum(nil)

	computed := hmac.New(sha256.New, secretKey)
	computed.Write([]byte(dataCheckString))
	computedHash := hex.EncodeToString(computed.Sum(nil))

	if !hmac.Equal([]byte(computedHash), []byte(hash)) {
		return "", "", "", false
	}

	authDate, err := strconv.ParseInt(params.Get("auth_date"), 10, 64)
	if err != nil {
		return "", "", "", false
	}
	if time.Now().Unix()-authDate > 24*3600 {
		return "", "", "", false
	}

	userField := params.Get("user")
	if userField == "" {
		return "", "", "", false
	}
	var u struct {
		ID        int64  `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Username  string `json:"username"`
		PhotoURL  string `json:"photo_url"`
	}
	if err := json.Unmarshal([]byte(userField), &u); err != nil {
		return "", "", "", false
	}
	return strconv.FormatInt(u.ID, 10), u.FirstName, u.LastName, true
}
