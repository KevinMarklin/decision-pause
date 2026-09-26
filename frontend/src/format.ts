// Форматирование чисел для UI — только вывод, расчёт всегда на бэкенде.
export function money(n: number): string {
  return n.toLocaleString('ru-RU')
}

/** Свободный поток: «+275 673» / «−36 327». */
export function signedMoney(n: number): string {
  return n > 0 ? `+${money(n)}` : money(n)
}

export function pct(n: number): string {
  return `${n.toLocaleString('ru-RU', { maximumFractionDigits: 1 })}%`
}
