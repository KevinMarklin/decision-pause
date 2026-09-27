import { Link, Navigate, useNavigate } from 'react-router-dom'
import { useStore } from '../useStore'

export default function Consequences() {
  const navigate = useNavigate()
  const { decision, scenarioKey } = useStore()

  if (!decision || !scenarioKey) return <Navigate to="/" replace />
  const scenario = decision.scenarios.find((s) => s.key === scenarioKey)
  if (!scenario) return <Navigate to="/result" replace />

  return (
    <main className="screen">
      <Link className="nav-link" to="/result">
        ← К результату
      </Link>
      <h1>
        {scenario.emoji} {scenario.title}: последствия
      </h1>

      <div className="card consequences">
        {scenario.consequences.map((text) => (
          <p key={text}>{text}</p>
        ))}
      </div>

      <button className="btn btn-primary" onClick={() => navigate('/checklist')}>
        Далее: чек-лист →
      </button>
    </main>
  )
}
