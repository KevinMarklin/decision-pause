import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { analyzeMarketing } from '../api/client'
import type { MarketingInputs } from '../api/types'
import { useStore } from '../useStore'

const fields: { name: keyof MarketingInputs; label: string; placeholder: string }[] = [
  { name: 'budget', label: 'Бюджет на рекламу, ₽', placeholder: '300 000' }, { name: 'avg_check', label: 'Средний чек, ₽', placeholder: '10 000' },
  { name: 'margin_pct', label: 'Маржинальность, %', placeholder: '40' }, { name: 'lead_cost', label: 'Стоимость лида / клиента (CAC), ₽', placeholder: '1 500' },
  { name: 'expected_boost_pct', label: 'Ожидаемый прирост продаж, %', placeholder: '20' }, { name: 'current_revenue', label: 'Выручка в месяц, ₽', placeholder: '800 000' },
  { name: 'fixed_expenses', label: 'Расходы в месяц, ₽', placeholder: '600 000' }, { name: 'reserve', label: 'Резерв, ₽', placeholder: '500 000' },
]

export default function MarketingForm() {
  const navigate = useNavigate(); const { marketingDraft, setMarketingDraft, setMarketingResult } = useStore()
  const [values, setValues] = useState<Record<keyof MarketingInputs, string>>(() => Object.fromEntries(fields.map((f) => [f.name, marketingDraft?.[f.name] ? String(marketingDraft[f.name]) : ''])) as Record<keyof MarketingInputs, string>)
  const [error, setError] = useState('')
  const [blackSwan, setBlackSwan] = useState(Boolean(marketingDraft?.is_black_swan))
  async function onSubmit(event: FormEvent) {
    event.preventDefault()
    const numberValue = (name: keyof MarketingInputs): number =>
      Number(values[name].replace(/\s/g, '').replace(',', '.'))
    const input: MarketingInputs = {
      budget: numberValue('budget'),
      avg_check: numberValue('avg_check'),
      margin_pct: numberValue('margin_pct'),
      lead_cost: numberValue('lead_cost'),
      expected_boost_pct: numberValue('expected_boost_pct'),
      current_revenue: numberValue('current_revenue'),
      fixed_expenses: numberValue('fixed_expenses'),
      reserve: numberValue('reserve'),
    }
    if (Object.values(input).some((value) => !Number.isFinite(value))) { setError('Заполните все поля числами'); return }
    try {
      const request = { ...input, is_black_swan: blackSwan }
      setMarketingDraft(request)
      setMarketingResult(await analyzeMarketing(request))
      navigate('/marketing/result')
    } catch { setError('Не удалось проверить данные. Попробуйте ещё раз.') }
  }
  return <main className="screen"><Link className="nav-link" to="/">← Назад</Link><h1>Запустить рекламную кампанию</h1><form className="form" onSubmit={onSubmit} noValidate>
    <label className="panic-toggle"><input type="checkbox" checked={blackSwan} onChange={(e) => setBlackSwan(e.target.checked)} /><span><strong>🧪 Стресс-тест «Чёрный лебедь»</strong><small>Падение выручки на 20% и рост расходов на 15%</small></span></label>
    {fields.map((f) => <label className="field" key={f.name}><span className="field-label">{f.label}</span><input inputMode="decimal" name={f.name} placeholder={f.placeholder} value={values[f.name]} onChange={(e) => setValues((prev) => ({ ...prev, [f.name]: e.target.value }))} /></label>)}
    {error ? <span className="field-error">{error}</span> : null}<button className="btn btn-primary" type="submit">Проверить данные →</button>
  </form></main>
}
