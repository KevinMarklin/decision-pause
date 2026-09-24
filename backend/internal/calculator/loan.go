package calculator

import (
	"math"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

// AnnuityPayment — ежемесячный аннуитетный платёж.
// Ставка 0% → равные доли тела кредита.
func AnnuityPayment(amount, ratePct float64, months int) float64 {
	if months <= 0 || amount == 0 {
		return 0
	}
	if ratePct == 0 {
		return amount / float64(months)
	}
	i := ratePct / 1200
	return amount * i / (1 - math.Pow(1+i, float64(-months)))
}

// CreditResult — сводка по кредиту. Деньги округляются до целого рубля,
// платёж — один раз, остальные показатели считаются от него.
func CreditResult(in model.Input) model.CreditResult {
	payment := roundMoney(AnnuityPayment(in.LoanAmount, in.InterestRatePct, in.LoanTermMonths))
	total := roundMoney(payment * float64(in.LoanTermMonths))
	return model.CreditResult{
		MonthlyPayment: payment,
		TotalPaid:      total,
		Overpay:        roundMoney(total - in.LoanAmount),
	}
}
