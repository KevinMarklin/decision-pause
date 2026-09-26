import { useState, type FormEvent } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { postDecision } from '../api/client'
import { ApiError, type Inputs } from '../api/types'
import { FIELDS } from '../fields'
import { useStore } from '../useStore'

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
        setFormError('РћС‚РєСЂРѕР№С‚Рµ РїСЂРёР»РѕР¶РµРЅРёРµ Р·Р°РЅРѕРІРѕ вЂ” СЃРµСЃСЃРёСЏ РїСЂРѕС‚СѓС…Р»Р°.')
      } else if (err instanceof ApiError && err.code === 'validation' && err.fields) {
        setFieldErrs(err.fields)
      } else if (err instanceof ApiError && err.code === 'unknown_field' && err.unknownField) {
        setFieldErrs({ [err.unknownField]: 'РЅРµРёР·РІРµСЃС‚РЅРѕРµ РїРѕР»Рµ С„РѕСЂРјС‹' })
      } else {
        setFormError('Р§С‚Рѕ-С‚Рѕ РїРѕС€Р»Рѕ РЅРµ С‚Р°Рє, РїРѕРїСЂРѕР±СѓР№С‚Рµ РµС‰С‘ СЂР°Р·.')
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
        в†ђ РР·РјРµРЅРёС‚СЊ РґР°РЅРЅС‹Рµ
      </Link>
      <h1>РџСЂРѕРІРµСЂРєР° РґР°РЅРЅС‹С…</h1>

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
          РР·РјРµРЅРёС‚СЊ
        </button>
        <button className="btn btn-primary" type="submit" disabled={busy}>
          {busy ? 'Р Р°СЃС‡С‘С‚вЂ¦' : 'Р Р°СЃСЃС‡РёС‚Р°С‚СЊ'}
        </button>
      </form>
    </main>
  )
}
