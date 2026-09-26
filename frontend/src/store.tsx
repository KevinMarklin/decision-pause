import { createContext, useContext, useState, type ReactNode } from 'react'
import type { Decision, Inputs, ScenarioKey } from './api/types'

// Общий стейт пути «анкета → проверка → результат»: данные живут в памяти,
// на перезагрузку их не восстанавливаем (расчёт — быстрый снимок).
interface Store {
  draft: Inputs | null
  setDraft: (d: Inputs) => void
  decision: Decision | null
  setDecision: (d: Decision | null) => void
  scenarioKey: ScenarioKey | null
  setScenarioKey: (k: ScenarioKey | null) => void
}

const StoreContext = createContext<Store | null>(null)

export function StoreProvider({ children }: { children: ReactNode }) {
  const [draft, setDraft] = useState<Inputs | null>(null)
  const [decision, setDecision] = useState<Decision | null>(null)
  const [scenarioKey, setScenarioKey] = useState<ScenarioKey | null>(null)
  return (
    <StoreContext.Provider
      value={{ draft, setDraft, decision, setDecision, scenarioKey, setScenarioKey }}
    >
      {children}
    </StoreContext.Provider>
  )
}

export function useStore(): Store {
  const ctx = useContext(StoreContext)
  if (!ctx) throw new Error('useStore must be used within StoreProvider')
  return ctx
}
