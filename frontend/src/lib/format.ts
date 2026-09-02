import type { BillingPeriod, Category, Status } from '../api/types.ts'

export const statusLabel: Record<Status, string> = {
  active: 'Активно',
  expiring: 'Скоро срок',
  expired: 'Просрочено',
  cancelled: 'Отменено',
  archived: 'Архив',
  paid: 'Оплачено',
}

/** Лист дерева «Категории», который обычно совпадает с типом записи. */
const kindDefaultCategory: Record<string, string> = {
  domain: 'Домены',
  subscription: 'Подписки',
  license: 'Лицензии',
  tax: 'Налоги',
  rent: 'Аренда',
  contract: 'Договоры',
  insurance: 'Страховки',
  vehicle: 'Авто',
  mobile: 'Связь',
}

export function suggestCategoryId(kindSlug: string, cats: FlatCategory[]): string {
  const name = kindDefaultCategory[kindSlug]
  if (!name) {
    return ''
  }
  return cats.find((c) => c.name === name)?.id ?? ''
}

export const billingLabel: Record<BillingPeriod, string> = {
  one_time: 'Разово',
  monthly: 'Ежемесячно',
  yearly: 'Ежегодно',
}

export function formatDate(iso: string): string {
  if (!iso) {
    return '—'
  }
  const day = iso.length === 10 ? `${iso}T00:00:00Z` : iso
  const d = new Date(day)
  if (Number.isNaN(d.getTime())) {
    return iso
  }
  return d.toLocaleDateString('ru-RU', { timeZone: 'UTC' })
}

export function formatDateTime(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) {
    return iso
  }
  return d.toLocaleString('ru-RU', { dateStyle: 'short', timeStyle: 'short' })
}

export function formatMoney(amount: number, currency: string): string {
  return `${new Intl.NumberFormat('ru-RU').format(amount)} ${currency}`
}

export function monthLabel(ym: string): string {
  const [y, m] = ym.split('-')
  const d = new Date(Date.UTC(Number(y), Number(m) - 1, 1))
  return d.toLocaleDateString('ru-RU', { month: 'short', year: 'numeric', timeZone: 'UTC' })
}

export type FlatCategory = { id: string; name: string; depth: number }

export function flattenCategories(tree: Category[], depth = 0): FlatCategory[] {
  const out: FlatCategory[] = []
  for (const node of tree) {
    out.push({ id: node.id, name: node.name, depth })
    out.push(...flattenCategories(node.children ?? [], depth + 1))
  }
  return out
}

export function findCategoryName(tree: Category[], id: string | null): string {
  if (!id) {
    return '—'
  }
  for (const node of flattenCategories(tree)) {
    if (node.id === id) {
      return node.name
    }
  }
  return '—'
}

export function parseCSVHeaders(text: string): string[] {
  const first = text.split(/\r?\n/, 1)[0] ?? ''
  const out: string[] = []
  let cur = ''
  let quoted = false
  for (const ch of first) {
    if (ch === '"') {
      quoted = !quoted
      continue
    }
    if (ch === ',' && !quoted) {
      out.push(cur.trim())
      cur = ''
      continue
    }
    cur += ch
  }
  out.push(cur.trim())
  return out.filter(Boolean)
}

export const csvFields: { key: string; label: string }[] = [
  { key: 'title', label: 'Название' },
  { key: 'kind_slug', label: 'Тип (slug)' },
  { key: 'expires_at', label: 'Срок оплаты' },
  { key: 'cost_amount', label: 'Сумма' },
  { key: 'currency', label: 'Валюта' },
  { key: 'vendor', label: 'Поставщик' },
  { key: 'billing_period', label: 'Период' },
  { key: 'category_name', label: 'Раздел' },
  { key: 'tags', label: 'Теги' },
  { key: 'status', label: 'Статус' },
  { key: 'notify_before_days', label: 'Напомнить за (дней)' },
]
