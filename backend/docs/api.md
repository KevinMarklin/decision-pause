# Anti-Impulse API Contract

Base URL (dev): `http://localhost:8080` — в Vite используйте proxy:
`server: { proxy: { '/api': 'http://localhost:8080' } }`.

---

## Аутентификация

Сессий нет, но идентичность пользователя защищена подписью MAX.

### Прод: `MAX_REQUIRE_INIT_DATA=true`

- Клиент шлёт **сырую строку** `window.WebApp.initData` в заголовке **`X-Max-Init-Data`**.
- Сервер проверяет HMAC-SHA256-подпись токеном бота (алгоритм MAX) и берёт
  `user.id` уже **из подписанного payload** — клиентскому значению не верим.
- Нет заголовка или подпись не сошлась → `401 {"error":"unauthorized"}`.
- Заголовок `X-Max-User-Id` в этом режиме **игнорируется**.
- Требуется `MAX_BOT_TOKEN`: без него сервер не стартует.

### Dev (по умолчанию)

- `X-Max-Init-Data` не обязателен; **валидная подпись имеет приоритет**.
- Иначе берётся legacy `X-Max-User-Id` (значение из
  `window.WebApp.initDataUnsafe.user.id`) → история пользователя.
- Заголовков нет → общая история по всем пользователям.
- `X-Max-User-Id` есть, но не число / ≤ 0 → `400 {"error":"invalid_user_id"}`
  (раньше такой запрос молча отдавал чужую историю).

Свежесть подписи: `auth_date` ≤ 1 часа, «будущее» ≤ 5 минут.

---

## POST /api/decisions

Создать анализ (валидация → расчёт → сохранение). Один запрос — весь результат.

Заголовки: `Content-Type: application/json`, опционально
`X-Max-Init-Data` (прод) или `X-Max-User-Id` (dev).

```json
{
  "type": "loan",
  "inputs": {
    "loan_amount": 2000000,
    "loan_term_months": 36,
    "interest_rate_pct": 20,
    "revenue": 800000,
    "expenses": 600000,
    "reserve": 500000,
    "revenue_growth_pct": 30,
    "expense_growth_pct": 15,
    "purpose": "Закупка оборудования"
  }
}
```

`type` — только `"loan"` (можно не указывать). `purpose` — строка до 500 символов.

**JSON строгий:**

- только `application/json` (иначе `415 unsupported_media_type`);
- тело ≤ 64 КиБ (иначе `413 payload_too_large`);
- неизвестные поля запрещены — иначе `400 unknown_field` с именем поля.
  Это ловит опечатки: `interest_rate` вместо `interest_rate_pct` молча дал бы 0%;
- данные после первого значения JSON запрещены → `400 invalid_json`.

**Ответ `201`:**

```json
{
  "id": "uuid",
  "created_at": "2026-09-25T01:09:21Z",
  "type": "loan",
  "inputs": { "...": "эхо всех входных полей" },
  "credit": {
    "monthly_payment": 74327,
    "total_paid": 2675772,
    "overpay": 675772
  },
  "scenarios": [
    {
      "key": "expected",
      "title": "Ожидаемый",
      "emoji": "🟢",
      "revenue": 1040000,
      "expenses": 690000,
      "loan_payment": 74327,
      "cash_flow": 275673,
      "debt_load_pct": 7.1,
      "reserve_after_3m": 1327019,
      "reserve_months": null,
      "consequences": [
        "Платёж покрывается из текущей выручки, свободный поток +275 673 ₽/мес."
      ]
    },
    { "key": "moderate", "title": "Умеренно негативный", "emoji": "🟡", "...": "выручка × 0.9" },
    { "key": "negative", "title": "Негативный", "emoji": "🔴", "...": "выручка × 0.7" }
  ],
  "checklist": [
    "Хватит ли резерва?",
    "Что будет при падении выручки?",
    "Учтены ли дополнительные расходы?",
    "Есть ли план погашения?",
    "Реалистичен ли прогноз роста?"
  ]
}
```

Поля `scenarios[]` (порядок всегда: expected → moderate → negative):

| Поле | Тип | Смысл |
|---|---|---|
| `cash_flow` | number | свободный денежный поток, ₽/мес (может быть < 0) |
| `debt_load_pct` | number | платёж / выручка, % |
| `reserve_after_3m` | number | резерв через 3 мес. при таком потоке, ₽ |
| `reserve_months` | number\|null | сколько мес. хватит резерва при дефиците, иначе null |
| `consequences` | string[] | тексты последствий (готово для блока «Последствия») |

`credit`, `title`/`emoji` сценариев, `consequences` и `checklist` **пересчитываются
при каждом чтении** — это производные поля, в БД не хранятся.

**Ошибки:**

- `400` валидация:
```json
{ "error": "validation", "fields": { "loan_amount": "должна быть больше 0" } }
```
Поля: `loan_amount` > 0 · `loan_term_months` 1–360 · `interest_rate_pct` 0–100 ·
`revenue` > 0 · `expenses` ≥ 0 · `reserve` ≥ 0 · `revenue_growth_pct`/`expense_growth_pct` −100…500 · `purpose` ≤ 500 симв.
- `400` `{"error":"invalid_json"}` · `400` `{"error":"unknown_field","field":"имя"}`
- `400` `{"error":"unsupported decision type"}`
- `401` `{"error":"unauthorized"}` — только при `MAX_REQUIRE_INIT_DATA=true`
- `413` `{"error":"payload_too_large"}` · `415` `{"error":"unsupported_media_type"}`
- `500` `{"error":"internal"}` (включая панику — сервер не падает)

---

## GET /api/decisions/{id}

Тот же объект, что и POST-ответ (для перезагрузки/истории).

- `404` → `{"error":"not_found"}`
- `400` `{"error":"invalid_id"}` — `{id}` не является UUID

## DELETE /api/decisions/{id}

Удалить снимок расчёта. Изменить нельзя — решение это снимок на момент расчёта
(пересчёт = создать заново).

- `204` — удалено, тела нет
- `404` → `{"error":"not_found"}` — записи нет **или** она принадлежит другому
  пользователю (сознательно неразличимо)
- `400` `{"error":"invalid_id"}`

Право на удаление: в проде — только владелец из подписи; в dev без заголовка
пользователя (`max_user_id IS NULL`) запись удалит любой запрос.

## GET /api/decisions?limit=20

История (новые сверху, `limit` 1–100, default 20). Фильтр по пользователю, если
есть `X-Max-Init-Data`/`X-Max-User-Id`, иначе общая история.

- `400` `{"error":"invalid_limit"}` — `limit` не число или вне 1–100

```json
{ "items": [
  {
    "id": "uuid",
    "created_at": "...",
    "type": "loan",
    "loan_amount": 2000000,
    "monthly_payment": 74327,
    "cash_flows": { "expected": 275673, "moderate": 171673, "negative": -36327 }
  }
] }
```

## GET /health

Живость API **и** доступность БД. Эндпоинт **публичный**: под auth-middleware
не попадает, поэтому работает и в прод-режиме — им пользуются балансировщик
и мониторинг, которые initData не шлют.

- `{"status":"ok","db":"ok"}` — всё в порядке
- `503` `{"status":"degraded","db":"down"}` — БД недоступна

---

## Сводная таблица ошибок

| Статус | `error` | Когда |
|---|---|---|
| 400 | `validation` | не прошли правила полей (тело — `fields`) |
| 400 | `invalid_json` | синтаксис, хвост после значения |
| 400 | `unknown_field` | лишнее поле (тело — `field`) |
| 400 | `unsupported decision type` | `type` ≠ `loan` |
| 400 | `invalid_id` | `{id}` не UUID |
| 400 | `invalid_limit` | `limit` вне 1–100 |
| 400 | `invalid_user_id` | битый legacy-заголовок |
| 401 | `unauthorized` | нет/не прошла подпись initData в проде |
| 404 | `not_found` | записи нет или чужая |
| 413 | `payload_too_large` | тело > 64 КиБ |
| 415 | `unsupported_media_type` | не JSON |
| 500 | `internal` | ошибка сервера/паника |
| 503 | `status=degraded` | БД недоступна (`/health`) |

---

## Заметки для фронта

- Валюты — рубли, целые числа; кириллица в JSON — UTF-8.
- `credit` и `checklist` можно показывать без расчёта на клиенте — всё считает backend.
- Коэффициенты сценариев: moderate = ожидаемая выручка × 0.9, negative × 0.7.
- MAX Bridge: `https://st.max.ru/js/max-web-app.js`.
  Прод: слать **сырую** `window.WebApp.initData` в `X-Max-Init-Data`.
  Dev: `window.WebApp.initDataUnsafe.user.id` в `X-Max-User-Id`.
- Ловите `unknown_field` — это опечатка в имени поля формы, а не ошибка сервера.
