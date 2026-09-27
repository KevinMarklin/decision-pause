package bot

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	maxbot "github.com/max-messenger/max-bot-api-client-go"
	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

// Корневой сертификат Минцифры — им подписан platform-api2.max.ru.
//
//go:embed russian_trusted_root_ca.pem
var maxRootCA []byte

const welcomeText = "🧠 Анти-импульс\n\n" +
	"Пауза перед важным бизнес-решением.\n" +
	"Мы не принимаем решение за вас — показываем возможные последствия."

const hintText = "Отправьте /start, чтобы начать анализ."

const helpText = "🧠 Анти-импульс — справка\n\n" +
	"/start — приветствие и кнопка анализа\n" +
	"/help — эта справка\n\n" +
	"Дальше путь в Mini App: анкета из 9 полей → расчёт платежа и трёх " +
	"сценариев → последствия и чек-лист. Мы не говорим «делай/не делай» — " +
	"показываем, к чему приведёт решение."

// replyFor — чистая функция: команда (GetCommand) → (текст, кнопка или nil).
// cmd == "" — пользователь только что открыл бота (BotStartedUpdate).
func replyFor(cmd, miniApp, miniAppURL string) (string, schemes.ButtonInterface) {
	// "/start что-то" → "/start": учитываем только первое слово
	if fields := strings.Fields(cmd); len(fields) > 0 {
		cmd = fields[0]
	}
	switch {
	case cmd == "" || cmd == "/start" || strings.HasPrefix(cmd, "/start@"):
		return welcomeText, startButton(miniApp, miniAppURL)
	case cmd == "/help" || strings.HasPrefix(cmd, "/help@"):
		return helpText, startButton(miniApp, miniAppURL)
	default:
		return hintText, nil
	}
}

// buildMessage — собирает исходящее сообщение с опциональной кнопкой.
func buildMessage(chatID int64, text string, btn schemes.ButtonInterface) *maxbot.Message {
	msg := maxbot.NewMessage().SetChat(chatID).SetText(text)
	if btn != nil {
		msg.AddKeyboard(maxbot.InlineKeyboard(maxbot.Row(btn)))
	}
	return msg
}

// Run — long polling бота. Блокирует до отмены ctx или ошибки клиента.
// miniApp / miniAppURL — кнопка «Начать анализ»: open_app либо ссылка-фолбэк.
func Run(ctx context.Context, token, miniApp, miniAppURL string) error {
	client, err := httpClientWithMaxCA()
	if err != nil {
		return fmt.Errorf("http client: %w", err)
	}

	api, err := maxbot.New(token, maxbot.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("bot client: %w", err)
	}

	// проверка токена и связи с API на старте
	botInfo, err := api.Bots.GetBot(ctx)
	if err != nil {
		return fmt.Errorf("bot api check: %w", err)
	}
	slog.Info("max bot: connected, long polling started", "name", botInfo.Name, "username", botInfo.Username)

	// меню команд в интерфейсе чата (/start, /help) — не фатально при ошибке.
	// Прямой вызов: библиотека шлёт PATCH /me, который API больше не знает,
	// актуальный метод — PATCH /me/commands (см. dev.max.ru/docs-api).
	if err := setCommands(ctx, client, token); err != nil {
		slog.Warn("bot: set commands menu failed", "err", err)
	} else {
		slog.Info("bot: commands menu updated")
	}

	// ошибки long-poll/API иначе копятся в канале молча (буфер по умолчанию).
	go func() {
		for err := range api.GetErrors() {
			slog.Warn("bot: api error", "err", err)
		}
	}()

	for update := range api.GetUpdates(ctx) {
		slog.Info("bot: update received", "type", fmt.Sprintf("%T", update))
		switch upd := update.(type) {
		case *schemes.BotStartedUpdate:
			slog.Info("bot: started by user", "chat_id", upd.ChatId)
			text, btn := replyFor("", miniApp, miniAppURL)
			send(ctx, api, upd.ChatId, buildMessage(upd.ChatId, text, btn))
		case *schemes.MessageCreatedUpdate:
			chatID := upd.Message.Recipient.ChatId
			cmd := upd.GetCommand()
			slog.Info("bot: message", "chat_id", chatID,
				"chat_type", string(upd.Message.Recipient.ChatType), "command", cmd)
			if upd.Message.Recipient.ChatType != schemes.DIALOG {
				continue // отвечаем только в личке
			}
			text, btn := replyFor(cmd, miniApp, miniAppURL)
			send(ctx, api, chatID, buildMessage(chatID, text, btn))
		}
	}
	return ctx.Err()
}

const startButtonText = "▶️ Начать анализ"

// startButton — open_app, если задано имя мини-аппа, иначе deeplink-ссылка.
func startButton(miniApp, miniAppURL string) schemes.ButtonInterface {
	if miniApp != "" {
		return maxbot.BtnApp(startButtonText, miniApp, "", 0)
	}
	if miniAppURL != "" {
		return maxbot.BtnLink(startButtonText, miniAppURL)
	}
	return nil
}

// maxAPIBase — актуальный домен API MAX (platform-api устарел).
const maxAPIBase = "https://platform-api2.max.ru"

// setCommands — PATCH /me/commands: команды в подсказке «/» у бота.
func setCommands(ctx context.Context, client *http.Client, token string) error {
	body, err := json.Marshal(map[string]any{"commands": []schemes.BotCommand{
		{Name: "start", Description: "Начать анализ"},
		{Name: "help", Description: "Справка по боту"},
	}})
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, maxAPIBase+"/me/commands", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("PATCH /me/commands: %s: %s", resp.Status, msg)
	}
	return nil
}

// httpClientWithMaxCA — системный пул корневых сертификатов + сертификат Минцифры.
func httpClientWithMaxCA() (*http.Client, error) {
	pool, err := x509.SystemCertPool()
	if err != nil || pool == nil {
		pool = x509.NewCertPool()
	}
	if !pool.AppendCertsFromPEM(maxRootCA) {
		return nil, fmt.Errorf("cannot parse max root ca")
	}
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: pool, MinVersion: tls.VersionTLS12}
	// 40с > 30с long-poll: обычный цикл успевает, а мёртвое соединение
	// обрывается по таймауту, а не висит бесконечно (Timeout=0 грозил зависанием).
	return &http.Client{Transport: transport, Timeout: 40 * time.Second}, nil
}

func send(ctx context.Context, api *maxbot.Api, chatID int64, msg *maxbot.Message) {
	res, err := api.Messages.SendWithResult(ctx, msg)
	if err != nil {
		slog.Warn("bot: send failed", "chat_id", chatID, "err", err)
		return
	}
	// что реально сохранил сервер (проверка обрезки многострочного текста)
	slog.Info("bot: sent", "chat_id", chatID, "text", res.Body.Text)
}
