package repository

import (
	"context"
	"errors"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/KevinMarklin/decision-pause/backend/internal/calculator"
	"github.com/KevinMarklin/decision-pause/backend/internal/model"
)

// Интеграционные тесты требуют живой Postgres:
//
//	docker compose up -d
//	TEST_DATABASE_URL=postgres://anti:anti@localhost:5432/antipulse?sslmode=disable \
//		go test ./internal/repository/...
//
// Без TEST_DATABASE_URL все тесты пропускаются.
func openTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL не задан — интеграционный тест пропущен")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	pool, err := Open(ctx, url)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := Migrate(url); err != nil {
		pool.Close()
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

// testUserSeq — гарантия уникальности внутри процесса.
// Разрешение time.Now() на Windows грубое: два подряд вызова могут вернуть
// одно и то же значение, и id двух «разных» пользователей совпадут.
var testUserSeq atomic.Int64

// newTestUserID — уникальный id, чтобы тесты не цеплялись за чужие записи.
func newTestUserID(t *testing.T) int64 {
	t.Helper()
	return time.Now().UnixNano()%1_000_000_000_000 + testUserSeq.Add(1)
}

func makeDecision(userID *int64) *model.Decision {
	in := model.Input{
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
	return &model.Decision{
		MaxUserID: userID,
		Type:      "loan",
		Status:    "calculated",
		CreatedAt: time.Now().UTC(),
		Inputs:    in,
		Result:    calculator.BuildResult(in),
	}
}

func cleanupUser(t *testing.T, pool *pgxpool.Pool, userID int64) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, "DELETE FROM decisions WHERE max_user_id = $1", userID)
	})
}

func TestIntegration_Ping(t *testing.T) {
	pool := openTestDB(t)
	if err := New(pool).Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestIntegration_CreateAndGet(t *testing.T) {
	pool := openTestDB(t)
	repo := New(pool)
	ctx := context.Background()
	user := newTestUserID(t)
	cleanupUser(t, pool, user)

	d := makeDecision(&user)
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if d.ID == "" {
		t.Fatal("id не заполнен из БД")
	}

	got, err := repo.Get(ctx, d.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.MaxUserID == nil || *got.MaxUserID != user {
		t.Errorf("max_user_id = %v, want %d", got.MaxUserID, user)
	}
	if got.Inputs.LoanAmount != 2_000_000 || got.Inputs.Purpose != "Закупка оборудования" {
		t.Errorf("inputs = %+v", got.Inputs)
	}
	if len(got.Result.Scenarios) != 3 {
		t.Fatalf("сценариев = %d, want 3", len(got.Result.Scenarios))
	}

	wantKeys := []string{"expected", "moderate", "negative"}
	for i, key := range wantKeys {
		if got.Result.Scenarios[i].Key != key {
			t.Errorf("сценарий[%d] = %q, want %q", i, got.Result.Scenarios[i].Key, key)
		}
	}
	if cf := got.Result.Scenarios[2].CashFlow; cf != -36_327 {
		t.Errorf("negative cash_flow = %v, want -36327", cf)
	}
	if got.Result.Scenarios[2].ReserveMonths == nil || *got.Result.Scenarios[2].ReserveMonths != 13 {
		t.Errorf("negative reserve_months = %v, want 13", got.Result.Scenarios[2].ReserveMonths)
	}
	// производные поля в БД не хранятся — их добавляет service.decorate
	if got.Result.Checklist != nil || got.Result.Credit.MonthlyPayment != 0 {
		t.Errorf("credit/checklist не должны приходить из репозитория")
	}
}

func TestIntegration_GetNotFound(t *testing.T) {
	pool := openTestDB(t)
	_, err := New(pool).Get(context.Background(), "00000000-0000-4000-8000-000000000000")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestIntegration_List(t *testing.T) {
	pool := openTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	userA := newTestUserID(t)
	userB := newTestUserID(t)
	cleanupUser(t, pool, userA)
	cleanupUser(t, pool, userB)

	first := makeDecision(&userA)
	if err := repo.Create(ctx, first); err != nil {
		t.Fatalf("create first: %v", err)
	}
	second := makeDecision(&userA)
	if err := repo.Create(ctx, second); err != nil {
		t.Fatalf("create second: %v", err)
	}
	other := makeDecision(&userB)
	if err := repo.Create(ctx, other); err != nil {
		t.Fatalf("create other: %v", err)
	}

	onlyA, err := repo.List(ctx, &userA, 20)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	ids := map[string]bool{}
	for _, it := range onlyA {
		ids[it.ID] = true
		if it.CashFlows["negative"] != -36_327 {
			t.Errorf("cash_flows = %v", it.CashFlows)
		}
		if it.MonthlyPayment != 74_327 || it.LoanAmount != 2_000_000 {
			t.Errorf("сводка = %+v", it)
		}
	}
	if !ids[first.ID] || !ids[second.ID] {
		t.Errorf("в истории пользователя A нет его записей: %v", ids)
	}
	if ids[other.ID] {
		t.Errorf("запись пользователя B попала в историю A")
	}
	if onlyA[0].ID != second.ID {
		t.Errorf("новые сверху: первый = %s, want %s", onlyA[0].ID, second.ID)
	}

	all, err := repo.List(ctx, nil, 100)
	if err != nil {
		t.Fatalf("List(nil): %v", err)
	}
	if len(all) < 3 {
		t.Errorf("общая история = %d записей, want >= 3", len(all))
	}

	limited, err := repo.List(ctx, &userA, 1)
	if err != nil {
		t.Fatalf("List(limit=1): %v", err)
	}
	if len(limited) != 1 {
		t.Errorf("limit=1 → %d записей", len(limited))
	}
}

func TestIntegration_Delete(t *testing.T) {
	pool := openTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	owner := newTestUserID(t)
	stranger := newTestUserID(t)
	cleanupUser(t, pool, owner)
	cleanupUser(t, pool, stranger)

	d := makeDecision(&owner)
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// чужой пользователь удалить не может
	ok, err := repo.Delete(ctx, d.ID, &stranger)
	if err != nil {
		t.Fatalf("Delete(stranger): %v", err)
	}
	if ok {
		t.Error("чужой пользователь удалил запись")
	}
	if _, err := repo.Get(ctx, d.ID); err != nil {
		t.Errorf("запись должна остаться после неудачной попытки: %v", err)
	}

	// владелец может
	ok, err = repo.Delete(ctx, d.ID, &owner)
	if err != nil {
		t.Fatalf("Delete(owner): %v", err)
	}
	if !ok {
		t.Error("владелец не удалил запись")
	}
	if _, err := repo.Get(ctx, d.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("после удаления Get = %v, want ErrNotFound", err)
	}

	// несуществующая запись → ok=false
	ok, err = repo.Delete(ctx, "00000000-0000-4000-8000-000000000000", nil)
	if err != nil || ok {
		t.Errorf("Delete(unknown) = %v/%v, want false/nil", ok, err)
	}
}

func TestIntegration_DeleteCascades(t *testing.T) {
	pool := openTestDB(t)
	repo := New(pool)
	ctx := context.Background()

	user := newTestUserID(t)
	cleanupUser(t, pool, user)

	d := makeDecision(&user)
	if err := repo.Create(ctx, d); err != nil {
		t.Fatalf("Create: %v", err)
	}

	var inputs, scenarios int
	if err := pool.QueryRow(ctx,
		"SELECT (SELECT count(*) FROM decision_inputs WHERE decision_id = $1),"+
			"       (SELECT count(*) FROM scenarios WHERE decision_id = $1)", d.ID,
	).Scan(&inputs, &scenarios); err != nil {
		t.Fatalf("count: %v", err)
	}
	if inputs != 1 || scenarios != 3 {
		t.Fatalf("до удаления: inputs=%d scenarios=%d, want 1/3", inputs, scenarios)
	}

	if _, err := repo.Delete(ctx, d.ID, &user); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if err := pool.QueryRow(ctx,
		"SELECT (SELECT count(*) FROM decision_inputs WHERE decision_id = $1),"+
			"       (SELECT count(*) FROM scenarios WHERE decision_id = $1)", d.ID,
	).Scan(&inputs, &scenarios); err != nil {
		t.Fatalf("count after delete: %v", err)
	}
	if inputs != 0 || scenarios != 0 {
		t.Errorf("каскад не сработал: inputs=%d scenarios=%d", inputs, scenarios)
	}
}
