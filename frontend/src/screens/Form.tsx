import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import type { Inputs } from '../api/types'
import {
  EMPTY_VALUES,
  FIELDS,
  parseNumber,
  type FieldErrors,
  type FormValues,
} from '../fields'
import { useStore } from '../useStore'

// validate вЂ” С‚Рµ Р¶Рµ РїСЂР°РІРёР»Р°, С‡С‚Рѕ Сѓ Р±СЌРєРµРЅРґР° (СЃРј. api.md, service.Validate).
function validate(values: FormValues): FieldErrors {
  const e: FieldErrors = {}
  const num = (name: keyof Inputs): number => parseNumber(values[name])

  const amount = num('loan_amount')
  if (!Number.isFinite(amount) || amount <= 0) e.loan_amount = 'РґРѕР»Р¶РЅР° Р±С‹С‚СЊ Р±РѕР»СЊС€Рµ 0'

  const term = num('loan_term_months')
  if (!Number.isFinite(term) || !Number.isInteger(term) || term < 1 || term > 360)
    e.loan_term_months = 'РѕС‚ 1 РґРѕ 360 РјРµСЃСЏС†РµРІ'

  const rate = num('interest_rate_pct')
  if (!Number.isFinite(rate) || rate < 0 || rate > 100) e.interest_rate_pct = 'РѕС‚ 0 РґРѕ 100%'

  const revenue = num('revenue')
  if (!Number.isFinite(revenue) || revenue <= 0) e.revenue = 'РґРѕР»Р¶РЅР° Р±С‹С‚СЊ Р±РѕР»СЊС€Рµ 0'

  const expenses = num('expenses')
  if (!Number.isFinite(expenses) || expenses < 0) e.expenses = 'РЅРµ РјРѕР¶РµС‚ Р±С‹С‚СЊ РѕС‚СЂРёС†Р°С‚РµР»СЊРЅРѕР№'

  const reserve = num('reserve')
  if (!Number.isFinite(reserve) || reserve < 0) e.reserve = 'РЅРµ РјРѕР¶РµС‚ Р±С‹С‚СЊ РѕС‚СЂРёС†Р°С‚РµР»СЊРЅРѕР№'

  const revGrowth = num('revenue_growth_pct')
  if (!Number.isFinite(revGrowth) || revGrowth < -100 || revGrowth > 500)
    e.revenue_growth_pct = 'РѕС‚ -100 РґРѕ 500%'

  const expGrowth = num('expense_growth_pct')
  if (!Number.isFinite(expGrowth) || expGrowth < -100 || expGrowth > 500)
    e.expense_growth_pct = 'РѕС‚ -100 РґРѕ 500%'

  if (values.purpose.length > 500) e.purpose = 'РЅРµ РґР»РёРЅРЅРµРµ 500 СЃРёРјРІРѕР»РѕРІ'
  return e
}

function toInputs(v: FormValues): Inputs {
  return {
    loan_amount: parseNumber(v.loan_amount),
    loan_term_months: parseNumber(v.loan_term_months),
    interest_rate_pct: parseNumber(v.interest_rate_pct),
    revenue: parseNumber(v.revenue),
    expenses: parseNumber(v.expenses),
    reserve: parseNumber(v.reserve),
    revenue_growth_pct: parseNumber(v.revenue_growth_pct),
    expense_growth_pct: parseNumber(v.expense_growth_pct),
    purpose: v.purpose.trim(),
  }
}

// В«РР·РјРµРЅРёС‚СЊВ» РЅР° СЌРєСЂР°РЅРµ РїСЂРѕРІРµСЂРєРё РґРѕР»Р¶РЅРѕ РІРѕР·РІСЂР°С‰Р°С‚СЊ Р·Р°РїРѕР»РЅРµРЅРЅСѓСЋ С„РѕСЂРјСѓ.
function fromInputs(ini: Inputs): FormValues {
  return {
    loan_amount: String(ini.loan_amount),
    loan_term_months: String(ini.loan_term_months),
    interest_rate_pct: String(ini.interest_rate_pct),
    revenue: String(ini.revenue),
    expenses: String(ini.expenses),
    reserve: String(ini.reserve),
    revenue_growth_pct: String(ini.revenue_growth_pct),
    expense_growth_pct: String(ini.expense_growth_pct),
    purpose: ini.purpose,
  }
}

export default function Form() {
  const navigate = useNavigate()
  const { draft, setDraft } = useStore()
  const [values, setValues] = useState<FormValues>(() =>
    draft ? fromInputs(draft) : EMPTY_VALUES,
  )
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
        в†ђ РќР°Р·Р°Рґ
      </Link>
      <h1>Р’Р·СЏС‚СЊ РєСЂРµРґРёС‚ РЅР° СЂР°Р·РІРёС‚РёРµ</h1>

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
          РџСЂРѕРІРµСЂРёС‚СЊ РґР°РЅРЅС‹Рµ в†’
        </button>
      </form>
    </main>
  )
}
