package domain

import "time"

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleCoach  Role = "coach"
	RoleClient Role = "client"
	RoleParent Role = "parent"
)

// User is a single account. Telegram, MAX and password are just ways to
// authenticate the same user; the primary identifier is User.ID (not TelegramID).
type User struct {
	ID           int64   `json:"id"`
	Phone        *string `json:"phone,omitempty"`
	PasswordHash *string `json:"-"`
	TelegramID   *string `json:"telegram_id,omitempty"`
	MaxID        *string `json:"max_id,omitempty"`
	FirstName    *string `json:"first_name,omitempty"`
	LastName     *string `json:"last_name,omitempty"`
	AvatarURL    *string `json:"avatar_url,omitempty"`
	Role         Role    `json:"role"`
	JWTRefresh   *string `json:"-"`

	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// HasPassword reports whether the user can log in with phone+password.
func (u *User) HasPassword() bool { return u.PasswordHash != nil && *u.PasswordHash != "" }

// AuthMethods lists the enabled login methods for this account.
func (u *User) AuthMethods() []string {
	methods := []string{}
	if u.HasPassword() {
		methods = append(methods, "password")
	}
	if u.TelegramID != nil && *u.TelegramID != "" {
		methods = append(methods, "telegram")
	}
	if u.MaxID != nil && *u.MaxID != "" {
		methods = append(methods, "max")
	}
	return methods
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
	ID          int64    `json:"id"`
	Name        *string  `json:"name,omitempty"`
	CoachID     *int64   `json:"coach_id,omitempty"`
	MaxMembers  *int     `json:"max_members,omitempty"`
	Schedule    *string  `json:"schedule,omitempty"`
	Price       *float64 `json:"price,omitempty"`
	Location    *string  `json:"location,omitempty"`
	Active      int      `json:"active"`
	MemberCount int      `json:"member_count"`
}

type GroupMember struct {
	ID        int64  `json:"id"`
	GroupID   int64  `json:"group_id"`
	ClientID  int64  `json:"client_id"`
	Role      string `json:"role"`
	JoinedAt  string `json:"joined_at"`
	ClientName string `json:"client_name,omitempty"`
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

type DailyAttendance struct {
	Date     string `json:"date"`
	ClientID int64  `json:"client_id"`
	Present  bool   `json:"present"`
	MarkedBy *int64 `json:"marked_by,omitempty"`
}

type DateAttendanceClient struct {
	ClientID int64   `json:"client_id"`
	FullName string  `json:"full_name"`
	Time     string  `json:"time"`
	Present  *bool   `json:"present"`
	Photo    *string `json:"photo,omitempty"`
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

// LessonEntry is a single athlete's training session. Each row = one specific
// athlete on a specific date/time. This replaces the old lessons+attendance
// model for the schedule feature.
type LessonEntry struct {
	ID        int64         `json:"id"`
	Date      string        `json:"date"`
	Time      string        `json:"time"`
	ClientID  int64         `json:"client_id"`
	CoachID   *int64        `json:"coach_id,omitempty"`
	GroupID   *int64        `json:"group_id,omitempty"`
	Duration  int           `json:"duration"`
	Status    LessonStatus  `json:"status"`
	Comment   *string       `json:"comment,omitempty"`
	CreatedAt string        `json:"created_at,omitempty"`
	UpdatedAt string        `json:"updated_at,omitempty"`
}

// ScheduleEntry is the API response type for the schedule: a lesson entry
// joined with the athlete's name for display.
type ScheduleEntry struct {
	ID         int64        `json:"id"`
	Date       string       `json:"date"`
	Time       string       `json:"time"`
	ClientID   int64        `json:"client_id"`
	ClientName string       `json:"client_name"`
	CoachID    *int64       `json:"coach_id,omitempty"`
	Duration   int          `json:"duration"`
	Status     LessonStatus `json:"status"`
	GroupID    *int64       `json:"group_id,omitempty"`
	GroupName  *string      `json:"group_name,omitempty"`
}

// Recipient is a coach's client eligible for notifications.
type Recipient struct {
	ClientID int64  `json:"client_id"`
	FullName string `json:"full_name"`
	UserID   *int64 `json:"user_id"`
}

type SubscriptionStatus string

const (
	SubTrial   SubscriptionStatus = "trial"
	SubActive  SubscriptionStatus = "active"
	SubExpired SubscriptionStatus = "expired"
	SubCanceled SubscriptionStatus = "canceled"
)

type CoachSubscription struct {
	ID        int64              `json:"id"`
	CoachID   int64              `json:"coach_id"`
	Status    SubscriptionStatus `json:"status"`
	TrialStart string            `json:"trial_start"`
	TrialEnd  *string            `json:"trial_end,omitempty"`
	PaidUntil *string            `json:"paid_until,omitempty"`
	CreatedAt string             `json:"created_at"`
	UpdatedAt *string            `json:"updated_at,omitempty"`
}

type ClientSubscription struct {
	ID          int64   `json:"id"`
	ClientID    int64   `json:"client_id"`
	Type        string  `json:"type"`
	Price       float64 `json:"price"`
	BoughtAt    string  `json:"bought_at"`
	EndsAt      string  `json:"ends_at"`
	LessonsLeft int     `json:"lessons_left"`
	Freeze      int     `json:"freeze"`
	CreatedAt   string  `json:"created_at"`
}

type SocialLink struct {
	ID       int64  `json:"id"`
	CoachID  int64  `json:"coach_id"`
	Platform string `json:"platform"`
	URL      *string `json:"url,omitempty"`
	Enabled  bool   `json:"enabled"`
}

type ParentNotifPref struct {
	ID           int64 `json:"id"`
	ParentUserID int64 `json:"parent_user_id"`
	ChildID      int64 `json:"child_id"`
	LessonStart  bool  `json:"lesson_start"`
	LessonEnd15  bool  `json:"lesson_end_15"`
	LessonMissed bool  `json:"lesson_missed"`
}

type ChildLessonStatus struct {
	ClientID       int64  `json:"client_id"`
	FullName       string `json:"full_name"`
	Date           string `json:"date"`
	Time           string `json:"time"`
	Duration       int    `json:"duration"`
	Status         string `json:"status"`
	IsToday        bool   `json:"is_today"`
	IsOngoing      bool   `json:"is_ongoing"`
	MinutesLeft    *int   `json:"minutes_left,omitempty"`
	MinutesUntil   *int   `json:"minutes_until,omitempty"`
	HasLessonToday bool   `json:"has_lesson_today"`
}

type StatisticsResponse struct {
	Period         string            `json:"period"`
	DateFrom       string            `json:"date_from"`
	DateTo         string            `json:"date_to"`
	Trainings      MetricValue       `json:"trainings"`
	Clients        MetricValue       `json:"clients"`
	Income         IncomeMetric      `json:"income"`
	Attendance     MetricValue       `json:"attendance"`
	Debtors        DebtorsSummary    `json:"debtors"`
	IncomeChart    []ChartPoint      `json:"income_chart"`
	QuickOverview  QuickOverview     `json:"quick_overview"`
}

type MetricValue struct {
	Value    float64 `json:"value"`
	Change   float64 `json:"change"`
	Label    string  `json:"label"`
}

type IncomeMetric struct {
	Value    float64 `json:"value"`
	Change   float64 `json:"change"`
	Label    string  `json:"label"`
}

type DebtorsSummary struct {
	Count     int            `json:"count"`
	TotalDebt float64        `json:"total_debt"`
	Items     []DebtorItem   `json:"items"`
}

type DebtorItem struct {
	ClientID  int64   `json:"client_id"`
	FullName  string  `json:"full_name"`
	Phone     *string `json:"phone,omitempty"`
	Debt      float64 `json:"debt"`
	EndsAt    *string `json:"ends_at,omitempty"`
}

type ChartPoint struct {
	Label string  `json:"label"`
	Value float64 `json:"value"`
}

type QuickOverview struct {
	NewClients      int     `json:"new_clients"`
	AverageCheck    float64 `json:"average_check"`
	CanceledCount   int     `json:"canceled_count"`
	AvgAttendance   float64 `json:"avg_attendance"`
	AvgGroupSize    float64 `json:"avg_group_size"`
	BusiestDay      string  `json:"busiest_day"`
	PopularTime     string  `json:"popular_time"`
}

// ---------- Bot v2 domain ----------

type Student struct {
	ID             int64   `json:"id"`
	FullName       string  `json:"full_name"`
	BirthDate      *string `json:"birth_date,omitempty"`
	Age            *int    `json:"age,omitempty"`
	Level          string  `json:"level"`
	Phone          *string `json:"phone,omitempty"`
	AdditionalInfo *string `json:"additional_info,omitempty"`
	Status         string  `json:"status"`
	ClientID       *int64  `json:"client_id,omitempty"`
	CreatedAt      string  `json:"created_at"`
	UpdatedAt      *string `json:"updated_at,omitempty"`
}

type RelationType string

const (
	RelSelf     RelationType = "self"
	RelParent   RelationType = "parent"
	RelGuardian RelationType = "guardian"
)

type Relationship struct {
	ID        int64   `json:"id"`
	UserID    int64   `json:"user_id"`
	StudentID int64   `json:"student_id"`
	Relation  RelationType `json:"relation"`
	CreatedAt string  `json:"created_at"`
}

type LeadStatus string

const (
	LeadPending  LeadStatus = "pending"
	LeadApproved LeadStatus = "approved"
	LeadRejected LeadStatus = "rejected"
)

type Lead struct {
	ID              int64     `json:"id"`
	TelegramID      string    `json:"telegram_id"`
	FullName        string    `json:"full_name"`
	Phone           *string   `json:"phone,omitempty"`
	TargetName      *string   `json:"target_name,omitempty"`
	TargetAge       *int      `json:"target_age,omitempty"`
	TargetLevel     string    `json:"target_level"`
	RegType         string    `json:"reg_type"`
	Status          LeadStatus `json:"status"`
	CreatedAt       string    `json:"created_at"`
	ReviewedAt      *string   `json:"reviewed_at,omitempty"`
	ReviewedBy      *int64    `json:"reviewed_by,omitempty"`
	CreatedUserID   *int64    `json:"created_user_id,omitempty"`
	CreatedStudentID *int64    `json:"created_student_id,omitempty"`
}

type TrainingTemplate struct {
	ID       int64  `json:"id"`
	GroupID  int64  `json:"group_id"`
	Weekday  int    `json:"weekday"`
	Time     string `json:"time"`
	Duration int    `json:"duration"`
}

type Training struct {
	ID       int64  `json:"id"`
	GroupID  *int64 `json:"group_id,omitempty"`
	CoachID  *int64 `json:"coach_id,omitempty"`
	Date    string `json:"date"`
	Time    string `json:"time"`
	Duration int    `json:"duration"`
	Status   string `json:"status"`
	Location *string `json:"location,omitempty"`
	Comment  *string `json:"comment,omitempty"`
}

type TrainingAbsence struct {
	ID         int64  `json:"id"`
	TrainingID int64  `json:"training_id"`
	StudentID  int64  `json:"student_id"`
	Reason     string `json:"reason"`
	CreatedAt  string `json:"created_at"`
}

type NotificationPref struct {
	UserID        int64  `json:"user_id"`
	StudentID     *int64 `json:"student_id,omitempty"`
	ReminderDay    int    `json:"reminder_day"`
	ReminderHours  int    `json:"reminder_hours"`
	LessonsLow     int    `json:"lessons_low"`
	SubExpiring    int    `json:"sub_expiring"`
	News           int    `json:"news"`
}

type ClientMessage struct {
	ID        int64  `json:"id"`
	UserID    *int64 `json:"user_id,omitempty"`
	StudentID *int64 `json:"student_id,omitempty"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}
