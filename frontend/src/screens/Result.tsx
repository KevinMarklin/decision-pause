import { Link, Navigate, useNavigate } from 'react-router-dom'
import { money, pct, signedMoney } from '../format'
import { useStore } from '../store'

export default function Result() {
  const navigate = useNavigate()
  const { decision, setScenarioKey } = useStore()

  if (!decision) return <Navigate to="/" replace />
  const { credit } = decision

  return (
    <main className="screen">
      <Link className="nav-link" to="/">
        ← На главную
      </Link>
      <h1>Результат</h1>

      <section className="card">
        <h2>💳 Кредит</h2>
        <div className="rows">
          <div className="row">
            <span className="muted">Платёж в месяц</span>
            <strong>{money(credit.monthly_payment)} ₽</strong>
          </div>
          <div className="row">
            <span className="muted">Всего выплат</span>
            <strong>{money(credit.total_paid)} ₽</strong>
          </div>
          <div className="row">
            <span className="muted">Переплата</span>
            <strong>{money(credit.overpay)} ₽</strong>
          </div>
        </div>
      </section>

      {decision.scenarios.map((s) => (
        <section className={`card scenario scenario-${s.key}`} key={s.key}>
          <h2>
            {s.emoji} {s.title}
          </h2>
          <div className="rows">
            <div className="row">
              <span className="muted">Выручка</span>
              <span>{money(s.revenue)} ₽</span>
            </div>
            <div className="row">
              <span className="muted">Расходы</span>
              <span>{money(s.expenses)} ₽</span>
            </div>
            <div className="row">
              <span className="muted">Платёж по кредиту</span>
              <span>{money(s.loan_payment)} ₽</span>
            </div>
            <div className="row">
              <span className="muted">Свободный поток</span>
              <strong className={s.cash_flow < 0 ? 'flow-neg' : 'flow-pos'}>
                {signedMoney(s.cash_flow)} ₽/мес
              </strong>
            </div>
            <div className="row">
              <span className="muted">Долговая нагрузка</span>
              <span>{pct(s.debt_load_pct)}</span>
            </div>
            <div className="row">
              <span className="muted">Резерв через 3 мес.</span>
              <span>{money(s.reserve_after_3m)} ₽</span>
            </div>
            {s.reserve_months !== null ? (
              <div className="row">
                <span className="muted">Резерва хватит на</span>
                <span>{s.reserve_months} мес.</span>
              </div>
            ) : null}
          </div>
          <button
            className="btn"
            onClick={() => {
              setScenarioKey(s.key)
              navigate('/consequences')
            }}
          >
            Последствия →
          </button>
        </section>
      ))}

      <Link className="btn btn-primary" to="/checklist">
        К чек-листу →
      </Link>
    </main>
  )
}
