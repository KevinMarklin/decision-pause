package calculator

import (
	"testing"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

func TestCreditResult_StandardAnnuity(t *testing.T) {
	in := model.Input{LoanAmount: 2_000_000, LoanTermMonths: 36, InterestRatePct: 20}
	got := CreditResult(in)

	if got.MonthlyPayment != 74_327 {
		t.Errorf("MonthlyPayment = %v, want 74327", got.MonthlyPayment)
	}
	if got.TotalPaid != 2_675_772 {
		t.Errorf("TotalPaid = %v, want 2675772", got.TotalPaid)
	}
	if got.Overpay != 675_772 {
		t.Errorf("Overpay = %v, want 675772", got.Overpay)
	}
}

func TestCreditResult_ZeroRate(t *testing.T) {
	in := model.Input{LoanAmount: 2_000_000, LoanTermMonths: 36, InterestRatePct: 0}
	got := CreditResult(in)

	if got.MonthlyPayment != 55_556 {
		t.Errorf("MonthlyPayment = %v, want 55556", got.MonthlyPayment)
	}
	if got.Overpay != 16 {
		t.Errorf("Overpay = %v, want 16 (артефакт округления)", got.Overpay)
	}
}

func TestCreditResult_InvalidTerm(t *testing.T) {
	in := model.Input{LoanAmount: 1_000_000, LoanTermMonths: 0, InterestRatePct: 20}
	if got := CreditResult(in); got.MonthlyPayment != 0 {
		t.Errorf("MonthlyPayment = %v, want 0", got.MonthlyPayment)
	}
}

func TestAnnuityPayment_Deterministic(t *testing.T) {
	a := AnnuityPayment(1_000_000, 25, 24)
	b := AnnuityPayment(1_000_000, 25, 24)
	if a != b {
		t.Errorf("расчёт недетерминирован: %v != %v", a, b)
	}
	if a <= 0 {
		t.Errorf("платёж должен быть > 0, got %v", a)
	}
}
