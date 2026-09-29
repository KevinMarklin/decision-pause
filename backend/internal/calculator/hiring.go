package calculator

import (
	"math"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

const defaultHiringTaxRate = 0.30

// CalculateHiring calculates the full monthly employee cost and break-even.
func CalculateHiring(in model.HiringInput) model.HiringResult {
	taxRate := in.TaxRate
	if taxRate <= 0 {
		taxRate = defaultHiringTaxRate
	}
	totalCost := in.Salary*(1+taxRate) + in.OverheadCosts
	result := model.HiringResult{
		Salary:        in.Salary,
		TaxRate:       taxRate,
		OverheadCosts: in.OverheadCosts,
		TotalCost:     totalCost,
		MarginPercent: in.MarginPercent,
	}
	if in.MarginPercent > 0 {
		result.BreakEvenRevenue = totalCost / (in.MarginPercent / 100)
	}
	if in.AverageCheck > 0 && result.BreakEvenRevenue > 0 {
		result.BreakEvenDeals = int(math.Ceil(result.BreakEvenRevenue / in.AverageCheck))
	}
	if in.Salary > 0 {
		result.PaybackMultiplier = result.BreakEvenRevenue / in.Salary
	}
	return result
}
