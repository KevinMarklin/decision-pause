import { useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { useStore } from '../useStore'

export default function Checklist() {
  const navigate = useNavigate()
  const { decision, setDraft, setDecision, setScenarioKey } = useStore()
  const [checked, setChecked] = useState<boolean[]>(() =>
    decision ? decision.checklist.map(() => false) : [],
  )

  if (!decision) return <Navigate to="/" replace />

  function toggle(i: number) {
    setChecked((prev) => prev.map((v, idx) => (idx === i ? !v : v)))
  }

  function finish() {
    setDraft(null)
    setDecision(null)
    setScenarioKey(null)
    navigate('/')
  }

  return (
    <main className="screen">
      <Link className="nav-link" to="/result">
        в†ђ Рљ СЂРµР·СѓР»СЊС‚Р°С‚Сѓ
      </Link>
      <h1>Р§РµРє-Р»РёСЃС‚ РїРµСЂРµРґ СЂРµС€РµРЅРёРµРј</h1>
      <p className="disclaimer">
        РћС‚РІРµС‚СЊС‚Рµ РЅР° РІРѕРїСЂРѕСЃС‹ С‡РµСЃС‚РЅРѕ вЂ” СЌС‚Рѕ РїРѕСЃР»РµРґРЅРёР№ С€Р°Рі РїРµСЂРµРґ С‚РµРј, РєР°Рє СЂРµС€РёС‚СЊ.
      </p>

      <div className="card checklist">
        {decision.checklist.map((item, i) => (
          <label className="check-item" key={item}>
            <input
              type="checkbox"
              checked={checked[i] ?? false}
              onChange={() => toggle(i)}
            />
            <span>{item}</span>
          </label>
        ))}
      </div>

      <button className="btn btn-primary" onClick={finish}>
        Р—Р°РІРµСЂС€РёС‚СЊ Р°РЅР°Р»РёР·
      </button>
    </main>
  )
}
