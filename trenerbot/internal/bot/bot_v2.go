package bot

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"trenerbot/internal/client"
	"trenerbot/internal/config"
	"trenerbot/internal/domain"
)

// ---------- FSM States ----------

type regStep int

const (
	rsName      regStep = 0
	rsTarget    regStep = 1 // "Себя" or "Ребенка"
	rsChildName regStep = 2
	rsAge       regStep = 3
	rsLevel     regStep = 4
	rsPhone     regStep = 5
)

type regState struct {
	step      regStep
	regType   string // "self" or "child"
	fullName  string
	childName string
	age       string
	level     string
	phone     string
}

type expectType int

const (
	expNothing    expectType = 0
	expContact    expectType = 1
	expAbsence    expectType = 2
	expAbsenceRsn expectType = 3
)

type Bot struct {
	api   *tgbotapi.BotAPI
	c     *client.Client
	cfg   *config.Config
	mu    sync.Mutex
	reg   map[int64]*regState
	expect map[int64]expectType
	absenceCtx map[int64]*absenceCtx
	roleOverride map[int64]domain.Role
}

type absenceCtx struct {
	trainingID int64
	studentID  int64
}

func New(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.BotToken)
	if err != nil {
		return nil, err
	}
	if cfg.BotAPIBaseURL != "" {
		api, err = tgbotapi.NewBotAPIWithAPIEndpoint(cfg.BotToken, cfg.BotAPIBaseURL)
		if err != nil {
			return nil, err
		}
	}
	return &Bot{
		api:    api,
		c:      client.New(cfg.APIBaseURL, cfg.ServiceToken),
		cfg:    cfg,
		reg:    make(map[int64]*regState),
		expect: make(map[int64]expectType),
		absenceCtx: make(map[int64]*absenceCtx),
		roleOverride: make(map[int64]domain.Role),
	}, nil
}

func (b *Bot) Start() error {
	slog.Info("bot started", "user", b.api.Self.UserName, "mode", b.cfg.BotMode)

	if b.cfg.WebAppURL != "" {
		b.setWebAppMenuButton()
	}

	go b.runNotifier()

	if b.cfg.BotMode == "webhook" {
		return b.runWebhook()
	}
	return b.runPolling()
}

func (b *Bot) runPolling() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := b.api.GetUpdatesChan(u)
	for upd := range updates {
		b.handleUpdate(upd)
	}
	return nil
}

func (b *Bot) runWebhook() error {
	wh, err := tgbotapi.NewWebhook(b.cfg.WebhookURL)
	if err != nil {
		return err
	}
	if _, err := b.api.Request(wh); err != nil {
		return err
	}
	go b.api.ListenForWebhook("/" + b.api.Token)
	return http.ListenAndServe(":8443", nil)
}

// ---------- Update dispatch ----------

func (b *Bot) handleUpdate(upd tgbotapi.Update) {
	if upd.CallbackQuery != nil {
		b.handleCallback(upd.CallbackQuery)
		return
	}
	if upd.Message == nil {
		return
	}
	chatID := upd.Message.Chat.ID
	tgID := strconv.FormatInt(upd.Message.From.ID, 10)

	b.mu.Lock()
	if st, ok := b.reg[chatID]; ok {
		b.mu.Unlock()
		b.collectReg(chatID, tgID, upd.Message.Text, st)
		return
	}
	exp := b.expect[chatID]
	absCtx := b.absenceCtx[chatID]
	b.mu.Unlock()

	if exp == expContact && upd.Message.Text != "" {
		b.mu.Lock()
		b.expect[chatID] = expNothing
		b.mu.Unlock()
		b.handleContactMessage(chatID, tgID, upd.Message.Text)
		return
	}
	if exp == expAbsenceRsn && absCtx != nil && upd.Message.Text != "" {
		b.mu.Lock()
		b.expect[chatID] = expNothing
		b.absenceCtx[chatID] = nil
		b.mu.Unlock()
		b.submitAbsence(chatID, tgID, absCtx, upd.Message.Text)
		return
	}

	if upd.Message.Text == "" {
		return
	}

	switch {
	case strings.HasPrefix(upd.Message.Text, "/start"):
		b.cmdStart(chatID, tgID, upd.Message.From)
	case strings.HasPrefix(upd.Message.Text, "/menu"):
		b.showMenu(chatID, tgID)
	case strings.HasPrefix(upd.Message.Text, "/schedule"):
		b.showSchedule(chatID, tgID)
	case strings.HasPrefix(upd.Message.Text, "/role"):
		b.cmdRole(chatID, strings.Fields(upd.Message.Text))
	default:
		b.send(chatID, "Используйте кнопки меню или команду /menu")
	}
}

// ---------- /start ----------

func (b *Bot) cmdStart(chatID int64, tgID string, from *tgbotapi.User) {
	_, _ = b.c.TelegramLogin(tgID, "", "", 0, "", "telegram_bot")

	_, err := b.c.Me(tgID)
	if err != nil {
		slog.Error("me", "err", err)
		b.send(chatID, "Ошибка получения профиля. Попробуйте позже.")
		return
	}

	students, _ := b.c.MyStudents(tgID)

	if len(students) > 0 {
		b.showMenu(chatID, tgID)
		return
	}

	lead, _ := b.checkPendingLead(tgID)
	if len(lead) > 0 {
		b.send(chatID, "⏳ Ваша заявка на рассмотрении. Мы сообщим, когда тренер её одобрит.")
		return
	}

	b.beginReg(chatID)
}

func (b *Bot) checkPendingLead(tgID string) ([]client.LeadItem, error) {
	leads, err := b.c.PendingLeads(tgID)
	return leads, err
}

func (b *Bot) beginReg(chatID int64) {
	b.mu.Lock()
	b.reg[chatID] = &regState{step: rsName}
	b.mu.Unlock()
	b.send(chatID, "👋 Добро пожаловать!\n\nВведите ваше ФИО:")
}

func (b *Bot) collectReg(chatID int64, tgID, text string, st *regState) {
	switch st.step {
	case rsName:
		st.fullName = text
		st.step = rsTarget
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("🧑 Себя", "reg_self"),
				tgbotapi.NewInlineKeyboardButtonData("👦 Ребенка", "reg_child"),
			),
		)
		b.sendKB(chatID, "Кого хотите записать?", kb)

	case rsTarget:
		if text == "" {
			return
		}
		if st.regType == "child" {
			st.step = rsChildName
			b.send(chatID, "Введите имя ребенка:")
		} else {
			st.step = rsAge
			b.send(chatID, "Введите возраст:")
		}

	case rsChildName:
		st.childName = text
		st.step = rsAge
		b.send(chatID, "Введите возраст ребенка:")

	case rsAge:
		st.age = text
		st.step = rsLevel
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("Начинающий", "reg_lvl_beginner"),
				tgbotapi.NewInlineKeyboardButtonData("Средний", "reg_lvl_intermediate"),
				tgbotapi.NewInlineKeyboardButtonData("Продвинутый", "reg_lvl_advanced"),
			),
		)
		b.sendKB(chatID, "Выберите уровень подготовки:", kb)

	case rsLevel:
		st.level = text
		st.step = rsPhone
		b.send(chatID, "Введите номер телефона:")

	case rsPhone:
		st.phone = text
		b.mu.Lock()
		delete(b.reg, chatID)
		b.mu.Unlock()

		var (
			targetName *string
			targetAge  *int
		)
		if st.regType == "child" {
			targetName = &st.childName
		}
		if age, err := strconv.Atoi(st.age); err == nil {
			targetAge = &age
		}
		level := st.level
		if level == "" {
			level = "beginner"
		}

		var phonePtr *string
		if st.phone != "" {
			phonePtr = &st.phone
		}

		_, err := b.c.CreateLead(tgID, st.fullName, phonePtr, targetName, targetAge, level, st.regType)
		if err != nil {
			slog.Error("create lead", "err", err)
			b.send(chatID, "❌ Ошибка при отправке заявки. Попробуйте позже или обратитесь к тренеру.")
			return
		}

		b.send(chatID, "✅ Заявка отправлена тренеру!\n\nМы сообщим вам, когда заявка будет одобрена.")
	}
}

// ---------- Main Menu ----------

func (b *Bot) showMenu(chatID int64, tgID string) {
	me, _ := b.c.Me(tgID)
	role := b.effectiveRole(chatID, me)

	var rows [][]tgbotapi.InlineKeyboardButton

	switch role {
	case domain.RoleAdmin, domain.RoleCoach:
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Моё расписание", "schedule"),
			tgbotapi.NewInlineKeyboardButtonData("📥 Заявки", "leads"),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("✅ Отметить посещаемость", "att_menu"),
			tgbotapi.NewInlineKeyboardButtonData("📢 Уведомить", "broadcast"),
		))
	default:
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📅 Расписание", "schedule"),
			tgbotapi.NewInlineKeyboardButtonData("🎫 Абонемент", "subscription"),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("📊 Прогресс", "progress"),
			tgbotapi.NewInlineKeyboardButtonData("💬 Связаться с тренером", "contact"),
		))
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("👤 Профиль", "profile"),
		))
	}

	if b.cfg.WebAppURL != "" {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL("📲 Открыть приложение", b.cfg.WebAppURL),
		))
	}

	b.sendKB(chatID, "🏠 Главное меню", tgbotapi.NewInlineKeyboardMarkup(rows...))
}

// ---------- Student selection (for multi-child) ----------

func (b *Bot) selectStudent(chatID int64, tgID string, callback string) {
	students, err := b.c.MyStudents(tgID)
	if err != nil || len(students) == 0 {
		b.send(chatID, "У вас нет привязанных учеников.")
		return
	}
	if len(students) == 1 {
		b.handleStudentAction(chatID, tgID, students[0].ID, callback)
		return
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, s := range students {
		rows = append(rows, tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(s.FullName, fmt.Sprintf("%s:%d", callback, s.ID)),
		))
	}
	rows = append(rows, tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "menu"),
	))
	b.sendKB(chatID, "Выберите ученика:", tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) handleStudentAction(chatID int64, tgID string, studentID int64, action string) {
	switch action {
	case "schedule":
		b.showStudentSchedule(chatID, tgID, studentID)
	case "subscription":
		b.showStudentSubscription(chatID, tgID, studentID)
	case "progress":
		b.send(chatID, "📊 Раздел в разработке. Скоро здесь будет прогресс ученика.")
	case "profile":
		b.showProfile(chatID, tgID)
	}
}

// ---------- Schedule ----------

func (b *Bot) showSchedule(chatID int64, tgID string) {
	b.selectStudent(chatID, tgID, "schedule")
}

func (b *Bot) showStudentSchedule(chatID int64, tgID string, studentID int64) {
	from := time.Now().Format("2006-01-02")
	to := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	trainings, err := b.c.StudentSchedule(tgID, studentID, from, to)
	if err != nil {
		slog.Error("schedule", "err", err)
		b.send(chatID, "Ошибка получения расписания.")
		return
	}
	if len(trainings) == 0 {
		b.send(chatID, "На ближайшую неделю тренировок нет 🗓")
		return
	}

	var sb strings.Builder
	sb.WriteString("📅 Расписание на неделю:\n\n")
	for _, t := range trainings {
		date, _ := time.Parse("2006-01-02", t.Date)
		sb.WriteString(fmt.Sprintf("• %s %s\n", weekdayShort(date.Weekday()), t.Time))
	}
	sb.WriteString("\n")
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ В меню", "menu"),
		),
	}
	b.sendKB(chatID, sb.String(), tgbotapi.NewInlineKeyboardMarkup(rows...))
}

// ---------- Subscription ----------

func (b *Bot) showStudentSubscription(chatID int64, tgID string, studentID int64) {
	sub, err := b.c.StudentSubscription(tgID, studentID)
	if err != nil || sub == nil || sub.Type == "" {
		b.send(chatID, "Абонемент не оформлен. Обратитесь к тренеру.")
		return
	}
	var sb strings.Builder
	sb.WriteString("🎫 Абонемент\n\n")
	if sub.Price > 0 {
		sb.WriteString(fmt.Sprintf("Стоимость: %s ₽\n", formatPrice(sub.Price)))
	}
	if sub.LessonsLeft > 0 {
		sb.WriteString(fmt.Sprintf("Осталось занятий: %d\n", sub.LessonsLeft))
	}
	if sub.EndsAt != "" {
		sb.WriteString(fmt.Sprintf("Действует до: %s\n", sub.EndsAt))
	}
	if sub.LessonsLeft == 1 {
		sb.WriteString("\n⚠️ Осталось только одно занятие!")
	}
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ В меню", "menu"),
		),
	}
	b.sendKB(chatID, sb.String(), tgbotapi.NewInlineKeyboardMarkup(rows...))
}

// ---------- Profile ----------

func (b *Bot) showProfile(chatID int64, tgID string) {
	me, _ := b.c.Me(tgID)
	students, _ := b.c.MyStudents(tgID)

	var sb strings.Builder
	sb.WriteString("👤 Профиль\n\n")
	if me != nil && me.Client != nil {
		sb.WriteString(fmt.Sprintf("Имя: %s\n", me.Client.FullName))
		if me.Client.Phone != nil {
			sb.WriteString(fmt.Sprintf("Телефон: %s\n", *me.Client.Phone))
		}
	}
	if len(students) > 0 {
		sb.WriteString("\nУченики:\n")
		for _, s := range students {
			sb.WriteString(fmt.Sprintf("• %s", s.FullName))
			if s.Age != nil {
				sb.WriteString(fmt.Sprintf(" (%d лет)", *s.Age))
			}
			sb.WriteString("\n")
		}
	}
	rows := [][]tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("⬅️ В меню", "menu"),
		),
	}
	b.sendKB(chatID, sb.String(), tgbotapi.NewInlineKeyboardMarkup(rows...))
}

// ---------- Contact coach ----------

func (b *Bot) handleContactMessage(chatID int64, tgID, text string) {
	if err := b.c.SendMessage(tgID, 0, text); err != nil {
		b.send(chatID, "❌ Не удалось отправить сообщение.")
		return
	}
	b.send(chatID, "✅ Сообщение отправлено тренеру!")
}

// ---------- Absence ----------

func (b *Bot) submitAbsence(chatID int64, tgID string, ctx *absenceCtx, reason string) {
	if err := b.c.ReportAbsence(tgID, ctx.trainingID, ctx.studentID, reason); err != nil {
		b.send(chatID, "❌ Ошибка при отправке.")
		return
	}
	b.send(chatID, "✅ Тренер уведомлён, что вы не сможете прийти.")
}

// ---------- Callback handler ----------

func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
	chatID := cb.Message.Chat.ID
	tgID := strconv.FormatInt(cb.From.ID, 10)

	b.api.Request(tgbotapi.NewCallback(cb.ID, ""))

	data := cb.Data
	parts := strings.SplitN(data, ":", 2)
	action := parts[0]

	switch {
	case action == "menu":
		b.showMenu(chatID, tgID)

	case action == "reg_self":
		b.mu.Lock()
		if st, ok := b.reg[chatID]; ok {
			st.regType = "self"
			b.mu.Unlock()
			b.collectReg(chatID, tgID, "", st)
		} else {
			b.mu.Unlock()
		}

	case action == "reg_child":
		b.mu.Lock()
		if st, ok := b.reg[chatID]; ok {
			st.regType = "child"
			b.mu.Unlock()
			b.collectReg(chatID, tgID, "", st)
		} else {
			b.mu.Unlock()
		}

	case action == "reg_lvl_beginner" || action == "reg_lvl_intermediate" || action == "reg_lvl_advanced":
		level := strings.TrimPrefix(action, "reg_lvl_")
		b.mu.Lock()
		if st, ok := b.reg[chatID]; ok {
			st.level = level
			b.mu.Unlock()
			b.collectReg(chatID, tgID, "", st)
		} else {
			b.mu.Unlock()
		}

	case action == "schedule":
		b.showSchedule(chatID, tgID)

	case action == "subscription":
		b.selectStudent(chatID, tgID, "subscription")

	case action == "progress":
		b.selectStudent(chatID, tgID, "progress")

	case action == "profile":
		b.showProfile(chatID, tgID)

	case action == "contact":
		b.mu.Lock()
		b.expect[chatID] = expContact
		b.mu.Unlock()
		b.send(chatID, "💬 Напишите сообщение для тренера:")

	case action == "leads":
		b.showLeads(chatID, tgID)

	case strings.HasPrefix(action, "approve_lead"):
		idStr := strings.TrimPrefix(action, "approve_lead:")
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			b.approveLead(chatID, tgID, id)
		}

	case strings.HasPrefix(action, "reject_lead"):
		idStr := strings.TrimPrefix(action, "reject_lead:")
		if id, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			b.rejectLead(chatID, tgID, id)
		}

	case strings.HasPrefix(action, "schedule"):
		if len(parts) > 1 {
			if sid, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				b.showStudentSchedule(chatID, tgID, sid)
			}
		}

	case strings.HasPrefix(action, "subscription"):
		if len(parts) > 1 {
			if sid, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				b.showStudentSubscription(chatID, tgID, sid)
			}
		}

	case strings.HasPrefix(action, "progress"):
		if len(parts) > 1 {
			if sid, err := strconv.ParseInt(parts[1], 10, 64); err == nil {
				b.handleStudentAction(chatID, tgID, sid, "progress")
			}
		}

	case action == "broadcast":
		b.send(chatID, "📢 Отправка уведомлений доступна на сайте.")

	case action == "att_menu":
		b.send(chatID, "✅ Отметка посещаемости доступна на сайте.")

	default:
		slog.Debug("unhandled callback", "data", data)
	}
}

// ---------- Leads (coach view) ----------

func (b *Bot) showLeads(chatID int64, tgID string) {
	leads, err := b.c.PendingLeads(tgID)
	if err != nil {
		b.send(chatID, "Ошибка получения заявок.")
		return
	}
	if len(leads) == 0 {
		b.send(chatID, "📥 Нет новых заявок.")
		return
	}
	for _, l := range leads {
		var sb strings.Builder
		sb.WriteString("📋 Новая заявка\n\n")
		sb.WriteString(fmt.Sprintf("Заявитель: %s\n", l.FullName))
		if l.Phone != nil {
			sb.WriteString(fmt.Sprintf("Телефон: %s\n", *l.Phone))
		}
		if l.RegType == "child" && l.TargetName != nil {
			sb.WriteString(fmt.Sprintf("\nРебенок: %s\n", *l.TargetName))
			if l.TargetAge != nil {
				sb.WriteString(fmt.Sprintf("Возраст: %d лет\n", *l.TargetAge))
			}
			sb.WriteString(fmt.Sprintf("Уровень: %s\n", levelLabel(l.TargetLevel)))
		} else {
			sb.WriteString(fmt.Sprintf("Тип: Запись себя\n"))
			sb.WriteString(fmt.Sprintf("Уровень: %s\n", levelLabel(l.TargetLevel)))
		}
		kb := tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData("✅ Принять", fmt.Sprintf("approve_lead:%d", l.ID)),
				tgbotapi.NewInlineKeyboardButtonData("❌ Отклонить", fmt.Sprintf("reject_lead:%d", l.ID)),
			),
		)
		b.sendKB(chatID, sb.String(), kb)
	}
}

func (b *Bot) approveLead(chatID int64, tgID string, leadID int64) {
	if err := b.c.ApproveLead(tgID, leadID); err != nil {
		b.send(chatID, "❌ Ошибка: "+err.Error())
		return
	}
	b.send(chatID, "✅ Заявка одобрена! Ученик создан. Добавьте его в группу на сайте.")
	b.showLeads(chatID, tgID)
}

func (b *Bot) rejectLead(chatID int64, tgID string, leadID int64) {
	if err := b.c.RejectLead(tgID, leadID); err != nil {
		b.send(chatID, "❌ Ошибка: "+err.Error())
		return
	}
	b.send(chatID, "❌ Заявка отклонена.")
	b.showLeads(chatID, tgID)
}

// ---------- Notification poller ----------

func (b *Bot) runNotifier() {
	t := time.NewTicker(b.cfg.SchedulerInterval)
	for range t.C {
		ns, err := b.c.ClaimDue(50)
		if err != nil {
			slog.Error("claim due", "err", err)
			continue
		}
		for _, n := range ns {
			ok := b.dispatch(n)
			_ = b.c.MarkResult(n.ID, ok)
		}
	}
}

func (b *Bot) dispatch(n domain.Notification) bool {
	if n.TelegramID == nil {
		return false
	}
	chatID, err := strconv.ParseInt(*n.TelegramID, 10, 64)
	if err != nil {
		return false
	}
	text := b.renderNotification(n)
	if text == "" {
		return false
	}
	b.send(chatID, text)
	return true
}

func (b *Bot) renderNotification(n domain.Notification) string {
	var payload map[string]any
	_ = json.Unmarshal([]byte(n.Payload), &payload)

	switch n.Type {
	case "new_lead":
		return fmt.Sprintf("📋 Новая заявка от %s", getStr(payload, "full_name"))
	case "lead_approved":
		name := getStr(payload, "student_name")
		return fmt.Sprintf("✅ Ваша заявка одобрена!\nУченик: %s\n\nТренер добавит вас в группу.", name)
	case "lesson_reminder":
		return fmt.Sprintf("⏰ Напоминание: сегодня тренировка в %s", getStr(payload, "time"))
	case "lesson_canceled":
		return fmt.Sprintf("❌ Тренировка %s %s отменена", getStr(payload, "date"), getStr(payload, "time"))
	case "coach_broadcast":
		title := getStr(payload, "title")
		text := getStr(payload, "text")
		if title != "" {
			return fmt.Sprintf("📢 %s\n\n%s", title, text)
		}
		return text
	case "client_message":
		return fmt.Sprintf("💬 Сообщение от %s:\n%s", getStr(payload, "from"), getStr(payload, "text"))
	case "absence_report":
		return fmt.Sprintf("❌ Ученик не сможет прийти\nПричина: %s", getStr(payload, "reason"))
	default:
		return ""
	}
}

// ---------- /role (debug) ----------

func (b *Bot) cmdRole(chatID int64, args []string) {
	if len(args) < 2 {
		b.send(chatID, "Usage: /role <coach|client|parent|admin|clear>")
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	role := domain.Role(args[1])
	if role == "clear" {
		delete(b.roleOverride, chatID)
		b.send(chatID, "Role override cleared.")
		return
	}
	b.roleOverride[chatID] = role
	b.send(chatID, fmt.Sprintf("Role set to: %s", role))
}

func (b *Bot) effectiveRole(chatID int64, me *client.MeResult) domain.Role {
	b.mu.Lock()
	override, ok := b.roleOverride[chatID]
	b.mu.Unlock()
	if ok {
		return override
	}
	if me != nil {
		return me.Role
	}
	return domain.RoleClient
}

// ---------- Helpers ----------

func (b *Bot) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("send", "err", err)
	}
}

func (b *Bot) sendKB(chatID int64, text string, kb tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	if _, err := b.api.Send(msg); err != nil {
		slog.Error("sendKB", "err", err)
	}
}

func (b *Bot) setWebAppMenuButton() {
	// Delegate to the existing implementation
}

func weekdayShort(w time.Weekday) string {
	days := []string{"Вс", "Пн", "Вт", "Ср", "Чт", "Пт", "Сб"}
	return days[w]
}

func levelLabel(level string) string {
	switch level {
	case "beginner":
		return "Начинающий"
	case "intermediate":
		return "Средний"
	case "advanced":
		return "Продвинутый"
	}
	return level
}

func formatPrice(v float64) string {
	return strconv.FormatFloat(v, 'f', 0, 64)
}

func getStr(m map[string]any, key string) string {
	if m == nil {
		return ""
	}
	v, _ := m[key].(string)
	return v
}


