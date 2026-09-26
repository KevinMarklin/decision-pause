import { useState, type FormEvent } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { postDecision } from '../api/client'
import { ApiError, type Inputs } from '../api/types'
import { FIELDS } from '../fields'
import { useStore } from '../store'

function formatValue(name: keyof Inputs, value: string | number): string {
  if (typeof value === 'number' && name !== 'purpose') {
    return value.toLocaleString('ru-RU')
  }
  return String(value)
}

export default function Review() {
  const navigate = useNavigate()
  const { draft, setDecision } = useStore()
  const [busy, setBusy] = useState(false)
  const [fieldErrs, setFieldErrs] = useState<Record<string, string>>({})
  const [formError, setFormError] = useState<string | null>(null)

  if (!draft) return <Navigate to="/" replace />

  async function calculate(inputs: Inputs) {
    setBusy(true)
    setFieldErrs({})
    setFormError(null)
    try {
      const decision = await postDecision(inputs)
      setDecision(decision)
      navigate('/result')
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        setFormError('Откройте приложение заново — сессия протухла.')
      } else if (err instanceof ApiError && err.code === 'validation' && err.fields) {
        setFieldErrs(err.fields)
      } else if (err instanceof ApiError && err.code === 'unknown_field' && err.unknownField) {
        setFieldErrs({ [err.unknownField]: 'неизвестное поле формы' })
      } else {
        setFormError('Что-то пошло не так, попробуйте ещё раз.')
      }
    } finally {
      setBusy(false)
    }
  }

  function onSubmit(e: FormEvent) {
    e.preventDefault()
    if (draft) void calculate(draft)
  }

  return (
    <main className="screen">
      <Link className="nav-link" to="/form">
        ← Изменить данные
      </Link>
      <h1>Проверка данных</h1>

      <div className="card review-list">
        {FIELDS.map((f) => (
          <div className="review-row" key={f.name}>
            <span className="muted">
              {f.label}
              {f.unit ? `, ${f.unit}` : ''}
            </span>
            <span className={fieldErrs[f.name] ? 'field-error' : undefined}>
              {fieldErrs[f.name] ?? formatValue(f.name, draft[f.name])}
            </span>
          </div>
        ))}
      </div>

      {formError ? <p className="field-error">{formError}</p> : null}

      <form className="form" onSubmit={onSubmit}>
        <button className="btn" type="button" onClick={() => navigate('/form')}>
          Изменить
        </button>
        <button className="btn btn-primary" type="submit" disabled={busy}>
          {busy ? 'Расчёт…' : 'Рассчитать'}
        </button>
      </form>
    </main>
  )
}
