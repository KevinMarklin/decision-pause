package calculator

import (
	"fmt"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

// consequences — правила → тексты последствий.
// Формулировки констатируют факт, не содержат советов «делай / не делай».
func consequences(s model.Scenario) []string {
	var out []string

	if s.CashFlow < 0 {
		out = append(out, fmt.Sprintf(
			"Расходы и платёж превышают выручку на %s ₽ в месяц — разница покрывается из резерва.",
			fmtAmount(-s.CashFlow)))
		if s.ReserveMonths != nil {
			out = append(out, fmt.Sprintf(
				"Резерва хватит примерно на %d мес. при сохранении такого дефицита.",
				*s.ReserveMonths))
		}
		if s.ReserveAfter3M < 0 {
			out = append(out, "Если сохранится такой темп, через 3 мес. резерв будет исчерпан.")
		}
	} else {
		out = append(out, fmt.Sprintf(
			"Платёж покрывается из текущей выручки, свободный поток +%s ₽/мес.",
			fmtAmount(s.CashFlow)))
	}

	if s.DebtLoadPct > 15 {
		out = append(out, fmt.Sprintf(
			"Платёж по кредиту занимает %.1f%% выручки.", s.DebtLoadPct))
	}
	return out
}
