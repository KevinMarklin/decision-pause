package model

// HiringInput — параметры оценки найма сотрудника.
type HiringInput struct {
	Salary         float64 `json:"salary"`
	TaxRate        float64 `json:"tax_rate,omitempty"`
	OverheadCosts  float64 `json:"overhead_costs,omitempty"`
	MarginPercent  float64 `json:"margin_percent"`
	AverageCheck   float64 `json:"average_check,omitempty"`
}

// HiringResult — стоимость сотрудника и точка безубыточности найма.
type HiringResult struct {
	Salary            float64 `json:"salary"`
	TaxRate           float64 `json:"tax_rate"`
	OverheadCosts     float64 `json:"overhead_costs"`
	TotalCost         float64 `json:"total_cost"`
	MarginPercent     float64 `json:"margin_percent"`
	BreakEvenRevenue  float64 `json:"break_even_revenue"`
	BreakEvenDeals    int     `json:"break_even_deals"`
	PaybackMultiplier float64 `json:"payback_multiplier"`
}
