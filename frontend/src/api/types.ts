export type Role = 'admin' | 'viewer'

export type Status = 'active' | 'expiring' | 'expired' | 'cancelled' | 'archived' | 'paid'

export type BillingPeriod = 'one_time' | 'monthly' | 'yearly'

export type AttrType = 'string' | 'number' | 'boolean'

export type TokenPair = {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export type PublicUser = {
  id: string
  email: string
  role: Role
}

export type AttrField = {
  key: string
  label: string
  type: AttrType
  required: boolean
}

export type Kind = {
  id: string
  slug: string
  name: string
  color: string
  attr_schema: AttrField[]
}

export type Category = {
  id: string
  parent_id: string | null
  name: string
  sort_order: number
  children: Category[]
}

export type Item = {
  id: string
  title: string
  description: string
  kind_id: string
  category_id: string | null
  vendor: string
  tags: string[]
  cost_amount: number
  currency: string
  billing_period: BillingPeriod
  started_at: string | null
  expires_at: string
  notify_before_days: number | null
  url: string
  account_hint: string
  status: Status
  attrs: Record<string, unknown>
  created_at: string
  updated_at: string
}

export type Renewal = {
  id: string
  item_id: string
  actor_id: string
  old_expires_at: string
  new_expires_at: string
  old_cost: number
  new_cost: number
  comment: string
  created_at: string
}

export type ItemList = {
  items: Item[]
  page: number
  per_page: number
  total: number
}

export type ItemCard = {
  item: Item
  renewals: Renewal[]
}

export type ItemFilter = {
  q?: string
  kind_id?: string
  status?: string
  category_id?: string
  vendor?: string
  expires_from?: string
  expires_to?: string
  cost_from?: string
  cost_to?: string
  billing_period?: string
  tag?: string
  sort?: string
  order?: string
  page?: number
  per_page?: number
}

export type ItemWrite = {
  title: string
  description?: string
  kind_id: string
  category_id?: string | null
  vendor?: string
  tags?: string[]
  cost_amount?: number
  currency?: string
  billing_period?: BillingPeriod
  started_at?: string
  expires_at: string
  notify_before_days?: number | null
  url?: string
  account_hint?: string
  status?: 'cancelled' | 'archived' | 'paid' | 'active'
  attrs?: Record<string, unknown>
}

export type Dashboard = {
  counts: {
    active: number
    expiring_7: number
    expiring_30: number
    expired: number
  }
  upcoming_cost: { currency: string; monthly: number; yearly: number }[]
  expirations_by_month: { month: string; count: number; amounts: { currency: string; amount: number }[] }[]
  cost_by_kind: { kind_id: string; currency: string; amount: number }[]
  soonest: { id: string; title: string; expires_at: string; status: Status; kind_id: string }[]
}

export type CalendarMonth = {
  year: number
  month: number
  days: { date: string; items: { id: string; title: string; status: Status }[] }[]
}

export type Notification = {
  id: string
  item_id: string
  to_status: Status
  title: string
  read_at: string | null
  created_at: string
}

export type NotificationList = {
  items: Notification[]
  page: number
  per_page: number
  total: number
}

export type AuditEntry = {
  id: string
  actor_id: string | null
  action: string
  entity: string
  entity_id: string
  before_json: unknown
  after_json: unknown
  created_at: string
}

export type AuditList = {
  items: AuditEntry[]
  page: number
  per_page: number
  total: number
}

export type CSVImportPreview = {
  rows: number
  valid: number
  errors: { line: number; message: string }[]
  preview: { title: string; kind_slug: string; expires_at: string; attrs?: Record<string, unknown> }[]
}

export type CSVImportResult = {
  created: number
}
