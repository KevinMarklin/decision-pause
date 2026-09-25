package config

import (
	"strings"
	"testing"
)

func TestValidate_RequireWithoutTokenFails(t *testing.T) {
	cfg := Config{RequireInitData: true}
	err := cfg.Validate()
	if err == nil {
		t.Fatal("ожидалась ошибка: Require без токена")
	}
	if !strings.Contains(err.Error(), "MAX_BOT_TOKEN") {
		t.Errorf("сообщение не подсказывает причину: %v", err)
	}
}

func TestValidate_Ok(t *testing.T) {
	cases := []Config{
		{},
		{MaxBotToken: "token"},
		{RequireInitData: true, MaxBotToken: "token"},
	}
	for _, cfg := range cases {
		if err := cfg.Validate(); err != nil {
			t.Errorf("%+v: %v", cfg, err)
		}
	}
}

func TestEnvBool(t *testing.T) {
	cases := map[string]bool{
		"true": true, "TRUE": true, "1": true, "on": true, "yes": true,
		"false": false, "0": false, "off": false, "no": false,
	}
	for raw, want := range cases {
		t.Setenv("DP_TEST_BOOL", raw)
		if got := envBool("DP_TEST_BOOL", !want); got != want {
			t.Errorf("envBool(%q) = %v, want %v", raw, got, want)
		}
	}

	t.Setenv("DP_TEST_BOOL", "maybe")
	if !envBool("DP_TEST_BOOL", true) {
		t.Error("нечитаемое значение должно давать fallback")
	}
	if envBool("DP_TEST_MISSING", false) {
		t.Error("отсутствующий ключ должен давать fallback")
	}
}

func TestSplitList(t *testing.T) {
	cases := map[string][]string{
		"":                            nil,
		"a":                           {"a"},
		"a, b ,c":                     {"a", "b", "c"},
		"https://x.ru,,https://y.ru,": {"https://x.ru", "https://y.ru"},
	}
	for raw, want := range cases {
		got := splitList(raw)
		if len(got) != len(want) {
			t.Errorf("splitList(%q) = %v, want %v", raw, got, want)
			continue
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("splitList(%q)[%d] = %q, want %q", raw, i, got[i], want[i])
			}
		}
	}
}
