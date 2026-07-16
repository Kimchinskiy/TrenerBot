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

// webAppLogin validates the Telegram Mini App initData signature and issues a token pair.
// This remains an ADDITIONAL login method (the site works standalone in any browser);
// it resolves to the same single User entity via the provider architecture.
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

		tokens, err := svc.LoginWithProvider(service.ProviderProfile{
			Provider:   "telegram",
			ExternalID: tgID,
			FirstName:  firstName,
			LastName:   lastName,
		}, currentUserID(r))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal")
			return
		}
		writeJSON(w, http.StatusOK, tokens)
	}
}

// TelegramWidgetUser is the identity extracted from a validated Telegram Login Widget payload.
type TelegramWidgetUser struct {
	ID        string
	FirstName string
	LastName  string
	Username  string
	PhotoURL  string
}

// validateWidgetData verifies the hash from the Telegram Login Widget (website login,
// https://core.telegram.org/widgets/login#checking-authorization). Unlike Mini App
// initData, the secret key here is SHA256(bot_token) and fields arrive as a flat map.
func validateWidgetData(fields map[string]string, botToken string) (*TelegramWidgetUser, bool) {
	hash := fields["hash"]
	if hash == "" {
		return nil, false
	}
	keys := make([]string, 0, len(fields))
	for k := range fields {
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
		sb.WriteString(fields[k])
	}

	secretKey := sha256.Sum256([]byte(botToken))
	mac := hmac.New(sha256.New, secretKey[:])
	mac.Write([]byte(sb.String()))
	computed := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(computed), []byte(hash)) {
		return nil, false
	}
	if ad, err := strconv.ParseInt(fields["auth_date"], 10, 64); err != nil || time.Now().Unix()-ad > 5*60 {
		return nil, false
	}
	return &TelegramWidgetUser{
		ID:        fields["id"],
		FirstName: fields["first_name"],
		LastName:  fields["last_name"],
		Username:  fields["username"],
		PhotoURL:  fields["photo_url"],
	}, true
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
