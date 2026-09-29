import { ApiError, type Decision, type HistoryItem, type HiringInputs, type HiringResult, type Inputs, type MarketingInputs, type MarketingResult } from './types'

/** Идентичность: прод — сырая подпись initData, dev — legacy id из unsafe. */
function identityHeaders(): Record<string, string> {
  const wa = window.WebApp
  if (wa?.initData) {
    return { 'X-Max-Init-Data': wa.initData }
  }
  const id = wa?.initDataUnsafe?.user?.id
  if (id != null && id > 0) {
    return { 'X-Max-User-Id': String(id) }
  }
  return {}
}

async function parseError(res: Response): Promise<ApiError> {
  let code = 'unknown'
  let fields: Record<string, string> | undefined
  let unknownField: string | undefined
  try {
    const body = (await res.json()) as {
      error?: string
      fields?: Record<string, string>
      field?: string
    }
    if (body.error) code = body.error
    fields = body.fields
    unknownField = body.field
  } catch {
    // не-JSON тело — оставляем код по умолчанию
  }
  return new ApiError(res.status, code, fields, unknownField)
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(path, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...identityHeaders(),
      ...init?.headers,
    },
  })
  if (!res.ok) throw await parseError(res)
  if (res.status === 204) return undefined as T
  return (await res.json()) as T
}

export function postDecision(inputs: Inputs): Promise<Decision> {
  return request<Decision>('/api/decisions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ type: 'loan', inputs }),
  })
}

export function analyzeMarketing(inputs: MarketingInputs): Promise<MarketingResult> {
  return request<MarketingResult>('/api/v1/analyze/marketing', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(inputs) })
}

export function analyzeHiring(inputs: HiringInputs): Promise<HiringResult> {
  return request<HiringResult>('/api/v1/analyze/hiring', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(inputs) })
}

export function getDecision(id: string): Promise<Decision> {
  return request<Decision>(`/api/decisions/${encodeURIComponent(id)}`)
}

export function listDecisions(limit = 20): Promise<{ items: HistoryItem[] }> {
  return request<{ items: HistoryItem[] }>(`/api/decisions?limit=${limit}`)
}

export function deleteDecision(id: string): Promise<void> {
  return request<void>(`/api/decisions/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}
