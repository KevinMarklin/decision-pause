package bot

import (
	"strings"
	"testing"

	"github.com/max-messenger/max-bot-api-client-go/schemes"
)

func TestReplyFor(t *testing.T) {
	const deeplink = "https://max.ru/t797_hakaton_max_bot?startapp"

	tests := []struct {
		name       string
		cmd        string
		miniApp    string
		miniAppURL string
		wantText   string
		wantBtn    any // nil | schemes.OpenAppButton | schemes.LinkButton
	}{
		{name: "BotStarted без env", cmd: "", wantText: welcomeText, wantBtn: nil},
		{name: "start без env", cmd: "/start", wantText: welcomeText, wantBtn: nil},
		{name: "start с параметром", cmd: "/start hello", wantText: welcomeText, wantBtn: nil},
		{name: "start с суффиксом бота", cmd: "/start@my_bot", wantText: welcomeText, wantBtn: nil},
		{name: "start с именем аппа", cmd: "/start", miniApp: "MyApp", wantText: welcomeText,
			wantBtn: schemes.OpenAppButton{WebApp: "MyApp", Button: schemes.Button{Text: startButtonText, Type: schemes.OPEN_APP}}},
		{name: "start с deeplink", cmd: "/start", miniAppURL: deeplink, wantText: welcomeText,
			wantBtn: schemes.LinkButton{Url: deeplink, Button: schemes.Button{Text: startButtonText, Type: schemes.LINK}}},
		{name: "имя аппа важнее ссылки", cmd: "/start", miniApp: "MyApp", miniAppURL: deeplink,
			wantText: welcomeText,
			wantBtn:  schemes.OpenAppButton{WebApp: "MyApp", Button: schemes.Button{Text: startButtonText, Type: schemes.OPEN_APP}}},
		{name: "help", cmd: "/help", wantText: helpText, wantBtn: nil},
		{name: "help@bot", cmd: "/help@my_bot", wantText: helpText, wantBtn: nil},
		{name: "обычный текст", cmd: "привет", wantText: hintText, wantBtn: nil},
		{name: "неизвестная команда", cmd: "/go", wantText: hintText, wantBtn: nil},
		{name: "undefined от GetCommand", cmd: schemes.CommandUndefined, wantText: hintText, wantBtn: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, btn := replyFor(tt.cmd, tt.miniApp, tt.miniAppURL)
			if text != tt.wantText {
				t.Errorf("text:\n got %q\nwant %q", text, tt.wantText)
			}
			switch want := tt.wantBtn.(type) {
			case nil:
				if btn != nil {
					t.Errorf("btn = %#v, want nil", btn)
				}
			case schemes.OpenAppButton:
				got, ok := btn.(schemes.OpenAppButton)
				if !ok {
					t.Fatalf("btn type = %T, want OpenAppButton", btn)
				}
				if got.WebApp != want.WebApp || got.Type != schemes.OPEN_APP || got.Text != startButtonText {
					t.Errorf("btn = %#v, want %#v", got, want)
				}
			case schemes.LinkButton:
				got, ok := btn.(schemes.LinkButton)
				if !ok {
					t.Fatalf("btn type = %T, want LinkButton", btn)
				}
				if got.Url != want.Url || got.Type != schemes.LINK || got.Text != startButtonText {
					t.Errorf("btn = %#v, want %#v", got, want)
				}
			}
		})
	}
}

func TestWelcomeText(t *testing.T) {
	if len(strings.Split(welcomeText, "\n")) != 4 { // заголовок + пустая строка + 2 строки
		t.Errorf("welcomeText должен быть многострочным, got %q", welcomeText)
	}
	for _, want := range []string{"Анти-импульс", "последствия"} {
		if !strings.Contains(welcomeText, want) {
			t.Errorf("welcomeText не содержит %q", want)
		}
	}
}

func TestBuildMessageNoKeyboardWithoutButton(t *testing.T) {
	msg := buildMessage(1, hintText, nil)
	if msg == nil {
		t.Fatal("buildMessage вернул nil")
	}
}
