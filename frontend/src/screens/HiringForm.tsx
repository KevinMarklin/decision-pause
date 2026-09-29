import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { analyzeHiring } from '../api/client'
import type { HiringResult } from '../api/types'
import { money } from '../format'

const initialValues = { salary: '100 000', margin_percent: '30', average_check: '25 000', tax_rate: '30', overhead_costs: '0' }

function numberValue(value: string): number {
  return Number(value.replace(/\s/g, '').replace(',', '.'))
}

export default function HiringForm() {
  const [values, setValues] = useState(initialValues)
  const [result, setResult] = useState<HiringResult | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    const timer = window.setTimeout(async () => {
      const salary = numberValue(values.salary)
      const margin = numberValue(values.margin_percent)
      const averageCheck = numberValue(values.average_check)
      const taxRate = numberValue(values.tax_rate) / 100
      const overhead = numberValue(values.overhead_costs)
      if (!Number.isFinite(salary) || salary <= 0 || !Number.isFinite(margin) || margin <= 0 || margin > 100) {
        setResult(null)
        return
      }
      try {
        setError('')
        setResult(await analyzeHiring({ salary, margin_percent: margin, average_check: Number.isFinite(averageCheck) ? averageCheck : 0, tax_rate: Number.isFinite(taxRate) ? taxRate : 0.3, overhead_costs: Number.isFinite(overhead) ? overhead : 0 }))
      } catch {
        setError('Не удалось пересчитать данные. Проверьте значения.')
      }
    }, 250)
    return () => window.clearTimeout(timer)
  }, [values])

  function update(name: keyof typeof initialValues, value: string) {
    setValues((previous) => ({ ...previous, [name]: value }))
  }

  return <main className="screen">
    <Link className="nav-link" to="/">← Назад</Link>
    <h1>💼 Расчёт найма сотрудника</h1>
    <p className="disclaimer">Показываем полную стоимость сотрудника и выручку, которую он должен приносить для выхода в ноль.</p>
    <form className="form" onSubmit={(event) => event.preventDefault()}>
      <label className="field"><span className="field-label">Оклад сотрудника, ₽</span><input inputMode="decimal" value={values.salary} onChange={(event) => update('salary', event.target.value)} placeholder="100 000" /></label>
      <label className="field"><span className="field-label">Маржинальность бизнеса, %</span><input inputMode="decimal" value={values.margin_percent} onChange={(event) => update('margin_percent', event.target.value)} placeholder="30" /></label>
      <label className="field"><span className="field-label">Средний чек сделки, ₽ <span className="muted">(опционально)</span></span><input inputMode="decimal" value={values.average_check} onChange={(event) => update('average_check', event.target.value)} placeholder="25 000" /></label>
      <details className="advanced-settings"><summary>Дополнительные настройки</summary><div className="form"><label className="field"><span className="field-label">Страховые взносы, %</span><input inputMode="decimal" value={values.tax_rate} onChange={(event) => update('tax_rate', event.target.value)} placeholder="30" /></label><label className="field"><span className="field-label">Рабочее место и софт, ₽/мес</span><input inputMode="decimal" value={values.overhead_costs} onChange={(event) => update('overhead_costs', event.target.value)} placeholder="0" /></label></div></details>
    </form>
    {error ? <span className="field-error">{error}</span> : null}
    {result ? <section className="card hiring-result"><span className="muted">Результат расчёта</span><div className="hiring-metric hiring-warning"><span>Реальная стоимость сотрудника</span><strong>{money(result.total_cost)} ₽/мес</strong><small>с учётом страховых взносов {result.tax_rate * 100}% и накладных расходов</small></div><div className="hiring-metric"><span>Выручка для выхода в ноль</span><strong>{money(result.break_even_revenue)} ₽/мес</strong></div>{result.break_even_deals > 0 ? <div className="hiring-metric"><span>Нужно продаж</span><strong>{result.break_even_deals} сделок/мес</strong></div> : null}<p className="hiring-insight">Чтобы окупить оклад, сотрудник должен приносить выручку в <strong>{result.payback_multiplier.toFixed(1)} раза</strong> больше своего оклада.</p></section> : null}
  </main>
}
