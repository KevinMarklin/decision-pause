import { useState, type ReactNode } from 'react'
import type { Decision, Inputs, MarketingInputs, MarketingResult, ScenarioKey } from './api/types'
import { StoreContext } from './storeContext'

// Контекст и хук вынесены в storeContext.ts / useStore.ts, чтобы файл
// компонента экспортировал только компонент (Fast Refresh).
export function StoreProvider({ children }: { children: ReactNode }) {
  const [draft, setDraft] = useState<Inputs | null>(null)
  const [decision, setDecision] = useState<Decision | null>(null)
  const [scenarioKey, setScenarioKey] = useState<ScenarioKey | null>(null)
  const [marketingDraft, setMarketingDraft] = useState<MarketingInputs | null>(null)
  const [marketingResult, setMarketingResult] = useState<MarketingResult | null>(null)
  return (
    <StoreContext.Provider
      value={{ draft, setDraft, decision, setDecision, scenarioKey, setScenarioKey, marketingDraft, setMarketingDraft, marketingResult, setMarketingResult }}
    >
      {children}
    </StoreContext.Provider>
  )
}
