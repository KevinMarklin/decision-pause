# DecisionPause / «Анти-импульс»

Сервис, который помогает **не принимать импульсивных финансовых решений**.
Первый сценарий — «Взять кредит на развитие бизнеса»: анкета → расчёт
аннуитета → три сценария (ожидаемый / умеренно негативный / негативный) →
тексты последствий → чек-лист. Продукт **не говорит «делай/не делай»** —
он показывает, к чему приведёт решение.

Форма выпуска: бот в мессенджере **MAX** + Mini App. Рассчитывает **только
backend** (LLM в MVP не участвует).

## Стек

| Слой | Технология |
|---|---|
| Backend | Go 1.25, `net/http` (паттерны Go 1.22+), `pgx`, `golang-migrate` |
| БД | PostgreSQL 16 (Docker) |
| Бот | `max-bot-api-client-go`, long polling, вшитый корневой CA Минцифры |
| Frontend | Vite + React + TS (см. [`backend/docs/frontend-plan.md`](backend/docs/frontend-plan.md)) |
| API-контракт | [`backend/docs/api.md`](backend/docs/api.md) |

## Формулы (эталон для сверки цифр)

**Аннуитетный платёж** (`internal/calculator/loan.go`):

```
i  = rate_pct / 1200               (месячная ставка)
P  = S · i / (1 − (1+i)^−n)        ставка 0% → P = S / n
total = P · n,   overpay = total − S
```

**Три сценария** (`internal/calculator/scenarios.go`):

```
выручка_ож   = выручка   · (1 + рост_выручки% / 100)
расходы_ож   = расходы   · (1 + рост_расходов% / 100)
expected     = выручка_ож              (прогноз пользователя)
moderate     = выручка_ож · 0.9        (коэффициент ModerateFactor)
negative     = выручка_ож · 0.7        (коэффициент NegativeFactor)
```

Расходы и платёж **одинаковы во всех сценариях**: расходы уже приняты,
платёж фиксирован графиком. Дальше в каждом сценарии:

```
cash_flow        = выручка − расходы − платёж
долговая нагрузка = платёж / выручка · 100%
резерв через 3 мес = резерв + 3 · cash_flow
месяцы резерва    = floor(резерв / |cash_flow|)   при cash_flow < 0
```

Коэффициенты `0.9` / `0.7` — в `ModerateFactor` / `NegativeFactor`,
меняются только осознанно (тесты и README пересчитываются вместе с ними).

**Эталонные цифры** (сценарий из api.md): `2 000 000 ₽ / 36 мес / 20%`
→ платёж **74 327 ₽**, переплата **675 772 ₽**, cash_flow
🟢 **+275 673** / 🟡 **+171 673** / 🔴 **−36 327** (резерва хватит на 13 мес).

## Запуск

### Вариант 1 — всё в Docker (прод-подобный)

```bash
cd backend
docker compose up -d --build      # postgres + app (миграции, бот, API)
curl http://localhost:8080/health # {"status":"ok","db":"ok"}
```

### Вариант 2 — dev (Windows, Go-процесс отдельно)

```powershell
cd backend
docker compose up -d              # только postgres
go run ./cmd/server               # ищет backend/.env
.\scripts\smoke.ps1               # e2e-проверка API, ждём SMOKE OK
```

Сервер обязан запускаться из `backend/` (иначе не найдётся `.env`).

## Переменные окружения

| Переменная | Назначение |
|---|---|
| `PORT` | HTTP-порт (default `8080`) |
| `DATABASE_URL` | строка подключения Postgres |
| `MAX_BOT_TOKEN` | токен бота; пусто → бот выключен |
| `MAX_REQUIRE_INIT_DATA` | `true` → prod-auth: без валидной подписи initData → 401 |
| `CORS_ORIGINS` | allowlist origin'ов Mini App |
| `MAX_MINI_APP_NAME` | имя мини-аппа → кнопка «▶️ Начать анализ» (open_app) |
| `MAX_MINI_APP_URL` | deeplink → fallback-кнопка-ссылка |
| `FRONTEND_DIST` | каталог собранного Mini App, статика отдаётся отсюда |
| `TEST_DATABASE_URL` | включает интеграционные тесты репозитория |

Шаблон — [`backend/.env.example`](backend/.env.example). `.env` в git не попадает.

## Тесты

```bash
cd backend
gofmt -l . && go vet ./...        # чисто
go test ./...                     # юниты (~90)

docker compose up -d
$env:TEST_DATABASE_URL='postgres://anti:anti@localhost:5432/antipulse?sslmode=disable'
go test ./internal/repository/... -count=1   # интеграционные (gated)

.\scripts\smoke.ps1               # e2e против живого API (30 проверок)
```

## Структура

```
backend/
  cmd/server/          composition root
  internal/calculator/ чистые функции: аннуитет, сценария, последствия, чек-лист
  internal/service/    валидация → расчёт → сохранение
  internal/repository/ pgx, транзакции, миграции (embedded)
  internal/handler/    роуты, middleware (лог/recover/CORS/security), статика
  internal/auth/       HMAC-валидация initData MAX
  internal/bot/        long polling, приветствие, CA Минцифры
  internal/model/      только структуры
  migrations/          0001_init, 0002_drop_consequences
  docs/                api.md, frontend-plan.md
  scripts/smoke.ps1    e2e-проверка
frontend/              Mini App (Vite + React + TS) — см. docs/frontend-plan.md
```
