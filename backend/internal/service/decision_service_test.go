package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
	"github.com/KevinMarklin/decision-pause/backend/internal/repository"
)

// fakeStore — фейк хранилища: записывает вызовы, отдаёт заготовленные ошибки.
type fakeStore struct {
	createErr error
	getErr    error
	listErr   error
	deleteErr error
	pingErr   error

	// deleteOK — результат операции удаления.
	deleteOK bool

	// getDecision — снимок, который вернёт Get.
	getDecision model.Decision

	created  []model.Decision
	deleted  string
	ownerSet *int64
	gotLimit int
	gotUser  *int64
}

func (f *fakeStore) Create(_ context.Context, d *model.Decision) error {
	if f.createErr != nil {
		return f.createErr
	}
	f.created = append(f.created, *d)
	return nil
}

func (f *fakeStore) Get(context.Context, string) (model.Decision, error) {
	return f.getDecision, f.getErr
}

func (f *fakeStore) List(_ context.Context, userID *int64, limit int) ([]model.DecisionListItem, error) {
	f.gotUser = userID
	f.gotLimit = limit
	if f.listErr != nil {
		return nil, f.listErr
	}
	return nil, nil
}

func (f *fakeStore) Delete(_ context.Context, id string, userID *int64) (bool, error) {
	f.deleted = id
	f.ownerSet = userID
	if f.deleteErr != nil {
		return false, f.deleteErr
	}
	return f.deleteOK, nil
}

func (f *fakeStore) Ping(context.Context) error { return f.pingErr }

func validInput() model.Input {
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

// --- Validate ----------------------------------------------------------------

func TestValidate_ValidInput(t *testing.T) {
	if fields := Validate(validInput()); fields != nil {
		t.Errorf("ожидалось nil, получено %v", fields)
	}
}

func TestValidate_FieldErrors(t *testing.T) {
	cases := []struct {
		name  string
		mut   func(*model.Input)
		field string
	}{
		{"loan_amount_zero", func(in *model.Input) { in.LoanAmount = 0 }, "loan_amount"},
		{"loan_amount_negative", func(in *model.Input) { in.LoanAmount = -1 }, "loan_amount"},
		{"loan_term_low", func(in *model.Input) { in.LoanTermMonths = 0 }, "loan_term_months"},
		{"loan_term_high", func(in *model.Input) { in.LoanTermMonths = 361 }, "loan_term_months"},
		{"rate_negative", func(in *model.Input) { in.InterestRatePct = -1 }, "interest_rate_pct"},
		{"rate_over_100", func(in *model.Input) { in.InterestRatePct = 101 }, "interest_rate_pct"},
		{"revenue_zero", func(in *model.Input) { in.Revenue = 0 }, "revenue"},
		{"expenses_negative", func(in *model.Input) { in.Expenses = -1 }, "expenses"},
		{"reserve_negative", func(in *model.Input) { in.Reserve = -1 }, "reserve"},
		{"revenue_growth_low", func(in *model.Input) { in.RevenueGrowthPct = -101 }, "revenue_growth_pct"},
		{"revenue_growth_high", func(in *model.Input) { in.RevenueGrowthPct = 501 }, "revenue_growth_pct"},
		{"expense_growth_low", func(in *model.Input) { in.ExpenseGrowthPct = -101 }, "expense_growth_pct"},
		{"expense_growth_high", func(in *model.Input) { in.ExpenseGrowthPct = 501 }, "expense_growth_pct"},
		{"purpose_too_long", func(in *model.Input) { in.Purpose = strings.Repeat("ы", 501) }, "purpose"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := validInput()
			tc.mut(&in)
			fields := Validate(in)
			if fields == nil {
				t.Fatal("ожидалась ошибка валидации")
			}
			if _, ok := fields[tc.field]; !ok {
				t.Errorf("нет поля %q в %v", tc.field, fields)
			}
		})
	}
}

func TestValidate_GrowthBoundsInclusive(t *testing.T) {
	in := validInput()
	in.RevenueGrowthPct = -100
	in.ExpenseGrowthPct = 500
	if fields := Validate(in); fields != nil {
		t.Errorf("границы должны быть включены: %v", fields)
	}
}

// --- Create ------------------------------------------------------------------

func TestCreate_InvalidInputDoesNotTouchStore(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)

	in := validInput()
	in.LoanAmount = 0
	_, err := svc.Create(context.Background(), nil, in)

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("err = %v, want ValidationError", err)
	}
	if len(store.created) != 0 {
		t.Errorf("хранилище не должно вызываться при ошибке валидации")
	}
}

func TestCreate_PersistsCalculatedResult(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)
	userID := int64(777)

	d, err := svc.Create(context.Background(), &userID, validInput())
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if len(store.created) != 1 {
		t.Fatalf("created = %d, want 1", len(store.created))
	}
	if store.created[0].MaxUserID == nil || *store.created[0].MaxUserID != 777 {
		t.Errorf("владелец не сохранён")
	}
	if store.created[0].Status != "calculated" || store.created[0].Type != "loan" {
		t.Errorf("type/status = %q/%q", store.created[0].Type, store.created[0].Status)
	}
	if len(d.Result.Scenarios) != 3 || len(d.Result.Checklist) != 5 {
		t.Errorf("сценарии=%d, чек-лист=%d", len(d.Result.Scenarios), len(d.Result.Checklist))
	}
	if d.Result.Credit.MonthlyPayment != 74_327 {
		t.Errorf("платёж = %v, want 74327", d.Result.Credit.MonthlyPayment)
	}
}

func TestCreate_StoreErrorWrapped(t *testing.T) {
	store := &fakeStore{createErr: errors.New("connection lost")}
	svc := New(store)

	_, err := svc.Create(context.Background(), nil, validInput())
	if err == nil || !strings.Contains(err.Error(), "save decision") {
		t.Fatalf("err = %v, want wrap 'save decision'", err)
	}
}

// --- Get / decorate ----------------------------------------------------------

func TestGet_DecoratesDerivedFields(t *testing.T) {
	months := 13
	store := &fakeStore{getDecision: model.Decision{
		ID:     "3b6f8a10-0000-4000-8000-000000000001",
		Type:   "loan",
		Inputs: validInput(),
		Result: model.DecisionResult{
			Scenarios: []model.Scenario{{
				Key: "negative", Revenue: 728_000, Expenses: 690_000,
				LoanPayment: 74_327, CashFlow: -36_327, DebtLoadPct: 10.2,
				ReserveAfter3M: 391_019, ReserveMonths: &months,
			}},
		},
	}}
	svc := New(store)

	d, err := svc.Get(context.Background(), "3b6f8a10-0000-4000-8000-000000000001")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if d.Result.Credit.MonthlyPayment != 74_327 {
		t.Errorf("credit не досчитан: %+v", d.Result.Credit)
	}
	s := d.Result.Scenarios[0]
	if s.Title == "" || s.Emoji == "" {
		t.Errorf("заголовок сценария не проставлен: %+v", s)
	}
	if len(s.Consequences) == 0 {
		t.Errorf("consequences не пересчитаны при чтении")
	}
	if len(d.Result.Checklist) != 5 {
		t.Errorf("checklist = %d", len(d.Result.Checklist))
	}
}

func TestGet_NotFoundPropagates(t *testing.T) {
	store := &fakeStore{getErr: repository.ErrNotFound}
	svc := New(store)

	if _, err := svc.Get(context.Background(), "x"); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

// --- Delete / Ping / List ----------------------------------------------------

func TestDelete_Ok(t *testing.T) {
	store := &fakeStore{deleteOK: true}
	svc := New(store)
	userID := int64(777)

	if err := svc.Delete(context.Background(), "some-id", &userID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if store.deleted != "some-id" {
		t.Errorf("deleted = %q", store.deleted)
	}
	if store.ownerSet == nil || *store.ownerSet != 777 {
		t.Errorf("владелец не передан в хранилище")
	}
}

func TestDelete_NotFoundBecomesErrNotFound(t *testing.T) {
	store := &fakeStore{deleteOK: false}
	svc := New(store)

	if err := svc.Delete(context.Background(), "some-id", nil); !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestDelete_StoreError(t *testing.T) {
	store := &fakeStore{deleteErr: errors.New("db down")}
	svc := New(store)

	err := svc.Delete(context.Background(), "some-id", nil)
	if err == nil || errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("err = %v, want wrapped db error", err)
	}
}

func TestPing(t *testing.T) {
	if err := New(&fakeStore{}).Ping(context.Background()); err != nil {
		t.Errorf("Ping: %v", err)
	}
	if err := New(&fakeStore{pingErr: errors.New("down")}).Ping(context.Background()); err == nil {
		t.Error("ожидалась ошибка")
	}
}

func TestList_PassesLimitAndUser(t *testing.T) {
	store := &fakeStore{}
	svc := New(store)
	userID := int64(42)

	if _, err := svc.List(context.Background(), &userID, 7); err != nil {
		t.Fatalf("List: %v", err)
	}
	if store.gotLimit != 7 {
		t.Errorf("limit = %d", store.gotLimit)
	}
	if store.gotUser == nil || *store.gotUser != 42 {
		t.Errorf("user = %v", store.gotUser)
	}
}
