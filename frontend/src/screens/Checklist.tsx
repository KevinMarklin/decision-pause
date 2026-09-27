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
        ← К результату
      </Link>
      <h1>Чек-лист перед решением</h1>
      <p className="disclaimer">
        Ответьте на вопросы честно — это последний шаг перед тем, как решить.
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
        Завершить анализ
      </button>
    </main>
  )
}
