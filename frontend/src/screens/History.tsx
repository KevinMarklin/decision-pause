import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { deleteDecision, listDecisions } from '../api/client'
import type { HistoryItem } from '../api/types'
import { dateRu, money, signedMoney } from '../format'

export default function History() {
  const [items, setItems] = useState<HistoryItem[] | null>(null)
  const [error, setError] = useState(false)
  const [tick, setTick] = useState(0)

  useEffect(() => {
    let cancelled = false
    listDecisions(20)
      .then((res) => {
        if (!cancelled) {
          setItems(res.items)
          setError(false)
        }
      })
      .catch(() => {
        if (!cancelled) setError(true)
      })
    return () => {
      cancelled = true
    }
  }, [tick])

  async function remove(id: string) {
    try {
      await deleteDecision(id)
    } catch {
      // 404 = уже удалён — списываем из списка в любом случае
    }
    setItems((prev) => (prev ? prev.filter((it) => it.id !== id) : prev))
  }

  return (
    <main className="screen">
      <Link className="nav-link" to="/">
        ← На главную
      </Link>
      <h1>История анализов</h1>

      {error ? (
        <>
          <p className="field-error">Не удалось загрузить историю.</p>
          <button className="btn" onClick={() => setTick((t) => t + 1)}>
            Повторить
          </button>
        </>
      ) : items === null ? (
        <p className="muted">Загрузка…</p>
      ) : items.length === 0 ? (
        <div className="card">
          <p>Пока пусто.</p>
          <p className="muted">Первый расчёт займёт меньше минуты.</p>
          <Link className="nav-link" to="/form">
            Начать анализ →
          </Link>
        </div>
      ) : (
        items.map((it) => (
          <div className="card history-item" key={it.id}>
            <div className="row">
              <span className="muted">{dateRu(it.created_at)}</span>
              <button
                className="icon-btn"
                aria-label="Удалить анализ"
                onClick={() => void remove(it.id)}
              >
                ✕
              </button>
            </div>
            <div className="row">
              <span>
                <strong>{money(it.loan_amount)} ₽</strong>
              </span>
              <span className="muted">платёж {money(it.monthly_payment)} ₽/мес</span>
            </div>
            <div className="history-flows">
              <span className={it.cash_flows.expected < 0 ? 'flow-neg' : 'flow-pos'}>
                🟢 {signedMoney(it.cash_flows.expected)}
              </span>
              <span className={it.cash_flows.moderate < 0 ? 'flow-neg' : 'flow-pos'}>
                🟡 {signedMoney(it.cash_flows.moderate)}
              </span>
              <span className={it.cash_flows.negative < 0 ? 'flow-neg' : 'flow-pos'}>
                🔴 {signedMoney(it.cash_flows.negative)}
              </span>
            </div>
          </div>
        ))
      )}
    </main>
  )
}
