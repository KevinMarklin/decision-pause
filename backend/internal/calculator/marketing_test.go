package calculator

import (
	"testing"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

func TestMarketingResult(t *testing.T) {
	in := model.MarketingInput{Budget: 300000, AvgCheck: 10000, MarginPct: 40, LeadCost: 1500, ExpectedBoostPct: 20, CurrentRevenue: 800000, FixedExpenses: 600000, Reserve: 500000}
	got := MarketingResult(in)
	if got.BreakEvenOrders != 75 || got.BreakEvenRevenue != 750000 { t.Fatalf("break-even = %d / %.0f", got.BreakEvenOrders, got.BreakEvenRevenue) }
	if len(got.Scenarios) != 3 { t.Fatalf("scenarios = %d, want 3", len(got.Scenarios)) }
	if got.Scenarios[0].AdditionalOrders != 240 { t.Errorf("expected orders = %.2f", got.Scenarios[0].AdditionalOrders) }
	if got.Scenarios[0].AdProfit != 660000 { t.Errorf("expected ad profit = %.0f", got.Scenarios[0].AdProfit) }
	if got.Scenarios[2].Status != "yellow" { t.Errorf("negative status = %q, want yellow", got.Scenarios[2].Status) }
}

func TestMarketingReserveMonths(t *testing.T) {
	in := model.MarketingInput{Budget: 300000, AvgCheck: 10000, MarginPct: 40, LeadCost: 1500, ExpectedBoostPct: -100, CurrentRevenue: 100000, FixedExpenses: 900000, Reserve: 500000}
	got := MarketingResult(in).Scenarios[2]
	if got.Status != "red" || got.ReserveMonths == nil || *got.ReserveMonths != 1 { t.Fatalf("red reserve = %q/%v", got.Status, got.ReserveMonths) }
}
