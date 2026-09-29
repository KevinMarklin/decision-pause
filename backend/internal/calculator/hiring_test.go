package calculator

import (
	"math"
	"testing"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

func TestCalculateHiringDefaultTax(t *testing.T) {
	got := CalculateHiring(model.HiringInput{Salary: 100000, MarginPercent: 30, AverageCheck: 25000})
	if got.TotalCost != 130000 { t.Fatalf("total cost = %.2f, want 130000", got.TotalCost) }
	if math.Abs(got.BreakEvenRevenue-433333.3333333333) > 0.000001 { t.Fatalf("break-even = %.2f", got.BreakEvenRevenue) }
	if got.BreakEvenDeals != 18 { t.Fatalf("deals = %d, want 18", got.BreakEvenDeals) }
}

func TestCalculateHiringWithOverhead(t *testing.T) {
	got := CalculateHiring(model.HiringInput{Salary: 100000, TaxRate: 0.20, OverheadCosts: 15000, MarginPercent: 40, AverageCheck: 30000})
	if got.TotalCost != 135000 { t.Fatalf("total cost = %.2f, want 135000", got.TotalCost) }
	if got.BreakEvenDeals != 12 { t.Fatalf("deals = %d, want 12", got.BreakEvenDeals) }
}

func TestCalculateHiringWithoutAverageCheckAndZeroMargin(t *testing.T) {
	got := CalculateHiring(model.HiringInput{Salary: 100000, MarginPercent: 0})
	if got.BreakEvenRevenue != 0 || got.BreakEvenDeals != 0 || got.PaybackMultiplier != 0 { t.Fatalf("zero margin result = %+v", got) }
}
