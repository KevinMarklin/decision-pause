// Типы контракта API — источник: backend/docs/api.md.

export interface Inputs {
  loan_amount: number
  loan_term_months: number
  interest_rate_pct: number
  revenue: number
  expenses: number
  reserve: number
  revenue_growth_pct: number
  expense_growth_pct: number
  purpose: string
}

export interface Credit {
  monthly_payment: number
  total_paid: number
  overpay: number
}

export type ScenarioKey = 'expected' | 'moderate' | 'negative'

export interface Scenario {
  key: ScenarioKey
  title: string
  emoji: string
  revenue: number
  expenses: number
  loan_payment: number
  cash_flow: number
  debt_load_pct: number
  reserve_after_3m: number
  reserve_months: number | null
  consequences: string[]
}

export interface Decision {
  id: string
  created_at: string
  type: 'loan'
  inputs: Inputs
  credit: Credit
  scenarios: Scenario[]
  checklist: string[]
}

export interface HistoryItem {
  id: string
  created_at: string
  type: string
  loan_amount: number
  monthly_payment: number
  cash_flows: Record<ScenarioKey, number>
}

export type FieldErrors = Partial<Record<keyof Inputs | 'unknown_field', string>>

/** Ошибка API: код из тела ответа + опциональные поля валидации. */
export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly fields?: Record<string, string>
  readonly unknownField?: string

  constructor(
    status: number,
    code: string,
    fields?: Record<string, string>,
    unknownField?: string,
  ) {
    super(code)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.fields = fields
    this.unknownField = unknownField
  }
}
