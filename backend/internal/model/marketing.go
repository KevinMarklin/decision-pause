package model

// MarketingInput — данные анкеты рекламной кампании.
type MarketingInput struct {
	Budget           float64 `json:"budget"`
	AvgCheck         float64 `json:"avg_check"`
	MarginPct        float64 `json:"margin_pct"`
	LeadCost         float64 `json:"lead_cost"`
	ExpectedBoostPct float64 `json:"expected_boost_pct"`
	CurrentRevenue   float64 `json:"current_revenue"`
	FixedExpenses    float64 `json:"fixed_expenses"`
	Reserve          float64 `json:"reserve"`
}

type MarketingScenario struct {
	Key             string  `json:"key"`
	Title           string  `json:"title"`
	Emoji           string  `json:"emoji"`
	AdditionalOrders float64 `json:"additional_orders"`
	AdRevenue       float64 `json:"ad_revenue"`
	AdProfit        float64 `json:"ad_profit"`
	ROI             float64 `json:"roi"`
	CashFlow        float64 `json:"cash_flow"`
	Status          string  `json:"status"`
	ReserveMonths   *int    `json:"reserve_months"`
	Consequence     string  `json:"consequence"`
}

type MarketingResult struct {
	BreakEvenOrders  int                `json:"break_even_orders"`
	BreakEvenRevenue float64            `json:"break_even_revenue"`
	Scenarios        []MarketingScenario `json:"scenarios"`
	Consequences     []string           `json:"consequences"`
	Checklist        []string           `json:"checklist"`
}
