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

export interface MarketingInputs { budget: number; avg_check: number; margin_pct: number; lead_cost: number; expected_boost_pct: number; current_revenue: number; fixed_expenses: number; reserve: number; is_black_swan?: boolean }
export interface MarketingScenario { key: ScenarioKey; title: string; emoji: string; additional_orders: number; ad_revenue: number; ad_cost: number; ad_profit: number; roi: number; cash_flow: number; status: 'green' | 'yellow' | 'red'; reserve_months: number | null; consequence: string; revenue: number; expenses: number; net_profit: number; margin_pct: number }
export interface MarketingResult { break_even_orders: number; break_even_revenue: number; scenarios: MarketingScenario[]; consequences: string[]; checklist: string[]; is_black_swan_active: boolean; budget?: number; reserve?: number }
export interface HiringInputs { salary: number; tax_rate?: number; overhead_costs?: number; margin_percent: number; average_check?: number }
export interface HiringResult { salary: number; tax_rate: number; overhead_costs: number; total_cost: number; margin_percent: number; break_even_revenue: number; break_even_deals: number; payback_multiplier: number }

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
