import { createContext } from 'react'
import type { Decision, Inputs, ScenarioKey } from './api/types'

// Общий стейт пути «анкета → проверка → результат»: данные живут в памяти,
// на перезагрузку их не восстанавливаем (расчёт — быстрый снимок).
export interface Store {
  draft: Inputs | null
  setDraft: (d: Inputs | null) => void
  decision: Decision | null
  setDecision: (d: Decision | null) => void
  scenarioKey: ScenarioKey | null
  setScenarioKey: (k: ScenarioKey | null) => void
}

export const StoreContext = createContext<Store | null>(null)
