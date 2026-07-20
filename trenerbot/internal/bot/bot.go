package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"trenerbot/internal/client"
	"trenerbot/internal/config"
	"trenerbot/internal/domain"
)

type regState struct {
	step             int // 0 name, 1 phone, 2 age, 3 medical
	fullName, phone, age, medical string
}

type adminState struct {
	step        int // 0 list, 1 select client, 2 set telegram_id, 3 set subscription_ends
	clientID    int64
	telegramID  string
	subscriptionEnds string
}

type Bot struct {
	api            *tgbotapi.BotAPI
	c              *client.Client
	cfg            *config.Config
	mu             sync.Mutex
	reg            map[int64]*regState
	activeLesson   map[int64]int64
	expectContact  map[int64]bool
	roleOverride   map[int64]domain.Role
	admin          map[int64]*adminState
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
		api:           api,
		c:             client.New(cfg.APIBaseURL, cfg.ServiceToken),
		cfg:           cfg,
		reg:           map[int64]*regState{},
		activeLesson:  map[int64]int64{},
		expectContact: map[int64]bool{},
		roleOverride:  map[int64]domain.Role{},
		admin:         map[int64]*adminState{},
	}, nil
}

func (b *Bot) Start() error {
	slog.Info("bot started", "username", b.api.Self.UserName, "mode", b.cfg.BotMode)

	// Persistent menu button that launches the Mini App.
	if b.cfg.WebAppURL != "" {
		if err := b.setWebAppMenuButton(); err != nil {
			slog.Warn("set chat menu button", "err", err)
		}
	}

	go b.runNotifier()

	var updates tgbotapi.UpdatesChannel
	if b.cfg.BotMode == "webhook" {
		wh, err := tgbotapi.NewWebhook(b.cfg.WebhookURL)
		if err != nil {
			return err
		}
		if _, err := b.api.Request(wh); err != nil {
			return err
		}
		updates = b.api.ListenForWebhook("/" + b.api.Token)
		go func() {
			addr := os.Getenv("WEBHOOK_LISTEN")
			if addr == "" {
				addr = ":8443"
			}
			_ = http.ListenAndServe(addr, nil)
		}()
	} else {
		u := tgbotapi.NewUpdate(0)
		u.Timeout = 60
		updates = b.api.GetUpdatesChan(u)
	}

	for upd := range updates {
		b.handleUpdate(upd)
	}
	return nil
}

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
	text := strings.TrimSpace(upd.Message.Text)

	// registration flow has priority
	b.mu.Lock()
	st := b.reg[chatID]
	b.mu.Unlock()
	if st != nil {
		b.collectReg(chatID, tgID, upd.Message.From, text)
		return
	}

	// photo from coach with an open lesson
	if upd.Message.Photo != nil {
		b.handlePhoto(chatID, tgID, upd.Message)
		return
	}

	// expecting a free-text message to coach
	b.mu.Lock()
	expect := b.expectContact[chatID]
	b.mu.Unlock()
	if expect {
		b.mu.Lock()
		b.expectContact[chatID] = false
		b.mu.Unlock()
		name := upd.Message.From.FirstName
		if err := b.c.MessageCoaches(name, text); err != nil {
			b.send(chatID, "Не удалось отправить сообщение тренеру.")
		} else {
			b.send(chatID, "Сообщение отправлено тренеру ✅")
		}
		return
	}

	// Handle admin text input (e.g., days for subscription)
	if b.handleAdminText(chatID, tgID, text) {
		return
	}

	if text == "" {
		return
	}
	switch {
	case strings.HasPrefix(text, "/start"):
		b.cmdStart(chatID, tgID, upd.Message.From)
	case strings.HasPrefix(text, "/role"):
		b.cmdRole(chatID, tgID, text)
	case strings.HasPrefix(text, "/admin"):
		b.cmdAdmin(chatID, tgID, text)
	case strings.HasPrefix(text, "/menu"):
		b.showMenu(chatID, tgID)
	case strings.HasPrefix(text, "/schedule"):
		b.showSchedule(chatID, tgID)
	case strings.HasPrefix(text, "/help"):
		b.send(chatID, "Команды: /start, /menu, /schedule, /help, /role, /admin")
	default:
		b.send(chatID, "Используйте кнопки меню или команду /menu")
	}
}

// ---------- Commands ----------

func (b *Bot) effectiveRole(chatID int64, me *client.MeResult) domain.Role {
	b.mu.Lock()
	defer b.mu.Unlock()
	if r, ok := b.roleOverride[chatID]; ok {
		return r
	}
	return me.Role
}

func (b *Bot) cmdRole(chatID int64, tgID string, text string) {
	parts := strings.Fields(text)
	if len(parts) < 2 {
		b.send(chatID, "Использование: /role <coach|client|parent|admin|clear>")
		return
	}
	var r domain.Role
	switch parts[1] {
	case "coach":
		r = domain.RoleCoach
	case "client":
		r = domain.RoleClient
	case "parent":
		r = domain.RoleParent
	case "admin":
		r = domain.RoleAdmin
	case "clear":
		b.mu.Lock()
		delete(b.roleOverride, chatID)
		b.mu.Unlock()
		b.send(chatID, "Роль сброшена к реальной из БД ✅")
		return
	default:
		b.send(chatID, "Неизвестная роль. Доступно: coach, client, parent, admin, clear")
		return
	}
	b.mu.Lock()
	b.roleOverride[chatID] = r
	b.mu.Unlock()
	b.send(chatID, fmt.Sprintf("Роль переключена на: %s ✅", r))
	b.showMenu(chatID, tgID)
}

// Admin panel commands
func (b *Bot) cmdAdmin(chatID int64, tgID string, text string) {
	me, err := b.c.Me(tgID)
	if err != nil {
		b.send(chatID, "Ошибка получения профиля.")
		return
	}
	role := b.effectiveRole(chatID, me)
	if role != domain.RoleAdmin {
		b.send(chatID, "⛔ Нет прав администратора.")
		return
	}

	parts := strings.Fields(text)
	if len(parts) < 2 {
		b.showAdminMenu(chatID)
		return
	}

	switch parts[1] {
	case "clients":
		b.showAdminClients(chatID, tgID, 0)
	case "grant":
		if len(parts) < 3 {
			b.send(chatID, "Использование: /admin grant <client_id> [telegram_id] [days]")
			return
		}
		clientID, _ := strconv.ParseInt(parts[2], 10, 64)
		var tgID string
		if len(parts) > 3 {
			tgID = parts[3]
		}
		days := 30
		if len(parts) > 4 {
			days, _ = strconv.Atoi(parts[4])
		}
		b.grantBotAccess(chatID, clientID, tgID, days)
	case "revoke":
		if len(parts) < 3 {
			b.send(chatID, "Использование: /admin revoke <client_id>")
			return
		}
		clientID, _ := strconv.ParseInt(parts[2], 10, 64)
		b.revokeBotAccess(chatID, tgID, clientID)
	case "search":
		if len(parts) < 3 {
			b.send(chatID, "Использование: /admin search <name_or_phone>")
			return
		}
		query := strings.Join(parts[2:], " ")
		b.searchClients(chatID, tgID, query)
	default:
		b.showAdminMenu(chatID)
	}
}

func (b *Bot) showAdminMenu(chatID int64) {
	rows := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("👥 Список клиентов", "admin_clients:0")},
		{tgbotapi.NewInlineKeyboardButtonData("🔍 Поиск клиента", "admin_search")},
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.sendWithKeyboard(chatID, "🔧 Панель администратора:", kb)
}

func (b *Bot) showAdminClients(chatID int64, tgID string, page int) {
	clients, err := b.c.ListClients(tgID)
	if err != nil {
		b.send(chatID, "Ошибка загрузки клиентов.")
		return
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("👥 Клиенты (стр. %d):\n\n", page+1))

	pageSize := 10
	start := page * pageSize
	end := start + pageSize
	if end > len(clients) {
		end = len(clients)
	}

	var rows [][]tgbotapi.InlineKeyboardButton
	for i := start; i < end; i++ {
		c := clients[i]
		access := "❌"
		if c.BotAccess {
			access = "✅"
		}
		subEnds := ""
		if c.SubscriptionEndsAt != nil && *c.SubscriptionEndsAt != "" {
			subEnds = fmt.Sprintf(" до %s", *c.SubscriptionEndsAt)
		}
		sb.WriteString(fmt.Sprintf("%d. %s %s%s\n", i+1, c.FullName, access, subEnds))
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %s", c.FullName, access), fmt.Sprintf("admin_client:%d", c.ID)),
		})
	}

	// Pagination
	if page > 0 {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", fmt.Sprintf("admin_clients:%d", page-1)),
		})
	}
	if end < len(clients) {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData("Вперед ➡️", fmt.Sprintf("admin_clients:%d", page+1)),
		})
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⬅️ В меню", "admin_menu"),
	})

	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.sendWithKeyboard(chatID, sb.String(), kb)
}

func (b *Bot) showAdminClientDetail(chatID int64, tgID string, clientID int64) {
	client, err := b.c.GetClient(tgID, clientID)
	if err != nil {
		b.send(chatID, "Клиент не найден.")
		return
	}

	access := "❌ Нет доступа"
	if client.BotAccess {
		access = "✅ Есть доступ"
	}
	subEnds := "—"
	if client.SubscriptionEndsAt != nil && *client.SubscriptionEndsAt != "" {
		subEnds = *client.SubscriptionEndsAt
	}

	text := fmt.Sprintf(`👤 %s
📞 %s
🆔 ID: %d
🤖 Бот: %s
📅 Подписка до: %s`, client.FullName, client.Phone, client.ID, access, subEnds)

	rows := [][]tgbotapi.InlineKeyboardButton{
		{tgbotapi.NewInlineKeyboardButtonData("✅ Выдать доступ", fmt.Sprintf("admin_grant:%d", clientID))},
		{tgbotapi.NewInlineKeyboardButtonData("❌ Отозвать доступ", fmt.Sprintf("admin_revoke:%d", clientID))},
		{tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "admin_clients:0")},
	}
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.sendWithKeyboard(chatID, text, kb)
}

func (b *Bot) grantBotAccess(chatID int64, clientID int64, tgID string, days int) {
	endsAt := time.Now().AddDate(0, 0, days).Format("2006-01-02")
	err := b.c.GrantBotAccess(clientID, tgID, endsAt)
	if err != nil {
		b.send(chatID, fmt.Sprintf("Ошибка: %v", err))
		return
	}
	b.send(chatID, fmt.Sprintf("✅ Доступ выдан клиенту #%d на %d дней (до %s)", clientID, days, endsAt))
	b.showAdminClientDetail(chatID, tgID, clientID)
}

func (b *Bot) revokeBotAccess(chatID int64, tgID string, clientID int64) {
	err := b.c.RevokeBotAccess(clientID)
	if err != nil {
		b.send(chatID, fmt.Sprintf("Ошибка: %v", err))
		return
	}
	b.send(chatID, fmt.Sprintf("✅ Доступ отозван у клиента #%d", clientID))
	b.showAdminClientDetail(chatID, tgID, clientID)
}

func (b *Bot) promptGrantDays(chatID int64, tgID string, clientID int64) {
	b.mu.Lock()
	b.admin[chatID] = &adminState{step: 1, clientID: clientID}
	b.mu.Unlock()
	b.send(chatID, "Введите количество дней подписки (по умолчанию 30):")
}

func (b *Bot) handleAdminText(chatID int64, tgID string, text string) bool {
	b.mu.Lock()
	st := b.admin[chatID]
	b.mu.Unlock()
	if st == nil {
		return false
	}

	switch st.step {
	case 1: // waiting for days
		days := 30
		if text != "" {
			d, err := strconv.Atoi(text)
			if err == nil && d > 0 {
				days = d
			}
		}
		b.mu.Lock()
		st.step = 2
		b.mu.Unlock()
		b.grantBotAccess(chatID, st.clientID, tgID, days)
		return true
	}
	return false
}

func (b *Bot) searchClients(chatID int64, tgID string, query string) {
	clients, err := b.c.SearchClients(tgID, query)
	if err != nil {
		b.send(chatID, "Ошибка поиска.")
		return
	}
	if len(clients) == 0 {
		b.send(chatID, "Клиенты не найдены.")
		return
	}

	var sb strings.Builder
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, c := range clients {
		access := "❌"
		if c.BotAccess {
			access = "✅"
		}
		sb.WriteString(fmt.Sprintf("%s %s (ID: %d)\n", access, c.FullName, c.ID))
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(fmt.Sprintf("%s %s", c.FullName, access), fmt.Sprintf("admin_client:%d", c.ID)),
		})
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⬅️ В меню", "admin_menu"),
	})
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.sendWithKeyboard(chatID, sb.String(), kb)
}

func (b *Bot) cmdStart(chatID int64, tgID string, from *tgbotapi.User) {
	// ensure user + client exist (self-registration, ТЗ §4)
	if _, err := b.c.TelegramLogin(tgID, "", "", 0, "", "telegram_bot"); err != nil {
		b.send(chatID, "Ошибка регистрации, попробуйте позже.")
		return
	}
	me, err := b.c.Me(tgID)
	if err != nil {
		b.send(chatID, "Ошибка получения профиля.")
		return
	}
	role := b.effectiveRole(chatID, me)
	
	// Check bot access for clients
	if role == domain.RoleClient {
		if me.Client == nil || me.Client.FullName == "" {
			b.beginReg(chatID)
			return
		}
		if me.Client != nil && !me.Client.BotAccess {
			b.send(chatID, "⛔ У вас нет доступа к боту. Обратитесь к администратору для оформления подписки.")
			return
		}
	}
	b.showMenu(chatID, tgID)
}

func (b *Bot) beginReg(chatID int64) {
	b.mu.Lock()
	b.reg[chatID] = &regState{step: 0}
	b.mu.Unlock()
	b.send(chatID, "Давайте зарегистрируемся. Отправьте ваше ФИО:")
}

func (b *Bot) collectReg(chatID int64, tgID string, from *tgbotapi.User, text string) {
	b.mu.Lock()
	st := b.reg[chatID]
	b.mu.Unlock()

	switch st.step {
	case 0:
		st.fullName = text
		st.step = 1
		b.send(chatID, "Телефон:")
	case 1:
		st.phone = text
		st.step = 2
		b.send(chatID, "Возраст (числом):")
	case 2:
		age, _ := strconv.Atoi(text)
		st.age = text
		_ = age
		st.step = 3
		b.send(chatID, "Медицинские ограничения (или '-' / пусто, если нет):")
	case 3:
		medical := text
		if medical == "-" {
			medical = ""
		}
		age, _ := strconv.Atoi(st.age)
		if _, err := b.c.TelegramLogin(tgID, st.fullName, st.phone, age, medical, "telegram_bot"); err != nil {
			b.send(chatID, "Не удалось сохранить профиль, попробуйте /start.")
		} else {
			b.send(chatID, "Регистрация завершена! ✅")
		}
		b.mu.Lock()
		delete(b.reg, chatID)
		b.mu.Unlock()
		b.showMenu(chatID, tgID)
	}
}

// ---------- Menu ----------

func (b *Bot) showMenu(chatID int64, tgID string) {
	me, err := b.c.Me(tgID)
	if err != nil {
		b.send(chatID, "Ошибка меню.")
		return
	}
	role := b.effectiveRole(chatID, me)
	var rows [][]tgbotapi.InlineKeyboardButton
	switch role {
	case domain.RoleCoach:
		// Check subscription
		subResp, subErr := b.c.CheckCoachSubscription(tgID)
		if subErr != nil || subResp == nil || !subResp.Active {
			rows = [][]tgbotapi.InlineKeyboardButton{
				{tgbotapi.NewInlineKeyboardButtonData("📅 Моё расписание", "schedule")},
				{tgbotapi.NewInlineKeyboardButtonData("💳 Оформить подписку", "subscription")},
			}
			kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
			b.sendWithKeyboard(chatID, "⛔ Подписка неактивна. Доступные функции ограничены.\nОформите подписку для полного доступа к боту.", kb)
			return
		}
		rows = [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("📅 Моё расписание", "schedule")},
			{tgbotapi.NewInlineKeyboardButtonData("✅ Отметить посещаемость", "att_menu")},
			{tgbotapi.NewInlineKeyboardButtonData("📷 Загрузить фото", "photo_menu")},
		}
	case domain.RoleParent:
		rows = [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("📅 Расписание детей", "schedule")},
		}
	default: // client
		rows = [][]tgbotapi.InlineKeyboardButton{
			{tgbotapi.NewInlineKeyboardButtonData("📅 Моё расписание", "schedule")},
			{tgbotapi.NewInlineKeyboardButtonData("💬 Связь с тренером", "contact")},
		}
	}
	if b.cfg.WebAppURL != "" {
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonURL("📲 Открыть приложение", b.cfg.WebAppURL),
		})
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	b.sendWithKeyboard(chatID, "Главное меню:", kb)
}

func (b *Bot) showSchedule(chatID int64, tgID string) {
	from := time.Now().Format("2006-01-02")
	to := time.Now().AddDate(0, 0, 7).Format("2006-01-02")
	lessons, err := b.c.ListLessons(tgID, from, to)
	if err != nil || len(lessons) == 0 {
		b.send(chatID, "На ближайшую неделю тренировок нет 🗓")
		return
	}
	var sb strings.Builder
	sb.WriteString("📅 Расписание:\n")
	for _, l := range lessons {
		loc := ""
		if l.Location != nil {
			loc = " @ " + *l.Location
		}
		sb.WriteString(fmt.Sprintf("• %s %s%s\n", l.Date, l.Time, loc))
	}
	b.send(chatID, sb.String())
}

// ---------- Attendance (coach) ----------

func (b *Bot) attMenu(chatID int64, tgID string) {
	today := time.Now().Format("2006-01-02")
	lessons, err := b.c.ListLessons(tgID, today, today)
	if err != nil || len(lessons) == 0 {
		b.send(chatID, "На сегодня тренировок нет.")
		return
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	for _, l := range lessons {
		loc := ""
		if l.Location != nil {
			loc = " @ " + *l.Location
		}
		label := fmt.Sprintf("%s %s%s", l.Date, l.Time, loc)
		rows = append(rows, []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(label, "lesson:"+strconv.FormatInt(l.ID, 10)),
		})
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "menu"),
	})
	b.sendWithKeyboard(chatID, "Выберите тренировку:", tgbotapi.NewInlineKeyboardMarkup(rows...))
}

func (b *Bot) showAttendance(chatID int64, messageID int, lessonID int64, tgID string) {
	att, err := b.c.ListAttendance(tgID, lessonID)
	if err != nil {
		b.send(chatID, "Не удалось загрузить посещаемость.")
		return
	}
	if len(att) == 0 {
		b.send(chatID, "Нет записанных участников на эту тренировку.")
		return
	}
	var rows [][]tgbotapi.InlineKeyboardButton
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Посещаемость (тренировка #%d):\n", lessonID))
	for i, a := range att {
		mark := "✗"
		if a.Present {
			mark = "✓"
		}
		sb.WriteString(fmt.Sprintf("%d. Участник #%d %s\n", i+1, a.ClientID, mark))
		yes := tgbotapi.NewInlineKeyboardButtonData("✓", fmt.Sprintf("att:%d:%d:1", lessonID, a.ClientID))
		no := tgbotapi.NewInlineKeyboardButtonData("✗", fmt.Sprintf("att:%d:%d:0", lessonID, a.ClientID))
		rows = append(rows, []tgbotapi.InlineKeyboardButton{yes, no})
	}
	rows = append(rows, []tgbotapi.InlineKeyboardButton{
		tgbotapi.NewInlineKeyboardButtonData("📷 Фото", "photo:"+strconv.FormatInt(lessonID, 10)),
		tgbotapi.NewInlineKeyboardButtonData("⬅️ Назад", "att_menu"),
	})
	kb := tgbotapi.NewInlineKeyboardMarkup(rows...)
	text := sb.String()
	if messageID != 0 {
		edit := tgbotapi.NewEditMessageText(chatID, messageID, text)
		edit.ReplyMarkup = &kb
		_, _ = b.api.Send(edit)
	} else {
		b.sendWithKeyboard(chatID, text, kb)
	}
}

// ---------- Callbacks ----------

func (b *Bot) handleCallback(cb *tgbotapi.CallbackQuery) {
	data := cb.Data
	tgID := strconv.FormatInt(cb.From.ID, 10)
	chatID := cb.Message.Chat.ID
	msgID := cb.Message.MessageID

	ack := tgbotapi.NewCallback(cb.ID, "")
	_, _ = b.api.Request(ack)

	switch {
	case data == "menu":
		b.showMenu(chatID, tgID)
	case data == "schedule":
		b.showSchedule(chatID, tgID)
	case data == "subscription":
		b.send(chatID, "Оформить подписку можно в веб-приложении: откройте Профиль → Стать тренером.\nИли напишите администратору.")
	case data == "contact":
		b.mu.Lock()
		b.expectContact[chatID] = true
		b.mu.Unlock()
		b.send(chatID, "Напишите сообщение — оно будет передано тренеру:")
	case data == "att_menu":
		b.attMenu(chatID, tgID)
	case data == "photo_menu":
		b.send(chatID, "Откройте тренировку через «Отметить посещаемость» и нажмите 📷 Фото, затем пришлите фото.")
	case strings.HasPrefix(data, "lesson:"):
		id, _ := strconv.ParseInt(strings.TrimPrefix(data, "lesson:"), 10, 64)
		b.showAttendance(chatID, msgID, id, tgID)
	case strings.HasPrefix(data, "photo:"):
		id, _ := strconv.ParseInt(strings.TrimPrefix(data, "photo:"), 10, 64)
		b.mu.Lock()
		b.activeLesson[chatID] = id
		b.mu.Unlock()
		b.send(chatID, "Пришлите фото тренировки 📷")
case strings.HasPrefix(data, "att:"):
		parts := strings.Split(strings.TrimPrefix(data, "att:"), ":")
		if len(parts) == 3 {
			lessonID, _ := strconv.ParseInt(parts[0], 10, 64)
			clientID, _ := strconv.ParseInt(parts[1], 10, 64)
			present := parts[2] == "1"
			if err := b.c.MarkAttendance(tgID, lessonID, clientID, present); err != nil {
				b.send(chatID, "Не удалось отметить.")
			} else {
				b.showAttendance(chatID, msgID, lessonID, tgID)
			}
		}
	case strings.HasPrefix(data, "admin_menu"):
		b.showAdminMenu(chatID)
	case strings.HasPrefix(data, "admin_clients:"):
		page, _ := strconv.Atoi(strings.TrimPrefix(data, "admin_clients:"))
		b.showAdminClients(chatID, tgID, page)
	case strings.HasPrefix(data, "admin_client:"):
		clientID, _ := strconv.ParseInt(strings.TrimPrefix(data, "admin_client:"), 10, 64)
		b.showAdminClientDetail(chatID, tgID, clientID)
	case strings.HasPrefix(data, "admin_grant:"):
		clientID, _ := strconv.ParseInt(strings.TrimPrefix(data, "admin_grant:"), 10, 64)
		b.promptGrantDays(chatID, tgID, clientID)
	case strings.HasPrefix(data, "admin_revoke:"):
		clientID, _ := strconv.ParseInt(strings.TrimPrefix(data, "admin_revoke:"), 10, 64)
		b.revokeBotAccess(chatID, tgID, clientID)
	}
}

// ---------- Photo upload (coach) ----------

func (b *Bot) handlePhoto(chatID int64, tgID string, msg *tgbotapi.Message) {
	b.mu.Lock()
	lessonID := b.activeLesson[chatID]
	b.mu.Unlock()
	if lessonID == 0 {
		b.send(chatID, "Сначала откройте тренировку и нажмите 📷 Фото.")
		return
	}
	photos := msg.Photo
	fileID := photos[len(photos)-1].FileID
	fileURL, err := b.api.GetFileDirectURL(fileID)
	if err != nil {
		b.send(chatID, "Не удалось получить файл.")
		return
	}
	resp, err := http.Get(fileURL)
	if err != nil {
		b.send(chatID, "Ошибка загрузки фото.")
		return
	}
	defer resp.Body.Close()
	_ = os.MkdirAll("data/uploads", 0o755)
	path := fmt.Sprintf("data/uploads/lesson_%d_%d.jpg", lessonID, time.Now().Unix())
	out, err := os.Create(path)
	if err != nil {
		b.send(chatID, "Ошибка сохранения.")
		return
	}
	_, _ = io.Copy(out, resp.Body)
	out.Close()
	if _, err := b.c.UploadFile(tgID, "lesson", lessonID, "photo", path); err != nil {
		b.send(chatID, "Не удалось прикрепить фото к тренировке.")
		return
	}
	b.send(chatID, "Фото прикреплено к тренировке 📎")
}

// ---------- Notifier (polls outbox, sends via Telegram) ----------

func (b *Bot) runNotifier() {
	ticker := time.NewTicker(b.cfg.SchedulerInterval)
	for range ticker.C {
		ns, err := b.c.ClaimDue(50)
		if err != nil {
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
	text := renderNotification(n)
	if text == "" {
		return false
	}
	b.send(chatID, text)
	return true
}

func renderNotification(n domain.Notification) string {
	var p map[string]any
	_ = json.Unmarshal([]byte(n.Payload), &p)
	switch n.Type {
	case "new_client":
		return fmt.Sprintf("🔔 Новый клиент: %v", p["name"])
	case "client_message":
		return fmt.Sprintf("💬 Сообщение от %v:\n%v", p["from"], p["text"])
	case "lesson_reminder":
		return fmt.Sprintf("⏰ Напоминание: сегодня тренировка в %v (%v)", p["time"], orEmpty(p["location"]))
	case "lesson_canceled":
		return fmt.Sprintf("❌ Тренировка %v %v отменена", p["date"], p["time"])
	case "coach_broadcast":
		title, _ := p["title"].(string)
		text, _ := p["text"].(string)
		if title != "" {
			return fmt.Sprintf("📢 %s\n\n%s", title, text)
		}
		return text
	default:
		return fmt.Sprintf("Уведомление: %s", n.Type)
	}
}

func orEmpty(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

// ---------- low-level send ----------

func (b *Bot) send(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	_, _ = b.api.Send(msg)
}

func (b *Bot) sendWithKeyboard(chatID int64, text string, kb tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = kb
	_, _ = b.api.Send(msg)
}

// setWebAppMenuButton installs a persistent chat menu button that opens the Mini App.
// The bundled telegram-bot-api version predates MenuButton helpers, so we call the
// Bot API directly.
func (b *Bot) setWebAppMenuButton() error {
	base := b.cfg.BotAPIBaseURL
	if base == "" {
		base = tgbotapi.APIEndpoint
	} else {
		base = strings.TrimRight(base, "/") + "/bot%s/%s"
	}
	url := fmt.Sprintf(base, b.cfg.BotToken, "setChatMenuButton")

	body, _ := json.Marshal(map[string]any{
		"menu_button": map[string]any{
			"type":    "web_app",
			"text":    "📲 Приложение",
			"web_app": map[string]any{"url": b.cfg.WebAppURL},
		},
	})
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}
