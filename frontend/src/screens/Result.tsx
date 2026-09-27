import { Link, Navigate, useNavigate } from 'react-router-dom'
import { money, pct, signedMoney } from '../format'
import { useStore } from '../useStore'

export default function Result() {
  const navigate = useNavigate()
  const { decision, setScenarioKey } = useStore()

  if (!decision) return <Navigate to="/" replace />
  const { credit } = decision

  return (
    <main className="screen">
      <Link className="nav-link" to="/">
        в†ђ РќР° РіР»Р°РІРЅСѓСЋ
      </Link>
      <h1>Р РµР·СѓР»СЊС‚Р°С‚</h1>

      <section className="card">
        <h2>рџ’і РљСЂРµРґРёС‚</h2>
        <div className="rows">
          <div className="row">
            <span className="muted">РџР»Р°С‚С‘Р¶ РІ РјРµСЃСЏС†</span>
            <strong>{money(credit.monthly_payment)} в‚Ѕ</strong>
          </div>
          <div className="row">
            <span className="muted">Р’СЃРµРіРѕ РІС‹РїР»Р°С‚</span>
            <strong>{money(credit.total_paid)} в‚Ѕ</strong>
          </div>
          <div className="row">
            <span className="muted">РџРµСЂРµРїР»Р°С‚Р°</span>
            <strong>{money(credit.overpay)} в‚Ѕ</strong>
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
              <span className="muted">Р’С‹СЂСѓС‡РєР°</span>
              <span>{money(s.revenue)} в‚Ѕ</span>
            </div>
            <div className="row">
              <span className="muted">Р Р°СЃС…РѕРґС‹</span>
              <span>{money(s.expenses)} в‚Ѕ</span>
            </div>
            <div className="row">
              <span className="muted">РџР»Р°С‚С‘Р¶ РїРѕ РєСЂРµРґРёС‚Сѓ</span>
              <span>{money(s.loan_payment)} в‚Ѕ</span>
            </div>
            <div className="row">
              <span className="muted">РЎРІРѕР±РѕРґРЅС‹Р№ РїРѕС‚РѕРє</span>
              <strong className={s.cash_flow < 0 ? 'flow-neg' : 'flow-pos'}>
                {signedMoney(s.cash_flow)} в‚Ѕ/РјРµСЃ
              </strong>
            </div>
            <div className="row">
              <span className="muted">Р”РѕР»РіРѕРІР°СЏ РЅР°РіСЂСѓР·РєР°</span>
              <span>{pct(s.debt_load_pct)}</span>
            </div>
            <div className="row">
              <span className="muted">Р РµР·РµСЂРІ С‡РµСЂРµР· 3 РјРµСЃ.</span>
              <span>{money(s.reserve_after_3m)} в‚Ѕ</span>
            </div>
            {s.reserve_months !== null ? (
              <div className="row">
                <span className="muted">Р РµР·РµСЂРІР° С…РІР°С‚РёС‚ РЅР°</span>
                <span>{s.reserve_months} РјРµСЃ.</span>
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
            РџРѕСЃР»РµРґСЃС‚РІРёСЏ в†’
          </button>
        </section>
      ))}

      <Link className="btn btn-primary" to="/checklist">
        Рљ С‡РµРє-Р»РёСЃС‚Сѓ в†’
      </Link>
    </main>
  )
}
