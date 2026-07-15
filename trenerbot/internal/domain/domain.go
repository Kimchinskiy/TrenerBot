package domain

import "time"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleCoach  Role = "coach"
	RoleClient Role = "client"
	RoleParent Role = "parent"
)

type User struct {
	ID         int64      `json:"id"`
	TelegramID *string    `json:"telegram_id,omitempty"`
	Role       Role       `json:"role"`
	JWTRefresh *string    `json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Client struct {
	ID                 int64   `json:"id"`
	UserID             *int64  `json:"user_id,omitempty"`
	FullName           string  `json:"full_name"`
	Photo              *string `json:"photo,omitempty"`
	BirthDate          *string `json:"birth_date,omitempty"`
	Age                *int    `json:"age,omitempty"`
	Phone              *string `json:"phone,omitempty"`
	Telegram           *string `json:"telegram,omitempty"`
	WhatsApp           *string `json:"whatsapp,omitempty"`
	ParentID           *int64  `json:"parent_id,omitempty"`
	SecondParentID     *int64  `json:"second_parent_id,omitempty"`
	Email              *string `json:"email,omitempty"`
	MedicalLimits      *string `json:"medical_limits,omitempty"`
	Note               *string `json:"note,omitempty"`
	Status             string  `json:"status"`
	RegisteredAt       *string `json:"registered_at,omitempty"`
	Source             *string `json:"source,omitempty"`
	BotAccess          bool    `json:"bot_access"`
	SubscriptionEndsAt *string `json:"subscription_ends_at,omitempty"`
}

type Coach struct {
	ID       int64   `json:"id"`
	UserID   *int64  `json:"user_id,omitempty"`
	FullName string  `json:"full_name"`
	Photo    *string `json:"photo,omitempty"`
	Contacts *string `json:"contacts,omitempty"`
	Position *string `json:"position,omitempty"`
	Sport    *string `json:"sport,omitempty"`
	Schedule *string `json:"schedule,omitempty"`
	GroupIDs *string `json:"group_ids,omitempty"`
}

type Group struct {
	ID         int64   `json:"id"`
	Name       *string `json:"name,omitempty"`
	CoachID    *int64  `json:"coach_id,omitempty"`
	MaxMembers *int    `json:"max_members,omitempty"`
	Schedule   *string `json:"schedule,omitempty"`
	Price      *float64 `json:"price,omitempty"`
	Location   *string `json:"location,omitempty"`
	Active     int     `json:"active"`
}

type LessonStatus string

const (
	LessonPlanned LessonStatus = "planned"
	LessonOngoing LessonStatus = "ongoing"
	LessonDone    LessonStatus = "done"
	LessonCanceled LessonStatus = "canceled"
	LessonMoved   LessonStatus = "moved"
)

type Lesson struct {
	ID       int64        `json:"id"`
	Date     string       `json:"date"` // YYYY-MM-DD
	Time     string       `json:"time"` // HH:MM
	CoachID  *int64       `json:"coach_id,omitempty"`
	Duration int          `json:"duration"`
	Status   LessonStatus `json:"status"`
	Location *string      `json:"location,omitempty"`
	Comment  *string      `json:"comment,omitempty"`
	GroupID  *int64       `json:"group_id,omitempty"`
}

type Attendance struct {
	LessonID int64      `json:"lesson_id"`
	ClientID int64      `json:"client_id"`
	Present  bool       `json:"present"`
	MarkedAt *string    `json:"marked_at,omitempty"`
	MarkedBy *int64     `json:"marked_by,omitempty"`
}

type Notification struct {
	ID             int64      `json:"id"`
	Channel        string     `json:"channel"`
	RecipientUserID int64     `json:"recipient_user_id"`
	TelegramID     *string    `json:"telegram_id,omitempty"`
	Type           string     `json:"type"`
	Payload        string     `json:"payload"` // JSON
	SendAt         time.Time  `json:"send_at"`
	Status         string     `json:"status"`
	SentAt         *time.Time `json:"sent_at,omitempty"`
}

type File struct {
	ID        int64   `json:"id"`
	OwnerType string  `json:"owner_type"`
	OwnerID   int64   `json:"owner_id"`
	Path      string  `json:"path"`
	Kind      string  `json:"kind"`
}
