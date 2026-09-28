import { Link, useNavigate } from 'react-router-dom'

export default function Home() {
  const navigate = useNavigate()
  return (
    <main className="screen">
      <h1>Анти-импульс</h1>
      <p className="disclaimer">
        Мы не принимаем решение за вас — показываем возможные последствия.
      </p>

      <button className="card card-link" onClick={() => navigate('/form')}>
        <span>
          <strong>Взять кредит на развитие</strong>
          <br />
          <span className="muted">Анкета → 3 сценария → последствия</span>
        </span>
      </button>

      <Link className="nav-link" to="/history">
        История анализов
      </Link>
    </main>
  )
}
