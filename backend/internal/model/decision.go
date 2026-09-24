package model

// Input — данные анкеты по решению «Взять кредит на развитие бизнеса».
type Input struct {
	LoanAmount       float64 `json:"loan_amount"`
	LoanTermMonths   int     `json:"loan_term_months"`
	InterestRatePct  float64 `json:"interest_rate_pct"`
	Revenue          float64 `json:"revenue"`
	Expenses         float64 `json:"expenses"`
	Reserve          float64 `json:"reserve"`
	RevenueGrowthPct float64 `json:"revenue_growth_pct"`
	ExpenseGrowthPct float64 `json:"expense_growth_pct"`
	Purpose          string  `json:"purpose"`
}

// CreditResult — показатели кредита.
type CreditResult struct {
	MonthlyPayment float64 `json:"monthly_payment"`
	TotalPaid      float64 `json:"total_paid"`
	Overpay        float64 `json:"overpay"`
}

// Scenario — один рассчитанный сценарий.
type Scenario struct {
	Key            string   `json:"key"`
	Title          string   `json:"title"`
	Emoji          string   `json:"emoji"`
	Revenue        float64  `json:"revenue"`
	Expenses       float64  `json:"expenses"`
	LoanPayment    float64  `json:"loan_payment"`
	CashFlow       float64  `json:"cash_flow"`
	DebtLoadPct    float64  `json:"debt_load_pct"`
	ReserveAfter3M float64  `json:"reserve_after_3m"`
	ReserveMonths  *int     `json:"reserve_months"`
	Consequences   []string `json:"consequences"`
}

// DecisionResult — полный результат анализа.
type DecisionResult struct {
	Credit    CreditResult `json:"credit"`
	Scenarios []Scenario   `json:"scenarios"`
	Checklist []string     `json:"checklist"`
}
