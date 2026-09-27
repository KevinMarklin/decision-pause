import type { Inputs } from './api/types'

// Описание полей анкеты: единый источник для формы и экрана проверки.
export type FormValues = Record<keyof Inputs, string>
export type FieldErrors = Partial<Record<keyof Inputs, string>>

export interface FieldDef {
  name: keyof Inputs
  label: string
  unit?: string
  placeholder?: string
  inputMode: 'numeric' | 'decimal' | 'text'
  integer?: boolean
}

export const EMPTY_VALUES: FormValues = {
  loan_amount: '',
  loan_term_months: '',
  interest_rate_pct: '',
  revenue: '',
  expenses: '',
  reserve: '',
  revenue_growth_pct: '',
  expense_growth_pct: '',
  purpose: '',
}

export const FIELDS: FieldDef[] = [
  { name: 'loan_amount', label: 'Сумма кредита', unit: '₽', placeholder: '2000000', inputMode: 'numeric', integer: true },
  { name: 'loan_term_months', label: 'Срок кредита', unit: 'мес', placeholder: '36', inputMode: 'numeric', integer: true },
  { name: 'interest_rate_pct', label: 'Ставка', unit: '%', placeholder: '20', inputMode: 'decimal' },
  { name: 'revenue', label: 'Выручка в месяц', unit: '₽', placeholder: '800000', inputMode: 'numeric', integer: true },
  { name: 'expenses', label: 'Расходы в месяц', unit: '₽', placeholder: '600000', inputMode: 'numeric', integer: true },
  { name: 'reserve', label: 'Резерв', unit: '₽', placeholder: '500000', inputMode: 'numeric', integer: true },
  { name: 'revenue_growth_pct', label: 'Рост выручки за 3 мес.', unit: '%', placeholder: '30', inputMode: 'decimal' },
  { name: 'expense_growth_pct', label: 'Рост расходов за 3 мес.', unit: '%', placeholder: '15', inputMode: 'decimal' },
  { name: 'purpose', label: 'Цель кредита', placeholder: 'Закупка оборудования', inputMode: 'text' },
]

export function parseNumber(v: string): number {
  return Number(v.trim().replace(',', '.'))
}
