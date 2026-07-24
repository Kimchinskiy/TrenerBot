package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"trenerbot/internal/domain"
)

// Client is the bot's HTTP adapter to the CRM backend (ТЗ §2/§15: bot talks only via API).
type Client struct {
	base string
	tok  string
	http *http.Client
}

func New(base, serviceToken string) *Client {
	if base == "" {
		base = "http://localhost:8080"
	}
	return &Client{base: base, tok: serviceToken, http: &http.Client{Timeout: 15 * time.Second}}
}

func (c *Client) do(method, path, telegramID string, body any, out any) error {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, c.base+path, rdr)
	if err != nil {
		return err
	}
	req.Header.Set("X-Service-Token", c.tok)
	if telegramID != "" {
		req.Header.Set("X-Telegram-Id", telegramID)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return fmt.Errorf("api %d: %s", resp.StatusCode, string(data))
	}
	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

type TelegramLoginResult struct {
	User   *domain.User  `json:"user"`
	Client *domain.Client `json:"client"`
	Token  string        `json:"token"`
}

func (c *Client) TelegramLogin(tgID, fullName, phone string, age int, medical, source string) (*TelegramLoginResult, error) {
	req := map[string]any{
		"telegram_id":    tgID,
		"full_name":      fullName,
		"phone":          phone,
		"age":            age,
		"medical_limits": medical,
		"source":         source,
	}
	var res TelegramLoginResult
	if err := c.do(http.MethodPost, "/api/auth/telegram", tgID, req, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

type MeResult struct {
	Role     domain.Role    `json:"role"`
	Client   *domain.Client `json:"client"`
	Children []domain.Client `json:"children"`
}

func (c *Client) Me(tgID string) (*MeResult, error) {
	var res MeResult
	if err := c.do(http.MethodGet, "/api/clients/me", tgID, nil, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (c *Client) ListLessons(tgID, from, to string) ([]domain.Lesson, error) {
	var out []domain.Lesson
	path := "/api/lessons?from=" + from + "&to=" + to
	if err := c.do(http.MethodGet, path, tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListCoaches(tgID string) ([]domain.Coach, error) {
	var out []domain.Coach
	if err := c.do(http.MethodGet, "/api/coaches", tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ListAttendance(tgID string, lessonID int64) ([]domain.Attendance, error) {
	var out []domain.Attendance
	if err := c.do(http.MethodGet, "/api/lessons/"+strconv.FormatInt(lessonID, 10)+"/attendance", tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) MarkAttendance(tgID string, lessonID, clientID int64, present bool) error {
	body := map[string]any{"client_id": clientID, "present": present}
	return c.do(http.MethodPost, "/api/lessons/"+strconv.FormatInt(lessonID, 10)+"/attendance", tgID, body, nil)
}

func (c *Client) RegisterClient(tgID string, lessonID, clientID int64) error {
	body := map[string]any{"client_id": clientID}
	return c.do(http.MethodPost, "/api/lessons/"+strconv.FormatInt(lessonID, 10)+"/register", tgID, body, nil)
}

// ClaimDue polls the notification outbox (system call, no telegram_id).
func (c *Client) ClaimDue(limit int) ([]domain.Notification, error) {
	var out []domain.Notification
	if err := c.do(http.MethodGet, "/api/notifications/due?limit="+strconv.Itoa(limit), "", nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) MarkResult(id int64, ok bool) error {
	body := map[string]any{"ok": ok}
	return c.do(http.MethodPost, "/api/notifications/"+strconv.FormatInt(id, 10)+"/result", "", body, nil)
}

func (c *Client) MessageCoaches(from, text string) error {
	body := map[string]any{"from": from, "text": text}
	return c.do(http.MethodPost, "/api/messages/coach", "", body, nil)
}

// Admin panel methods
type ClientListItem struct {
	ID                  int64   `json:"id"`
	FullName            string  `json:"full_name"`
	Phone               *string `json:"phone,omitempty"`
	BotAccess           bool    `json:"bot_access"`
	SubscriptionEndsAt *string `json:"subscription_ends_at,omitempty"`
}

func (c *Client) ListClients(tgID string) ([]ClientListItem, error) {
	var out []ClientListItem
	if err := c.do(http.MethodGet, "/api/admin/clients", tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GetClient(tgID string, clientID int64) (*domain.Client, error) {
	var out domain.Client
	if err := c.do(http.MethodGet, fmt.Sprintf("/api/admin/clients/%d", clientID), tgID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) GrantBotAccess(clientID int64, telegramID, endsAt string) error {
	body := map[string]any{"telegram_id": telegramID, "subscription_ends_at": endsAt}
	return c.do(http.MethodPost, fmt.Sprintf("/api/admin/clients/%d/grant-access", clientID), "", body, nil)
}

func (c *Client) RevokeBotAccess(clientID int64) error {
	return c.do(http.MethodPost, fmt.Sprintf("/api/admin/clients/%d/revoke-access", clientID), "", nil, nil)
}

func (c *Client) SearchClients(tgID, query string) ([]ClientListItem, error) {
	var out []ClientListItem
	path := "/api/admin/clients/search?q=" + url.QueryEscape(query)
	if err := c.do(http.MethodGet, path, tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// CoachSubscriptionResult holds the subscription status response.
type CoachSubscriptionResult struct {
	Subscription *domain.CoachSubscription `json:"subscription,omitempty"`
	Active       bool                      `json:"active"`
}

func (c *Client) CheckCoachSubscription(tgID string) (*CoachSubscriptionResult, error) {
	var out CoachSubscriptionResult
	if err := c.do(http.MethodGet, "/api/coach/subscription", tgID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// UploadFile sends a local file to the backend.
func (c *Client) UploadFile(tgID, ownerType string, ownerID int64, kind, path string) (int64, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", path)
	if err != nil {
		return 0, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return 0, err
	}
	_ = mw.WriteField("owner_type", ownerType)
	_ = mw.WriteField("owner_id", strconv.FormatInt(ownerID, 10))
	_ = mw.WriteField("kind", kind)
	_ = mw.Close()

	req, err := http.NewRequest(http.MethodPost, c.base+"/api/files", &buf)
	if err != nil {
		return 0, err
	}
	req.Header.Set("X-Service-Token", c.tok)
	req.Header.Set("X-Telegram-Id", tgID)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return 0, fmt.Errorf("upload %d: %s", resp.StatusCode, string(data))
	}
	var res struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return 0, err
	}
	return res.ID, nil
}

// ---------- Bot v2 API methods ----------

func (c *Client) CreateLead(tgID, fullName string, phone *string, targetName *string, targetAge *int, targetLevel, regType string) (int64, error) {
	body := map[string]any{
		"telegram_id":  tgID,
		"full_name":    fullName,
		"target_level": targetLevel,
		"reg_type":     regType,
	}
	if phone != nil {
		body["phone"] = *phone
	}
	if targetName != nil {
		body["target_name"] = *targetName
	}
	if targetAge != nil {
		body["target_age"] = *targetAge
	}
	var res struct {
		ID int64 `json:"id"`
	}
	if err := c.do(http.MethodPost, "/api/leads", tgID, body, &res); err != nil {
		return 0, err
	}
	return res.ID, nil
}

type LeadItem struct {
	ID          int64   `json:"id"`
	FullName    string  `json:"full_name"`
	Phone       *string `json:"phone,omitempty"`
	TargetName  *string `json:"target_name,omitempty"`
	TargetAge   *int    `json:"target_age,omitempty"`
	TargetLevel string  `json:"target_level"`
	RegType     string  `json:"reg_type"`
	Status      string  `json:"status"`
	CreatedAt   string  `json:"created_at"`
}

func (c *Client) PendingLeads(tgID string) ([]LeadItem, error) {
	var out []LeadItem
	if err := c.do(http.MethodGet, "/api/leads", tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ApproveLead(tgID string, leadID int64) error {
	body := map[string]any{"action": "approve"}
	return c.do(http.MethodPost, fmt.Sprintf("/api/leads/%d", leadID), tgID, body, nil)
}

func (c *Client) RejectLead(tgID string, leadID int64) error {
	body := map[string]any{"action": "reject"}
	return c.do(http.MethodPost, fmt.Sprintf("/api/leads/%d", leadID), tgID, body, nil)
}

type StudentResult struct {
	ID        int64   `json:"id"`
	FullName  string  `json:"full_name"`
	Age       *int    `json:"age,omitempty"`
	Level     string  `json:"level"`
	Phone     *string `json:"phone,omitempty"`
	Status    string  `json:"status"`
}

func (c *Client) MyStudents(tgID string) ([]StudentResult, error) {
	var out []StudentResult
	if err := c.do(http.MethodGet, "/api/me/students", tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

type TrainingResult struct {
	ID       int64   `json:"id"`
	Date     string  `json:"date"`
	Time     string  `json:"time"`
	Duration int     `json:"duration"`
	Status   string  `json:"status"`
	Location *string `json:"location,omitempty"`
	GroupID  *int64  `json:"group_id,omitempty"`
}

func (c *Client) StudentSchedule(tgID string, studentID int64, from, to string) ([]TrainingResult, error) {
	var out []TrainingResult
	path := fmt.Sprintf("/api/students/%d/trainings?from=%s&to=%s", studentID, from, to)
	if err := c.do(http.MethodGet, path, tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) ReportAbsence(tgID string, trainingID, studentID int64, reason string) error {
	body := map[string]any{"training_id": trainingID, "student_id": studentID, "reason": reason}
	return c.do(http.MethodPost, "/api/trainings/absence", tgID, body, nil)
}

type SubscriptionResult struct {
	Type        string  `json:"type"`
	Price       float64 `json:"price"`
	LessonsLeft int     `json:"lessons_left"`
	EndsAt      string  `json:"ends_at"`
}

func (c *Client) StudentSubscription(tgID string, studentID int64) (*SubscriptionResult, error) {
	var out SubscriptionResult
	path := fmt.Sprintf("/api/students/%d/subscription", studentID)
	if err := c.do(http.MethodGet, path, tgID, nil, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) SendMessage(tgID string, studentID int64, text string) error {
	body := map[string]any{"student_id": studentID, "text": text}
	return c.do(http.MethodPost, "/api/messages/coach", tgID, body, nil)
}

func (c *Client) Groups(tgID string) ([]domain.Group, error) {
	var out []domain.Group
	if err := c.do(http.MethodGet, "/api/groups", tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func (c *Client) GroupStudents(tgID string, groupID int64) ([]StudentResult, error) {
	var out []StudentResult
	path := fmt.Sprintf("/api/groups/%d/students", groupID)
	if err := c.do(http.MethodGet, path, tgID, nil, &out); err != nil {
		return nil, err
	}
	return out, nil
}
