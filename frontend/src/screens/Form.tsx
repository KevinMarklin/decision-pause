import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import type { Inputs } from '../api/types'
import { useStore } from '../store'

// Значения полей — строки (controlled input); числа парсим при сабмите.
type FormValues = Record<keyof Inputs, string>

type FieldErrors = Partial<Record<keyof Inputs, string>>

const EMPTY: FormValues = {
  loan_amount: '',
  loan_term_months: '',
  interest_rate_pct: '',
  revenue: '',
  expenses: '',
  reserve: '',
  revenue_growth_pct: '',
  expense_growth_pct: '',
  purpose: '',
}

interface FieldDef {
  name: keyof Inputs
  label: string
  unit?: string
  placeholder?: string
  inputMode: 'numeric' | 'decimal' | 'text'
  integer?: boolean
}

// Порядок = порядок в анкете. Правила идентичны backend/internal/service.Validate
// (сообщения совпадают дословно — серверная ошибка не будет сюрпризом).
const FIELDS: FieldDef[] = [
  { name: 'loan_amount', label: 'Сумма кредита', unit: '₽', placeholder: '2000000', inputMode: 'numeric', integer: true },
  { name: 'loan_term_months', label: 'Срок кредита', unit: 'мес', placeholder: '36', inputMode: 'numeric', integer: true },
  { name: 'interest_rate_pct', label: 'Ставка', unit: '%', placeholder: '20', inputMode: 'decimal' },
  { name: 'revenue', label: 'Выручка в месяц', unit: '₽', placeholder: '800000', inputMode: 'numeric', integer: true },
  { name: 'expenses', label: 'Расходы в месяц', unit: '₽', placeholder: '600000', inputMode: 'numeric', integer: true },
  { name: 'reserve', label: 'Резерв', unit: '₽', placeholder: '500000', inputMode: 'numeric', integer: true },
  { name: 'revenue_growth_pct', label: 'Рост выручки за 3 мес.', unit: '%', placeholder: '30', inputMode: 'decimal' },
  { name: 'expense_growth_pct', label: 'Рост расходов за 3 мес.', unit: '%', placeholder: '15', inputMode: 'decimal' },
  { name: 'purpose', label: 'Цель кредита', placeholder: 'Закупка оборудования', inputMode: 'text' },
]

function parse(v: string): number {
  return Number(v.trim().replace(',', '.'))
}

// validate — те же правила, что у бэкенда (см. api.md).
function validate(values: FormValues): FieldErrors {
  const e: FieldErrors = {}
  const num = (name: keyof Inputs): number => parse(values[name])

  const amount = num('loan_amount')
  if (!Number.isFinite(amount) || amount <= 0) e.loan_amount = 'должна быть больше 0'

  const term = num('loan_term_months')
  if (!Number.isFinite(term) || !Number.isInteger(term) || term < 1 || term > 360)
    e.loan_term_months = 'от 1 до 360 месяцев'

  const rate = num('interest_rate_pct')
  if (!Number.isFinite(rate) || rate < 0 || rate > 100) e.interest_rate_pct = 'от 0 до 100%'

  const revenue = num('revenue')
  if (!Number.isFinite(revenue) || revenue <= 0) e.revenue = 'должна быть больше 0'

  const expenses = num('expenses')
  if (!Number.isFinite(expenses) || expenses < 0) e.expenses = 'не может быть отрицательной'

  const reserve = num('reserve')
  if (!Number.isFinite(reserve) || reserve < 0) e.reserve = 'не может быть отрицательной'

  const revGrowth = num('revenue_growth_pct')
  if (!Number.isFinite(revGrowth) || revGrowth < -100 || revGrowth > 500)
    e.revenue_growth_pct = 'от -100 до 500%'

  const expGrowth = num('expense_growth_pct')
  if (!Number.isFinite(expGrowth) || expGrowth < -100 || expGrowth > 500)
    e.expense_growth_pct = 'от -100 до 500%'

  if (values.purpose.length > 500) e.purpose = 'не длиннее 500 символов'
  return e
}

function toInputs(v: FormValues): Inputs {
  return {
    loan_amount: parse(v.loan_amount),
    loan_term_months: parse(v.loan_term_months),
    interest_rate_pct: parse(v.interest_rate_pct),
    revenue: parse(v.revenue),
    expenses: parse(v.expenses),
    reserve: parse(v.reserve),
    revenue_growth_pct: parse(v.revenue_growth_pct),
    expense_growth_pct: parse(v.expense_growth_pct),
    purpose: v.purpose.trim(),
  }
}

export default function Form() {
  const navigate = useNavigate()
  const { setDraft } = useStore()
  const [values, setValues] = useState<FormValues>(EMPTY)
  const [errors, setErrors] = useState<FieldErrors>({})

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    const errs = validate(values)
    setErrors(errs)
    if (Object.keys(errs).length > 0) return
    setDraft(toInputs(values))
    navigate('/review')
  }

  return (
    <main className="screen">
      <Link className="nav-link" to="/">
        ← Назад
      </Link>
      <h1>Взять кредит на развитие</h1>

      <form className="form" onSubmit={onSubmit} noValidate>
        {FIELDS.map((f) => (
          <label className="field" key={f.name}>
            <span className="field-label">
              {f.label}
              {f.unit ? <span className="muted">, {f.unit}</span> : null}
            </span>
            <input
              name={f.name}
              value={values[f.name]}
              placeholder={f.placeholder}
              inputMode={f.inputMode}
              autoComplete="off"
              onChange={(ev) => {
                const { value } = ev.target
                setValues((prev) => ({ ...prev, [f.name]: value }))
                setErrors((prev) => ({ ...prev, [f.name]: undefined }))
              }}
            />
            {errors[f.name] ? <span className="field-error">{errors[f.name]}</span> : null}
          </label>
        ))}

        <button className="btn btn-primary" type="submit">
          Проверить данные →
        </button>
      </form>
    </main>
  )
}
