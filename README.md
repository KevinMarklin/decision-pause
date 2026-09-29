# DecisionPause / «Анти-импульс»

Сервис, который помогает **не принимать импульсивных финансовых решений**.
Первый сценарий — «Взять кредит на развитие бизнеса»: анкета → расчёт аннуитета → три сценария (ожидаемый / умеренно негативный / негативный) → тексты последствий → чек-лист. Продукт **не говорит «делай/не делай»** — он показывает, к чему приведёт решение.

Форма выпуска: бот в мессенджере **MAX** + Mini App. Рассчитывает **только backend** (LLM в MVP не участвует).

## Стек

| Слой | Технология |
|---|---|
| Backend | Go 1.25, `net/http` (паттерны Go 1.22+), `pgx`, `golang-migrate` |
| БД | PostgreSQL 16 (Docker) |
| Бот | `max-bot-api-client-go`, long polling, вшитый корневой CA Минцифры |
| Frontend | Vite + React + TS (см. [`docs/frontend-plan.md`](docs/frontend-plan.md)) |
| API-контракт | [`docs/api.md`](docs/api.md) |

---

## Файлы Docker и конфигурации

Проект содержит полный набор конфигурационных файлов и манифестов, необходимых для локальной разработки, сборки контейнеров и развёртывания в production-окружении:

### 1. Docker-файлы
* **`docker-compose.yml`** — главный манифест оркестрации сервисов приложения:
  * `postgres`: СУБД PostgreSQL 16 Alpine с автоматической проверкой состояния (`healthcheck` через `pg_isready`) и персистентным томом `pgdata`.
  * `app`: мультистадийный контейнер приложения, объединяющий Go-бэкенд, раздачу собранной статики Mini App и long-polling бота.
* **`backend/Dockerfile`** — сценарий multi-stage сборки образа:
  * **Stage 1 (Frontend):** сборка клиентского приложения Vite + React на базе Node.js (`npm run build`).
  * **Stage 2 (Backend Build):** компиляция бинарного файла Go (`CGO_ENABLED=0 go build`) со сжатием и оптимизацией флагов `-ldflags="-s -w"`.
  * **Stage 3 (Runner):** финальный минимальный образ на базе Alpine Linux, содержащий только скомпилированный Go-сервер и статические файлы Mini App.
* **`.dockerignore`** — оптимизирует контекст сборки Docker, предотвращая попадание временно созданных файлов, локальных версий, бинарников и секретов (`.git`, `node_modules`, `dist`, `.env`, `*.exe`, `pgdata`).

### 2. Конфигурационные файлы
* **`backend/.env.example`** — эталонный шаблон переменных окружения без секретов и токенов. На его основе создаётся локальный рабочий `.env` файл.

### 3. Управление зависимостями
* **Backend (Go):**
  * `backend/go.mod` — манифест модулей Go и версий зависимостей.
  * `backend/go.sum` — контрольные суммы (хеши) зависимостей для гарантированной воспроизводимости сборки.
* **Frontend (Node.js):**
  * `frontend/package.json` — описание скриптов сборки, внешних библиотек (React, Vite, TypeScript, Tailwind) и dev-зависимостей.
  * `frontend/package-lock.json` — зафиксированное дерево зависимостей Node.js.

---

## Формулы (эталон для сверки цифр)

**Аннуитетный платёж** (`internal/calculator/loan.go`):

```text
i   = rate_pct / 1200                (месячная ставка)
P   = S · i / (1 − (1+i)^−n)         ставка 0% → P = S / n
total = P · n,   overpay = total − S

выручка_ож   = выручка   · (1 + рост_выручки% / 100)
расходы_ож   = расходы   · (1 + рост_расходов% / 100)
expected     = выручка_ож               (прогноз пользователя)
moderate     = выручка_ож · 0.9         (коэффициент ModerateFactor)
negative     = выручка_ож · 0.7         (коэффициент NegativeFactor)


Три сценария (internal/calculator/scenarios.go):


выручка_ож   = выручка   · (1 + рост_выручки% / 100)
расходы_ож   = расходы   · (1 + рост_расходов% / 100)
expected     = выручка_ож               (прогноз пользователя)
moderate     = выручка_ож · 0.9         (коэффициент ModerateFactor)
negative     = выручка_ож · 0.7         (коэффициент NegativeFactor)


Расходы и платёж одинаковы во всех сценариях: расходы уже приняты, платёж фиксирован графиком. Дальше в каждом сценарии:

cash_flow         = выручка − расходы − платёж
долговая нагрузка = платёж / выручка · 100%
резерв через 3 мес = резерв + 3 · cash_flow
месяцы резерва    = floor(резерв / |cash_flow|)   при cash_flow < 0

Коэффициенты 0.9 / 0.7 — в ModerateFactor / NegativeFactor, меняются только осознанно (тесты и README пересчитываются вместе с ними).

Эталонные цифры (сценарий из docs/api.md): 2 000 000 ₽ / 36 мес / 20% → платёж 74 327 ₽, переплата 675 772 ₽, cash_flow 🟢 +275 673 / 🟡 +171 673 / 🔴 −36 327 (резерва хватит на 13 мес).

## Переменные окружения

| Переменная | Назначение |
|---|---|
| `PORT` | HTTP-порт сервера (default `8080`) |
| `DATABASE_URL` | Строка подключения Postgres |
| `MAX_BOT_TOKEN` | Токен бота в MAX; пусто → бот выключен |
| `MAX_REQUIRE_INIT_DATA` | `true` → prod-auth: без валидной подписи initData → 401 |
| `CORS_ORIGINS` | Allowlist origin'ов Mini App |
| `MAX_MINI_APP_NAME` | Имя мини-аппа → кнопка «▶️ Начать анализ» (open_app) |
| `MAX_MINI_APP_URL` | Deeplink → fallback-кнопка-ссылка |
| `FRONTEND_DIST` | Каталог собранного Mini App, статика отдаётся отсюда |
| `TEST_DATABASE_URL` | Включает интеграционные тесты репозитория |

Вариант 1 — всё в Docker (прод-подобный)
docker compose up -d --build      # postgres + app (миграции, бот, API, фронтенд)
curl http://localhost:8080/health # {"status":"ok","db":"ok"}

Вариант 2 — dev (Windows, Go-процесс отдельно)
cd backend
docker compose up -d              # только postgres
go run ./cmd/server               # ищет backend/.env
.\scripts\smoke.ps1               # e2e-проверка API, ждём SMOKE OK

Примечание: Сервер обязан запускаться из каталога backend/ (иначе не найдётся .env).


Тесты
cd backend
gofmt -l . && go vet ./...        # проверка форматирования и статический анализ
go test ./...                     # юниты (~90)

docker compose up -d
$env:TEST_DATABASE_URL='postgres://anti:anti@localhost:5432/antipulse?sslmode=disable'
go test ./internal/repository/... -count=1   # интеграционные тесты БД

.\scripts\smoke.ps1               # e2e сквозная проверка против живого API (30 проверок)


Структура проекта
.
├── docker-compose.yml       # оркестрация PostgreSQL и сервиса приложения
├── .dockerignore            # исключения контекста сборки Docker
├── backend/
│   ├── Dockerfile           # мультистадийный Dockerfile (React + Go)
│   ├── .env.example         # пример конфигурации окружения
│   ├── go.mod               # зависимости Go
│   ├── go.sum               # контрольные суммы зависимостей Go
│   ├── cmd/server/          # composition root (точка входа)
│   ├── internal/
│   │   ├── calculator/      # чистые функции: аннуитет, сценарии, последствия, чек-лист
│   │   ├── service/         # валидация → расчёт → сохранение
│   │   ├── repository/      # pgx, транзакции, встроенные миграции
│   │   ├── handler/         # роуты, middleware (лог/recover/CORS/security), статика
│   │   ├── auth/            # HMAC-валидация initData MAX
│   │   ├── bot/             # long polling, приветствие, CA Минцифры
│   │   └── model/           # структуры данных
│   ├── migrations/          # SQL миграции СУБД (0001_init, 0002_drop_consequences)
│   ├── docs/                # документация API и планов фронтенда
│   └── scripts/smoke.ps1    # e2e smoke-тесты
└── frontend/                # Mini App (Vite + React + TS)
    ├── package.json         # зависимости Node.js и скрипты
    └── package-lock.json    # лок-файл зависимостей Node.js
