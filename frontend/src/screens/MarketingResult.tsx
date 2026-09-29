import { Link, Navigate } from 'react-router-dom'
import { money } from '../format'
import { useStore } from '../useStore'

export default function MarketingResult() {
  const { marketingResult } = useStore(); if (!marketingResult) return <Navigate to="/marketing" replace />
  return <main className="screen"><Link className="nav-link" to="/marketing">← Изменить данные</Link><h1>Последствия рекламной кампании</h1>
    <section className="card hero"><span className="muted">Точка безубыточности</span><strong>Не менее {marketingResult.break_even_orders} заказов</strong><span>Выручка {money(marketingResult.break_even_revenue)} ₽</span></section>
    <div className="marketing-scenarios">{marketingResult.scenarios.map((s) => <section className={`card scenario-${s.key}`} key={s.key}><h2>{s.emoji} {s.title}</h2><div className="rows"><div className="row"><span>Доп. заказов</span><strong>{s.additional_orders.toFixed(1)}</strong></div><div className="row"><span>Прибыль от рекламы</span><strong>{money(s.ad_profit)} ₽</strong></div><div className="row"><span>ROI</span><strong>{s.roi.toFixed(1)}%</strong></div><div className="row"><span>Cash Flow</span><strong className={s.cash_flow < 0 ? 'flow-neg' : 'flow-pos'}>{money(s.cash_flow)} ₽</strong></div>{s.reserve_months !== null ? <div className="row"><span>Резерва хватит</span><strong>{s.reserve_months} мес.</strong></div> : null}</div><p>{s.consequence}</p></section>)}</div>
    <section className="card"><h2>Последствия</h2><div className="consequences">{marketingResult.consequences.map((item) => <p key={item}>{item}</p>)}</div></section>
    <section className="card"><h2>Перед запуском</h2><div className="checklist">{marketingResult.checklist.map((item) => <label className="check-item" key={item}><input type="checkbox" />{item}</label>)}</div></section>
  </main>
}
