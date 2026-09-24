# Anti-Impulse API Contract

Base URL (dev): `http://localhost:8080` — в Vite используйте proxy:
`server: { proxy: { '/api': 'http://localhost:8080' } }`.

Auth нет. Для истории передавайте заголовок `X-Max-User-Id` (значение из
`window.WebApp.initDataUnsafe.user.id`) — опционально, без него история общая.

---

## POST /api/decisions

Создать анализ (валидация → расчёт → сохранение). Один запрос — весь результат.

Заголовки: `Content-Type: application/json`, опционально `X-Max-User-Id: 12345`

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

**Ошибки:**

- `400` валидация:
```json
{ "error": "validation", "fields": { "loan_amount": "должна быть больше 0" } }
```
Поля: `loan_amount` > 0 · `loan_term_months` 1–360 · `interest_rate_pct` 0–100 ·
`revenue` > 0 · `expenses` ≥ 0 · `reserve` ≥ 0 · `revenue_growth_pct`/`expense_growth_pct` −100…500 · `purpose` ≤ 500 симв.
- `400` `{"error":"invalid_json"}` · `400` `{"error":"unsupported decision type"}`
- `500` `{"error":"internal"}`

---

## GET /api/decisions/{id}

Тот же объект, что и POST-ответ (для перезагрузки/истории). `404` → `{"error":"not_found"}`.

## GET /api/decisions?limit=20

История (новые сверху, `limit` 1–100, default 20). Фильтр по `X-Max-User-Id`, если заголовок есть.

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

`{"status":"ok"}` — проверка, что backend жив.

---

## Заметки для фронта

- Валюты — рубли, целые числа; кириллица в JSON — UTF-8.
- `credit` и `checklist` можно показывать без расчёта на клиенте — всё считает backend.
- Коэффициенты сценариев: moderate = ожидаемая выручка × 0.9, negative × 0.7.
- MAX Bridge: `https://st.max.ru/js/max-web-app.js`, `window.WebApp.initDataUnsafe.user.id` → заголовок `X-Max-User-Id`.
