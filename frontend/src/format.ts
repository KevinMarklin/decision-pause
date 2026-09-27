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

/** Дата анализа: «26.09.2026, 17:59». */
export function dateRu(iso: string): string {
  return new Date(iso).toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
