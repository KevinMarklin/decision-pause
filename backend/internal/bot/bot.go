package bot

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"

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

// Run — long polling бота. Блокирует до отмены ctx или ошибки клиента.
func Run(ctx context.Context, token string) error {
	client, err := httpClientWithMaxCA()
	if err != nil {
		return fmt.Errorf("http client: %w", err)
	}

	api, err := maxbot.New(token, maxbot.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("bot client: %w", err)
	}

	// проверка токена и связи с API на старте
	if _, err := api.Bots.GetBot(ctx); err != nil {
		return fmt.Errorf("bot api check: %w", err)
	}
	slog.Info("max bot: connected, long polling started")

	for update := range api.GetUpdates(ctx) {
		switch upd := update.(type) {
		case *schemes.BotStartedUpdate:
			send(ctx, api, upd.ChatId, welcomeMessage(upd.ChatId))
		case *schemes.MessageCreatedUpdate:
			if upd.GetCommand() == "/start" {
				send(ctx, api, upd.Message.Recipient.ChatId, welcomeMessage(upd.Message.Recipient.ChatId))
			} else {
				send(ctx, api, upd.Message.Recipient.ChatId,
					maxbot.NewMessage().SetChat(upd.Message.Recipient.ChatId).SetText(hintText))
			}
		}
	}
	return ctx.Err()
}

func welcomeMessage(chatID int64) *maxbot.Message {
	return maxbot.NewMessage().SetChat(chatID).SetText(welcomeText)
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
	return &http.Client{Transport: transport}, nil
}

func send(ctx context.Context, api *maxbot.Api, chatID int64, msg *maxbot.Message) {
	if err := api.Messages.Send(ctx, msg); err != nil {
		slog.Warn("bot: send failed", "chat_id", chatID, "err", err)
	}
}
