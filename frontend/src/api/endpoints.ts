import { api, apiBlob, type QueryValue } from './client.ts'
import type {
  AuditList,
  CalendarMonth,
  Category,
  CSVImportPreview,
  CSVImportResult,
  Dashboard,
  Item,
  ItemCard,
  ItemFilter,
  ItemList,
  ItemWrite,
  Kind,
  NotificationList,
  PublicUser,
  TokenPair,
} from './types.ts'

export function login(email: string, password: string): Promise<TokenPair> {
  return api<TokenPair>('/api/v1/auth/login', {
    method: 'POST',
    body: { email, password },
    auth: false,
    credentials: 'include',
  })
}

export function register(email: string, password: string): Promise<TokenPair> {
  return api<TokenPair>('/api/v1/auth/register', {
    method: 'POST',
    body: { email, password },
    auth: false,
    credentials: 'include',
  })
}

export function logout(): Promise<void> {
  return api<void>('/api/v1/auth/logout', { method: 'POST', credentials: 'include' })
}

export function logoutAll(): Promise<void> {
  return api<void>('/api/v1/auth/logout-all', { method: 'POST' })
}

export function getMe(): Promise<PublicUser> {
  return api<PublicUser>('/api/v1/me')
}

export function listKinds(): Promise<{ items: Kind[] }> {
  return api<{ items: Kind[] }>('/api/v1/kinds')
}

export function listCategories(): Promise<{ items: Category[] }> {
  return api<{ items: Category[] }>('/api/v1/categories')
}

export function createCategory(body: { name: string; parent_id?: string | null; sort_order?: number }): Promise<Category> {
  return api<Category>('/api/v1/categories', { method: 'POST', body })
}

export function patchCategory(
  id: string,
  body: { name?: string; parent_id?: string | null; sort_order?: number },
): Promise<Category> {
  return api<Category>(`/api/v1/categories/${id}`, { method: 'PATCH', body })
}

export function deleteCategory(id: string): Promise<void> {
  return api<void>(`/api/v1/categories/${id}`, { method: 'DELETE' })
}

export function listItems(filter: ItemFilter): Promise<ItemList> {
  return api<ItemList>('/api/v1/items', { query: filter as Record<string, QueryValue> })
}

export function getItem(id: string): Promise<ItemCard> {
  return api<ItemCard>(`/api/v1/items/${id}`)
}

export function createItem(body: ItemWrite): Promise<Item> {
  return api<Item>('/api/v1/items', { method: 'POST', body })
}

export function patchItem(id: string, body: Partial<ItemWrite>): Promise<Item> {
  return api<Item>(`/api/v1/items/${id}`, { method: 'PATCH', body })
}

export function deleteItem(id: string): Promise<void> {
  return api<void>(`/api/v1/items/${id}`, { method: 'DELETE' })
}

export function renewItem(
  id: string,
  body: { new_expires_at: string; new_cost?: number; comment?: string },
): Promise<Item> {
  return api<Item>(`/api/v1/items/${id}/renew`, { method: 'POST', body })
}

export function exportItems(filter: Omit<ItemFilter, 'page' | 'per_page'>): Promise<Blob> {
  return apiBlob('/api/v1/items/export', filter as Record<string, QueryValue>)
}

export function importItems(file: File, mapping: Record<string, string>, dryRun: boolean): Promise<CSVImportPreview | CSVImportResult> {
  const fd = new FormData()
  fd.append('file', file)
  fd.append('mapping', JSON.stringify(mapping))
  const query = dryRun ? { dry_run: true } : undefined
  return api<CSVImportPreview | CSVImportResult>('/api/v1/items/import', {
    method: 'POST',
    body: fd,
    query,
  })
}

export function getDashboard(): Promise<Dashboard> {
  return api<Dashboard>('/api/v1/dashboard')
}

export function getCalendar(year: number, month: number): Promise<CalendarMonth> {
  return api<CalendarMonth>('/api/v1/calendar', { query: { year, month } })
}

export function listNotifications(opts: { unread?: boolean; page?: number; per_page?: number }): Promise<NotificationList> {
  return api<NotificationList>('/api/v1/notifications', {
    query: {
      unread: opts.unread ? true : undefined,
      page: opts.page,
      per_page: opts.per_page,
    },
  })
}

export function readNotification(id: string): Promise<void> {
  return api<void>(`/api/v1/notifications/${id}/read`, { method: 'POST' })
}

export function readAllNotifications(): Promise<void> {
  return api<void>('/api/v1/notifications/read-all', { method: 'POST' })
}

export function listAudit(page: number, perPage: number): Promise<AuditList> {
  return api<AuditList>('/api/v1/audit', { query: { page, per_page: perPage } })
}
