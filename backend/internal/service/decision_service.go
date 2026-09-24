package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/KevinMarklin/decision-pause/backend/internal/calculator"
	"github.com/KevinMarklin/decision-pause/backend/internal/model"
	"github.com/KevinMarklin/decision-pause/backend/internal/repository"
)

// ValidationError — ошибки полей анкеты (отдаём клиенту как 400).
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string { return "validation failed" }

type DecisionService struct {
	repo *repository.DecisionRepo
}

func New(repo *repository.DecisionRepo) *DecisionService {
	return &DecisionService{repo: repo}
}

// Create — валидация → расчёт → сохранение.
func (s *DecisionService) Create(ctx context.Context, maxUserID *int64, in model.Input) (model.Decision, error) {
	if fields := Validate(in); len(fields) > 0 {
		return model.Decision{}, &ValidationError{Fields: fields}
	}
	d := model.Decision{
		MaxUserID: maxUserID,
		Type:      "loan",
		Status:    "calculated",
		CreatedAt: time.Now().UTC(),
		Inputs:    in,
		Result:    calculator.BuildResult(in),
	}
	if err := s.repo.Create(ctx, &d); err != nil {
		return model.Decision{}, fmt.Errorf("save decision: %w", err)
	}
	return d, nil
}

// Get — анализ по id, с заголовками сценариев и чек-листом.
func (s *DecisionService) Get(ctx context.Context, id string) (model.Decision, error) {
	d, err := s.repo.Get(ctx, id)
	if err != nil {
		return model.Decision{}, err
	}
	decorate(&d)
	return d, nil
}

func (s *DecisionService) List(ctx context.Context, maxUserID *int64, limit int) ([]model.DecisionListItem, error) {
	return s.repo.List(ctx, maxUserID, limit)
}

// decorate — добавляет производные поля, которые не храним в БД.
func decorate(d *model.Decision) {
	d.Result.Credit = calculator.CreditResult(d.Inputs)
	for i := range d.Result.Scenarios {
		d.Result.Scenarios[i].Title, d.Result.Scenarios[i].Emoji = model.ScenarioTitle(d.Result.Scenarios[i].Key)
	}
	d.Result.Checklist = calculator.Checklist()
}

// Validate — проверки полей анкеты, сообщения на русском для UI.
func Validate(in model.Input) map[string]string {
	fields := map[string]string{}

	if in.LoanAmount <= 0 {
		fields["loan_amount"] = "должна быть больше 0"
	}
	if in.LoanTermMonths < 1 || in.LoanTermMonths > 360 {
		fields["loan_term_months"] = "от 1 до 360 месяцев"
	}
	if in.InterestRatePct < 0 || in.InterestRatePct > 100 {
		fields["interest_rate_pct"] = "от 0 до 100%"
	}
	if in.Revenue <= 0 {
		fields["revenue"] = "должна быть больше 0"
	}
	if in.Expenses < 0 {
		fields["expenses"] = "не может быть отрицательной"
	}
	if in.Reserve < 0 {
		fields["reserve"] = "не может быть отрицательной"
	}
	if in.RevenueGrowthPct < -100 || in.RevenueGrowthPct > 500 {
		fields["revenue_growth_pct"] = "от -100 до 500%"
	}
	if in.ExpenseGrowthPct < -100 || in.ExpenseGrowthPct > 500 {
		fields["expense_growth_pct"] = "от -100 до 500%"
	}
	if len(strings.TrimSpace(in.Purpose)) > 500 {
		fields["purpose"] = "не длиннее 500 символов"
	}

	if len(fields) == 0 {
		return nil
	}
	return fields
}
