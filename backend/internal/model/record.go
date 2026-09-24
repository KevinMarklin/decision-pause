package model

import "time"

// Decision — сохранённый анализ: входные данные + результат расчёта.
type Decision struct {
	ID        string
	MaxUserID *int64
	Type      string
	Status    string
	CreatedAt time.Time
	Inputs    Input
	Result    DecisionResult
}

// DecisionListItem — краткая запись для истории.
type DecisionListItem struct {
	ID             string
	CreatedAt      time.Time
	Type           string
	LoanAmount     float64
	MonthlyPayment float64
	CashFlows      map[string]float64
}
