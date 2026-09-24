package calculator

import (
	"strings"
	"testing"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

func TestConsequences_PositiveFlow(t *testing.T) {
	s := model.Scenario{Key: "expected", Revenue: 1_040_000, LoanPayment: 74_327, CashFlow: 275_673, DebtLoadPct: 7.1}
	got := consequences(s)

	if len(got) != 1 {
		t.Fatalf("consequences = %v, want 1", got)
	}
	if !strings.Contains(got[0], "свободный поток +275 673 ₽") {
		t.Errorf("нет текста о свободном потоке: %q", got[0])
	}
}

func TestConsequences_DeficitWithReserve(t *testing.T) {
	months := 13
	s := model.Scenario{
		Key: "negative", Revenue: 728_000, LoanPayment: 74_327,
		CashFlow: -36_327, DebtLoadPct: 10.2,
		ReserveAfter3M: 391_019, ReserveMonths: &months,
	}
	got := consequences(s)

	if len(got) != 2 {
		t.Fatalf("consequences = %v, want 2", got)
	}
	if !strings.Contains(got[0], "разница покрывается из резерва") {
		t.Errorf("нет текста о дефиците: %q", got[0])
	}
	if !strings.Contains(got[1], "на 13 мес.") {
		t.Errorf("нет текста о сроке резерва: %q", got[1])
	}
}

func TestConsequences_ReserveExhausted(t *testing.T) {
	months := 1
	s := model.Scenario{
		Key: "negative", Revenue: 728_000, LoanPayment: 74_327,
		CashFlow: -36_327, DebtLoadPct: 10.2,
		ReserveAfter3M: -9_000, ReserveMonths: &months,
	}
	got := consequences(s)

	found := false
	for _, c := range got {
		if strings.Contains(c, "через 3 мес. резерв будет исчерпан") {
			found = true
		}
	}
	if !found {
		t.Errorf("нет текста об исчерпании резерва: %v", got)
	}
}

func TestConsequences_HighDebtLoad(t *testing.T) {
	s := model.Scenario{Key: "expected", Revenue: 300_000, LoanPayment: 74_327, CashFlow: 25_673, DebtLoadPct: 24.8}
	got := consequences(s)

	found := false
	for _, c := range got {
		if strings.Contains(c, "занимает 24.8% выручки") {
			found = true
		}
	}
	if !found {
		t.Errorf("нет текста о нагрузке: %v", got)
	}
}

func TestConsequences_NoAdviceWords(t *testing.T) {
	// сервис не говорит пользователю, что делать
	forbidden := []string{"вам нельзя", "не берите", "берите", "советуем", "рекомендуем взять", "решайте взять"}
	months := 2
	s := model.Scenario{
		Key: "negative", Revenue: 728_000, LoanPayment: 74_327,
		CashFlow: -36_327, DebtLoadPct: 20.5,
		ReserveAfter3M: -9_000, ReserveMonths: &months,
	}
	for _, c := range consequences(s) {
		lower := strings.ToLower(c)
		for _, f := range forbidden {
			if strings.Contains(lower, f) {
				t.Errorf("формулировка содержит запрещённое слово %q: %q", f, c)
			}
		}
	}
}

func TestFmtAmount(t *testing.T) {
	cases := map[float64]string{
		0:         "0",
		999:       "999",
		1_000:     "1 000",
		275_673:   "275 673",
		-36_327:   "36 327",
		1_327_019: "1 327 019",
	}
	for in, want := range cases {
		if got := fmtAmount(in); got != want {
			t.Errorf("fmtAmount(%v) = %q, want %q", in, got, want)
		}
	}
}
