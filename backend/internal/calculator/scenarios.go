package calculator

import (
	"math"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

// Коэффициенты сценариев — доля ожидаемой выручки.
// Зафиксированы в README, меняются только осознанно.
const (
	ModerateFactor = 0.9
	NegativeFactor = 0.7
)

// BuildScenarios — три сценария: ожидаемый (прогноз пользователя),
// умеренный и негативный. Расходы и платёж одинаковы во всех сценариях:
// расходы уже приняты, платёж фиксирован графиком.
func BuildScenarios(in model.Input, credit model.CreditResult) []model.Scenario {
	expectedRevenue := roundMoney(in.Revenue * (1 + in.RevenueGrowthPct/100))
	expectedExpenses := roundMoney(in.Expenses * (1 + in.ExpenseGrowthPct/100))

	defs := []struct {
		key     string
		revenue float64
	}{
		{"expected", expectedRevenue},
		{"moderate", roundMoney(expectedRevenue * ModerateFactor)},
		{"negative", roundMoney(expectedRevenue * NegativeFactor)},
	}

	scenarios := make([]model.Scenario, 0, len(defs))
	for _, d := range defs {
		title, emoji := model.ScenarioTitle(d.key)
		s := model.Scenario{
			Key:         d.key,
			Title:       title,
			Emoji:       emoji,
			Revenue:     d.revenue,
			Expenses:    expectedExpenses,
			LoanPayment: credit.MonthlyPayment,
		}
		s.CashFlow = roundMoney(s.Revenue - s.Expenses - s.LoanPayment)
		if s.Revenue > 0 {
			s.DebtLoadPct = round1(s.LoanPayment / s.Revenue * 100)
		}
		s.ReserveAfter3M = roundMoney(in.Reserve + 3*s.CashFlow)
		if s.CashFlow < 0 {
			months := int(math.Floor(in.Reserve / math.Abs(s.CashFlow)))
			s.ReserveMonths = &months
		}
		s.Consequences = Consequences(s)
		scenarios = append(scenarios, s)
	}
	return scenarios
}

// BuildResult — вся логика расчёта одним вызовом (точка входа для service).
func BuildResult(in model.Input) model.DecisionResult {
	credit := CreditResult(in)
	return model.DecisionResult{
		Credit:    credit,
		Scenarios: BuildScenarios(in, credit),
		Checklist: Checklist(),
	}
}
