package service

import (
	"errors"
	"regexp"
	"strings"
	"time"

	"trenerbot/internal/auth"
	"trenerbot/internal/domain"
)

// Auth errors surfaced to the HTTP layer.
var (
	ErrPhoneTaken       = errors.New("phone already registered")
	ErrInvalidCreds     = errors.New("invalid phone or password")
	ErrWeakPassword     = errors.New("password too short")
	ErrInvalidPhone     = errors.New("invalid phone")
	ErrProviderTaken    = errors.New("provider already linked to another account")
	ErrProviderDisabled = errors.New("provider not available")
)

// AuthTokens is the token pair returned to clients after a successful auth.
type AuthTokens struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         *domain.User `json:"user"`
}

var phoneRe = regexp.MustCompile(`\D`)

// normalizePhone strips everything but digits and a leading '+', so "+7 (999) 12"
// and "89991234567" collapse to a canonical form.
func normalizePhone(raw string) string {
	raw = strings.TrimSpace(raw)
	hasPlus := strings.HasPrefix(raw, "+")
	digits := phoneRe.ReplaceAllString(raw, "")
	if hasPlus {
		return "+" + digits
	}
	// Russian numbers: normalize leading 8 to +7 for consistency.
	if len(digits) == 11 && strings.HasPrefix(digits, "8") {
		return "+7" + digits[1:]
	}
	return "+" + digits
}

func isValidPhone(phone string) bool {
	d := strings.TrimPrefix(phone, "+")
	return len(d) >= 10 && len(d) <= 15
}

// issueTokens creates an access JWT + a persisted refresh token for the user.
func (s *Services) issueTokens(u *domain.User) (*AuthTokens, error) {
	access, err := s.Tokens.GenerateAccess(u.ID, u.Role)
	if err != nil {
		return nil, err
	}
	refresh, hash, err := auth.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	if err := s.Store.InsertRefreshToken(u.ID, hash, time.Now().Add(s.Tokens.RefreshTTL())); err != nil {
		return nil, err
	}
	return &AuthTokens{AccessToken: access, RefreshToken: refresh, User: u}, nil
}

// RegisterByPhone creates a new account (phone + password). If a client profile was
// already created for this phone (e.g. added by an admin/coach), it links to that
// existing account instead of creating a duplicate.
func (s *Services) RegisterByPhone(phoneRaw, password, firstName, lastName string) (*AuthTokens, error) {
	phone := normalizePhone(phoneRaw)
	if !isValidPhone(phone) {
		return nil, ErrInvalidPhone
	}
	if len(password) < 6 {
		return nil, ErrWeakPassword
	}
	existing, err := s.Store.UserByPhone(phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		// A user row with this phone already exists. Never authenticate into it during
		// registration: even a passwordless account (Telegram-only / admin-created /
		// migration-backfilled) must not be silently taken over by whoever knows the
		// number. Claiming such an account requires the authenticated link flow /
		// SetAccountPassword, not open registration.
		return nil, ErrPhoneTaken
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	u := domain.User{
		Phone:        &phone,
		PasswordHash: &hash,
		FirstName:    nullable(firstName),
		LastName:     nullable(lastName),
		Role:         domain.RoleClient,
	}
	uid, err := s.Store.InsertUser(u)
	if err != nil {
		return nil, err
	}
	// Create the linked student profile (mirrors TelegramLogin behaviour).
	full := strings.TrimSpace(firstName + " " + lastName)
	_, _ = s.Store.CreateStudentFull(domain.Student{
		UserID:   &uid,
		FullName: full,
		Phone:    &phone,
		Status:   "active",
		Source:   nullable("web"),
	})
	created, err := s.Store.UserByID(uid)
	if err != nil {
		return nil, err
	}
	_ = s.notifyCoaches("new_client", map[string]any{"user_id": uid, "name": full})
	return s.issueTokens(created)
}

// LoginByPhone authenticates with phone + password.
func (s *Services) LoginByPhone(phoneRaw, password string) (*AuthTokens, error) {
	phone := normalizePhone(phoneRaw)
	u, err := s.Store.UserByPhone(phone)
	if err != nil {
		return nil, err
	}
	if u == nil || !u.HasPassword() || !auth.CheckPassword(*u.PasswordHash, password) {
		return nil, ErrInvalidCreds
	}
	return s.issueTokens(u)
}

// Refresh rotates a refresh token: the old one is revoked and a new pair is issued.
func (s *Services) Refresh(refreshToken string) (*AuthTokens, error) {
	hash := auth.HashToken(refreshToken)
	userID, err := s.Store.RefreshTokenUser(hash)
	if err != nil {
		return nil, err
	}
	if userID == 0 {
		return nil, ErrInvalidCreds
	}
	u, err := s.Store.UserByID(userID)
	if err != nil || u == nil {
		return nil, ErrInvalidCreds
	}
	_ = s.Store.RevokeRefreshToken(hash) // rotate
	return s.issueTokens(u)
}

// Logout revokes a specific refresh token.
func (s *Services) Logout(refreshToken string) error {
	return s.Store.RevokeRefreshToken(auth.HashToken(refreshToken))
}

// ---------- External login providers (Telegram, MAX, ...) ----------

// ProviderProfile is the normalized identity returned by an auth provider.
type ProviderProfile struct {
	Provider   string // "telegram" | "max"
	ExternalID string
	FirstName  string
	LastName   string
	AvatarURL  string
	Phone      string // optional, if the provider shares it
}

// userByProvider looks up an account already linked to this provider identity.
func (s *Services) userByProvider(p ProviderProfile) (*domain.User, error) {
	switch p.Provider {
	case "telegram":
		return s.Store.UserByTelegram(p.ExternalID)
	case "max":
		return s.Store.UserByMaxID(p.ExternalID)
	default:
		return nil, ErrProviderDisabled
	}
}

func (s *Services) linkProvider(userID int64, p ProviderProfile) error {
	switch p.Provider {
	case "telegram":
		return s.Store.LinkTelegram(userID, p.ExternalID)
	case "max":
		return s.Store.LinkMax(userID, p.ExternalID)
	default:
		return ErrProviderDisabled
	}
}

// LoginWithProvider logs a user in via an external provider. Behaviour:
//   - if the provider identity is already linked -> log that user in;
//   - else if a currentUserID is supplied (linking flow) -> attach the provider
//     to that existing account;
//   - else create a fresh account for this provider identity.
//
// This guarantees Telegram/MAX and password all resolve to a single User entity.
func (s *Services) LoginWithProvider(p ProviderProfile, currentUserID int64) (*AuthTokens, error) {
	if p.ExternalID == "" {
		return nil, ErrInvalidCreds
	}
	existing, err := s.userByProvider(p)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return s.issueTokens(existing)
	}

	// Linking to the currently authenticated account. Reject if the provider identity
	// is already linked to a DIFFERENT account (prevents hijacking someone else's
	// Telegram/MAX identity onto the attacker's account).
	if currentUserID != 0 {
		if err := s.LinkProviderToUser(currentUserID, p); err != nil {
			return nil, err
		}
		u, err := s.Store.UserByID(currentUserID)
		if err != nil || u == nil {
			return nil, ErrInvalidCreds
		}
		return s.issueTokens(u)
	}

	// No silent phone-based auto-linking: a provider that shares a phone must not be
	// merged into an existing account without explicit, authenticated user consent
	// (use the /auth/link flow). Otherwise a phone collision could attach an
	// attacker-controlled identity to a stranger's account.

	// New account for this provider identity.
	u := domain.User{
		TelegramID: providerID(p, "telegram"),
		MaxID:      providerID(p, "max"),
		FirstName:  nullable(p.FirstName),
		LastName:   nullable(p.LastName),
		AvatarURL:  nullable(p.AvatarURL),
		Role:       domain.RoleClient,
	}
	uid, err := s.Store.InsertUser(u)
	if err != nil {
		return nil, err
	}
	full := strings.TrimSpace(p.FirstName + " " + p.LastName)
	_, _ = s.Store.CreateStudentFull(domain.Student{
		UserID:   &uid,
		FullName: full,
		Status:   "active",
		Source:   nullable(p.Provider),
	})
	created, err := s.Store.UserByID(uid)
	if err != nil {
		return nil, err
	}
	_ = s.notifyCoaches("new_client", map[string]any{"user_id": uid, "name": full})
	return s.issueTokens(created)
}

// LinkProviderToUser attaches a provider identity to an existing account, refusing
// to steal an identity already linked elsewhere.
func (s *Services) LinkProviderToUser(userID int64, p ProviderProfile) error {
	existing, err := s.userByProvider(p)
	if err != nil {
		return err
	}
	if existing != nil && existing.ID != userID {
		return ErrProviderTaken
	}
	return s.linkProvider(userID, p)
}

// SetAccountPassword sets/updates the password (and optionally phone) for an existing
// account — used when a Telegram/MAX-first user wants to enable phone+password login.
func (s *Services) SetAccountPassword(userID int64, phoneRaw, password string) error {
	if len(password) < 6 {
		return ErrWeakPassword
	}
	u, err := s.Store.UserByID(userID)
	if err != nil || u == nil {
		return ErrInvalidCreds
	}
	if phoneRaw != "" {
		phone := normalizePhone(phoneRaw)
		if !isValidPhone(phone) {
			return ErrInvalidPhone
		}
		if other, _ := s.Store.UserByPhone(phone); other != nil && other.ID != userID {
			return ErrPhoneTaken
		}
		u.Phone = &phone
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	u.PasswordHash = &hash
	if err := s.Store.UpdateUserProfile(*u); err != nil {
		return err
	}
	// Changing the password must end all existing sessions (token theft mitigation).
	return s.Store.RevokeAllRefreshTokens(userID)
}

func providerID(p ProviderProfile, provider string) *string {
	if p.Provider != provider {
		return nil
	}
	id := p.ExternalID
	return &id
}
