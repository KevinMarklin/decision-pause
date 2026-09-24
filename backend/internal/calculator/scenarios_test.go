package calculator

import (
	"testing"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

// эталонный вход из плана MVP
func baseInput() model.Input {
	return model.Input{
		LoanAmount:       2_000_000,
		LoanTermMonths:   36,
		InterestRatePct:  20,
		Revenue:          800_000,
		Expenses:         600_000,
		Reserve:          500_000,
		RevenueGrowthPct: 30,
		ExpenseGrowthPct: 15,
		Purpose:          "Закупка оборудования",
	}
}

func TestBuildScenarios_Etalon(t *testing.T) {
	in := baseInput()
	got := BuildScenarios(in, CreditResult(in))

	if len(got) != 3 {
		t.Fatalf("сценариев = %d, want 3", len(got))
	}

	expected := []struct {
		key            string
		revenue        float64
		cashFlow       float64
		debtLoad       float64
		reserveAfter3M float64
		reserveMonths  *int
	}{
		{"expected", 1_040_000, 275_673, 7.1, 1_327_019, nil},
		{"moderate", 936_000, 171_673, 7.9, 1_015_019, nil},
		{"negative", 728_000, -36_327, 10.2, 391_019, intPtr(13)},
	}

	for i, want := range expected {
		s := got[i]
		if s.Key != want.key {
			t.Errorf("[%d] key = %q, want %q", i, s.Key, want.key)
		}
		if s.Revenue != want.revenue {
			t.Errorf("[%s] revenue = %v, want %v", s.Key, s.Revenue, want.revenue)
		}
		if s.Expenses != 690_000 {
			t.Errorf("[%s] expenses = %v, want 690000", s.Key, s.Expenses)
		}
		if s.LoanPayment != 74_327 {
			t.Errorf("[%s] loan_payment = %v, want 74327", s.Key, s.LoanPayment)
		}
		if s.CashFlow != want.cashFlow {
			t.Errorf("[%s] cash_flow = %v, want %v", s.Key, s.CashFlow, want.cashFlow)
		}
		if s.DebtLoadPct != want.debtLoad {
			t.Errorf("[%s] debt_load = %v, want %v", s.Key, s.DebtLoadPct, want.debtLoad)
		}
		if s.ReserveAfter3M != want.reserveAfter3M {
			t.Errorf("[%s] reserve_after_3m = %v, want %v", s.Key, s.ReserveAfter3M, want.reserveAfter3M)
		}
		if (s.ReserveMonths == nil) != (want.reserveMonths == nil) {
			t.Errorf("[%s] reserve_months = %v, want %v", s.Key, s.ReserveMonths, want.reserveMonths)
		} else if s.ReserveMonths != nil && *s.ReserveMonths != *want.reserveMonths {
			t.Errorf("[%s] reserve_months = %d, want %d", s.Key, *s.ReserveMonths, *want.reserveMonths)
		}
	}
}

func TestBuildScenarios_ModerateFactor(t *testing.T) {
	in := baseInput()
	got := BuildScenarios(in, CreditResult(in))

	if got[1].Revenue != roundMoney(got[0].Revenue*ModerateFactor) {
		t.Errorf("moderate revenue = %v, want %v", got[1].Revenue, roundMoney(got[0].Revenue*ModerateFactor))
	}
	if got[2].Revenue != roundMoney(got[0].Revenue*NegativeFactor) {
		t.Errorf("negative revenue = %v, want %v", got[2].Revenue, roundMoney(got[0].Revenue*NegativeFactor))
	}
}

func TestBuildScenarios_NegativeReserveExhausted(t *testing.T) {
	in := baseInput()
	in.Reserve = 100_000 // маленький резерв при дефиците 36 327 ₽/мес
	got := BuildScenarios(in, CreditResult(in))
	neg := got[2]

	if neg.CashFlow >= 0 {
		t.Fatalf("cash_flow = %v, want < 0", neg.CashFlow)
	}
	if neg.ReserveAfter3M >= 0 {
		t.Errorf("reserve_after_3m = %v, want < 0", neg.ReserveAfter3M)
	}
	if *neg.ReserveMonths != 2 {
		t.Errorf("reserve_months = %d, want 2", *neg.ReserveMonths)
	}
}

func TestBuildResult(t *testing.T) {
	got := BuildResult(baseInput())

	if got.Checklist == nil || len(got.Checklist) != 5 {
		t.Fatalf("checklist = %v, want 5 пунктов", got.Checklist)
	}
	if len(got.Scenarios) != 3 {
		t.Fatalf("сценариев = %d, want 3", len(got.Scenarios))
	}
	if got.Credit.MonthlyPayment != 74_327 {
		t.Errorf("credit.monthly_payment = %v, want 74327", got.Credit.MonthlyPayment)
	}
}

func intPtr(v int) *int { return &v }
