import type { TokenPair } from './types.ts'

export class ApiError extends Error {
  readonly status: number
  readonly code: string
  readonly details?: Record<string, unknown>

  constructor(status: number, code: string, message: string, details?: Record<string, unknown>) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.details = details
  }
}

let accessToken: string | null = null
let refreshInFlight: Promise<boolean> | null = null
let onAuthLost: (() => void) | null = null
const accessListeners = new Set<(token: string | null) => void>()

export function setOnAuthLost(fn: (() => void) | null): void {
  onAuthLost = fn
}

export function getAccessToken(): string | null {
  return accessToken
}

export function subscribeAccess(fn: (token: string | null) => void): () => void {
  accessListeners.add(fn)
  return () => {
    accessListeners.delete(fn)
  }
}

export function setAccessToken(token: string | null): void {
  if (accessToken === token) {
    return
  }
  accessToken = token
  for (const fn of accessListeners) {
    fn(token)
  }
}

function isAuthPath(path: string): boolean {
  return path.startsWith('/api/v1/auth/')
}

export async function refreshSession(): Promise<boolean> {
  if (refreshInFlight) {
    return refreshInFlight
  }
  refreshInFlight = (async () => {
    try {
      const res = await fetch('/api/v1/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: '{}',
      })
      if (!res.ok) {
        setAccessToken(null)
        return false
      }
      const pair = (await res.json()) as TokenPair
      setAccessToken(pair.access_token)
      return true
    } catch {
      setAccessToken(null)
      return false
    } finally {
      refreshInFlight = null
    }
  })()
  return refreshInFlight
}

export type QueryValue = string | number | boolean | undefined

export function toQuery(params: Record<string, QueryValue>): string {
  const q = new URLSearchParams()
  for (const [key, value] of Object.entries(params)) {
    if (value === undefined || value === '') {
      continue
    }
    q.set(key, String(value))
  }
  const s = q.toString()
  return s ? `?${s}` : ''
}

type RequestOpts = {
  method?: string
  body?: unknown
  query?: Record<string, QueryValue>
  auth?: boolean
  credentials?: RequestCredentials
}

async function parseError(res: Response, text: string): Promise<ApiError> {
  let code = 'internal'
  let message = text || res.statusText
  let details: Record<string, unknown> | undefined
  try {
    const parsed = JSON.parse(text) as {
      error?: { code?: string; message?: string; details?: Record<string, unknown> }
    }
    if (parsed.error) {
      code = parsed.error.code ?? code
      message = parsed.error.message ?? message
      details = parsed.error.details
    }
  } catch {
    /* сырой ответ */
  }
  return new ApiError(res.status, code, message, details)
}

async function send(path: string, opts: RequestOpts): Promise<Response> {
  const url = `${path}${opts.query ? toQuery(opts.query) : ''}`
  const headers = new Headers()
  const isForm = opts.body instanceof FormData
  if (opts.body !== undefined && !isForm) {
    headers.set('Content-Type', 'application/json')
  }
  if (opts.auth !== false && accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }
  return fetch(url, {
    method: opts.method ?? 'GET',
    headers,
    body: isForm ? (opts.body as FormData) : opts.body !== undefined ? JSON.stringify(opts.body) : undefined,
    credentials: opts.credentials ?? 'same-origin',
  })
}

export async function api<T>(path: string, opts: RequestOpts = {}): Promise<T> {
  let res = await send(path, opts)
  if (res.status === 401 && opts.auth !== false && !isAuthPath(path)) {
    const ok = await refreshSession()
    if (ok) {
      res = await send(path, opts)
    } else {
      onAuthLost?.()
    }
  }
  if (res.status === 204) {
    return undefined as T
  }
  const text = await res.text()
  if (!res.ok) {
    throw await parseError(res, text)
  }
  if (!text) {
    return undefined as T
  }
  return JSON.parse(text) as T
}

export async function apiBlob(path: string, query?: Record<string, QueryValue>): Promise<Blob> {
  const opts: RequestOpts = { query }
  let res = await send(path, opts)
  if (res.status === 401) {
    const ok = await refreshSession()
    if (ok) {
      res = await send(path, opts)
    }
  }
  if (!res.ok) {
    throw await parseError(res, await res.text())
  }
  return res.blob()
}
