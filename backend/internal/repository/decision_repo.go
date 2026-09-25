package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

type DecisionRepo struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *DecisionRepo {
	return &DecisionRepo{pool: pool}
}

// Create — одна транзакция: decision + inputs + 3 сценария.
// ID и CreatedAt заполняются из БД.
func (r *DecisionRepo) Create(ctx context.Context, d *model.Decision) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx) //nolint:errcheck // при успеху commit, откат не нужен

	err = tx.QueryRow(ctx, `
		INSERT INTO decisions (max_user_id, type, status)
		VALUES ($1, $2, 'calculated')
		RETURNING id::text, created_at`,
		d.MaxUserID, d.Type,
	).Scan(&d.ID, &d.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert decision: %w", err)
	}

	in := d.Inputs
	_, err = tx.Exec(ctx, `
		INSERT INTO decision_inputs (
			decision_id, loan_amount, loan_term, interest_rate,
			revenue, expenses, reserve, revenue_growth, expense_growth, purpose
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		d.ID, in.LoanAmount, in.LoanTermMonths, in.InterestRatePct,
		in.Revenue, in.Expenses, in.Reserve, in.RevenueGrowthPct, in.ExpenseGrowthPct, in.Purpose,
	)
	if err != nil {
		return fmt.Errorf("insert inputs: %w", err)
	}

	for _, s := range d.Result.Scenarios {
		_, err = tx.Exec(ctx, `
			INSERT INTO scenarios (
				decision_id, type, revenue, expenses, loan_payment, cash_flow,
				debt_load_pct, reserve_after_3m, reserve_months
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
			d.ID, s.Key, s.Revenue, s.Expenses, s.LoanPayment, s.CashFlow,
			s.DebtLoadPct, s.ReserveAfter3M, s.ReserveMonths,
		)
		if err != nil {
			return fmt.Errorf("insert scenario %s: %w", s.Key, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Get — полный анализ по id (заголовки сценариев и чек-лист добавляет service).
func (r *DecisionRepo) Get(ctx context.Context, id string) (model.Decision, error) {
	var d model.Decision
	err := r.pool.QueryRow(ctx, `
		SELECT d.id::text, d.max_user_id, d.type, d.status, d.created_at,
		       i.loan_amount, i.loan_term, i.interest_rate,
		       i.revenue, i.expenses, i.reserve,
		       i.revenue_growth, i.expense_growth, i.purpose
		FROM decisions d
		JOIN decision_inputs i ON i.decision_id = d.id
		WHERE d.id = $1`,
		id,
	).Scan(
		&d.ID, &d.MaxUserID, &d.Type, &d.Status, &d.CreatedAt,
		&d.Inputs.LoanAmount, &d.Inputs.LoanTermMonths, &d.Inputs.InterestRatePct,
		&d.Inputs.Revenue, &d.Inputs.Expenses, &d.Inputs.Reserve,
		&d.Inputs.RevenueGrowthPct, &d.Inputs.ExpenseGrowthPct, &d.Inputs.Purpose,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return model.Decision{}, ErrNotFound
		}
		return model.Decision{}, fmt.Errorf("select decision: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT type, revenue, expenses, loan_payment, cash_flow,
		       debt_load_pct, reserve_after_3m, reserve_months
		FROM scenarios
		WHERE decision_id = $1
		ORDER BY CASE type
			WHEN 'expected' THEN 1
			WHEN 'moderate' THEN 2
			WHEN 'negative' THEN 3
		END`,
		id,
	)
	if err != nil {
		return model.Decision{}, fmt.Errorf("select scenarios: %w", err)
	}
	defer rows.Close()

	d.Result.Scenarios = make([]model.Scenario, 0, 3)
	for rows.Next() {
		var s model.Scenario
		if err := rows.Scan(
			&s.Key, &s.Revenue, &s.Expenses, &s.LoanPayment, &s.CashFlow,
			&s.DebtLoadPct, &s.ReserveAfter3M, &s.ReserveMonths,
		); err != nil {
			return model.Decision{}, fmt.Errorf("scan scenario: %w", err)
		}
		d.Result.Scenarios = append(d.Result.Scenarios, s)
	}
	if err := rows.Err(); err != nil {
		return model.Decision{}, fmt.Errorf("iterate scenarios: %w", err)
	}
	return d, nil
}

// List — история: по всем пользователям или по одному (maxUserID = nil → все).
func (r *DecisionRepo) List(ctx context.Context, maxUserID *int64, limit int) ([]model.DecisionListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT d.id::text, d.created_at, d.type,
		       i.loan_amount, sc.loan_payment, sc.type, sc.cash_flow
		FROM decisions d
		JOIN decision_inputs i ON i.decision_id = d.id
		JOIN scenarios sc ON sc.decision_id = d.id
		WHERE d.id IN (
			SELECT id FROM decisions
			WHERE ($1::bigint IS NULL OR max_user_id = $1)
			ORDER BY created_at DESC
			LIMIT $2
		)
		ORDER BY d.created_at DESC`,
		maxUserID, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("select list: %w", err)
	}
	defer rows.Close()

	byID := make(map[string]*model.DecisionListItem)
	order := make([]string, 0, limit)

	for rows.Next() {
		var (
			id, dtype, scenarioType string
			createdAt               time.Time
			loanAmount, monthlyPay  float64
			cashFlow                float64
		)
		if err := rows.Scan(&id, &createdAt, &dtype, &loanAmount, &monthlyPay, &scenarioType, &cashFlow); err != nil {
			return nil, fmt.Errorf("scan list row: %w", err)
		}
		item, ok := byID[id]
		if !ok {
			item = &model.DecisionListItem{
				ID:             id,
				CreatedAt:      createdAt,
				Type:           dtype,
				LoanAmount:     loanAmount,
				MonthlyPayment: monthlyPay,
				CashFlows:      make(map[string]float64, 3),
			}
			byID[id] = item
			order = append(order, id)
		}
		item.CashFlows[scenarioType] = cashFlow
	}

	out := make([]model.DecisionListItem, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out, rows.Err()
}

// Ping — для /health: жив ли пул соединений с БД.
func (r *DecisionRepo) Ping(ctx context.Context) error {
	return r.pool.Ping(ctx)
}

// Delete — удаляет анализ. Возвращает ok=false, если записи нет либо она
// принадлежит другому пользователю (внешне это неотличимо от 404).
// maxUserID=nil (dev без заголовка) снимает проверку владельца.
func (r *DecisionRepo) Delete(ctx context.Context, id string, maxUserID *int64) (bool, error) {
	tag, err := r.pool.Exec(ctx, `
		DELETE FROM decisions
		WHERE id = $1
		  AND ($2::bigint IS NULL OR max_user_id IS NULL OR max_user_id = $2)`,
		id, maxUserID,
	)
	if err != nil {
		return false, fmt.Errorf("delete decision: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}
