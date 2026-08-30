import { useMemo } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'

import { exportItems, listCategories, listItems, listKinds } from '../api/endpoints.ts'
import type { ItemFilter, Status } from '../api/types.ts'
import { Button, Field, PageState, PageTitle, Select, StatusBadge, TextInput } from '../components/ui.tsx'
import { useAuth } from '../hooks/useAuth.ts'
import { billingLabel, flattenCategories, formatDate, formatMoney, statusLabel } from '../lib/format.ts'

function readFilter(params: URLSearchParams): ItemFilter {
  const page = Number(params.get('page') ?? '1')
  return {
    q: params.get('q') ?? undefined,
    kind_id: params.get('kind_id') ?? undefined,
    status: params.get('status') ?? undefined,
    category_id: params.get('category_id') ?? undefined,
    vendor: params.get('vendor') ?? undefined,
    expires_from: params.get('expires_from') ?? undefined,
    expires_to: params.get('expires_to') ?? undefined,
    billing_period: params.get('billing_period') ?? undefined,
    tag: params.get('tag') ?? undefined,
    sort: params.get('sort') ?? 'expires_at',
    order: params.get('order') ?? 'asc',
    page: Number.isFinite(page) && page > 0 ? page : 1,
    per_page: 20,
  }
}

export function ItemsPage() {
  const { isAdmin } = useAuth()
  const [params, setParams] = useSearchParams()
  const filter = useMemo(() => readFilter(params), [params])

  const kinds = useQuery({ queryKey: ['kinds'], queryFn: listKinds })
  const cats = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const items = useQuery({ queryKey: ['items', filter], queryFn: () => listItems(filter) })

  const setField = (key: string, value: string) => {
    const next = new URLSearchParams(params)
    if (value) {
      next.set(key, value)
    } else {
      next.delete(key)
    }
    next.delete('page')
    setParams(next)
  }

  const exportCurrent = async () => {
    const rest = { ...filter }
    Reflect.deleteProperty(rest, 'page')
    Reflect.deleteProperty(rest, 'per_page')
    const blob = await exportItems(rest)
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = 'items.csv'
    a.click()
    URL.revokeObjectURL(url)
  }

  const flatCats = flattenCategories(cats.data?.items ?? [])
  const kindById = new Map((kinds.data?.items ?? []).map((k) => [k.id, k]))
  const total = items.data?.total ?? 0
  const page = items.data?.page ?? 1
  const pages = Math.max(1, Math.ceil(total / (items.data?.per_page ?? 20)))

  return (
    <div>
      <PageTitle
        title="Записи"
        subtitle={`${total} всего`}
        actions={
          <>
            <Button variant="outline" type="button" onClick={() => void exportCurrent()}>
              Экспорт CSV
            </Button>
            {isAdmin ? (
              <Link to="/items/new">
                <Button type="button">Новая запись</Button>
              </Link>
            ) : null}
          </>
        }
      />

      <div className="mb-4 grid gap-3 rounded-xl border border-slate-800 bg-slate-900/40 p-4 sm:grid-cols-2 lg:grid-cols-4">
        <Field label="Поиск">
          <TextInput
            defaultValue={filter.q ?? ''}
            placeholder="название, вендор, тег"
            onBlur={(e) => setField('q', e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter') {
                setField('q', (e.target as HTMLInputElement).value)
              }
            }}
          />
        </Field>
        <Field label="Статус">
          <Select value={filter.status ?? ''} onChange={(e) => setField('status', e.target.value)}>
            <option value="">Все</option>
            {(['active', 'expiring', 'expired', 'cancelled', 'archived'] as Status[]).map((s) => (
              <option key={s} value={s}>
                {statusLabel[s]}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Тип записи" hint="Вид платежа">
          <Select value={filter.kind_id ?? ''} onChange={(e) => setField('kind_id', e.target.value)}>
            <option value="">Все</option>
            {(kinds.data?.items ?? []).map((k) => (
              <option key={k.id} value={k.id}>
                {k.name}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Раздел" hint="Папка в дереве">
          <Select value={filter.category_id ?? ''} onChange={(e) => setField('category_id', e.target.value)}>
            <option value="">Все</option>
            {flatCats.map((c) => (
              <option key={c.id} value={c.id}>
                {'· '.repeat(c.depth)}
                {c.name}
              </option>
            ))}
          </Select>
        </Field>
        <Field label="Период">
          <Select value={filter.billing_period ?? ''} onChange={(e) => setField('billing_period', e.target.value)}>
            <option value="">Все</option>
            <option value="one_time">{billingLabel.one_time}</option>
            <option value="monthly">{billingLabel.monthly}</option>
            <option value="yearly">{billingLabel.yearly}</option>
          </Select>
        </Field>
        <Field label="Срок с">
          <TextInput type="date" value={filter.expires_from ?? ''} onChange={(e) => setField('expires_from', e.target.value)} />
        </Field>
        <Field label="Срок по">
          <TextInput type="date" value={filter.expires_to ?? ''} onChange={(e) => setField('expires_to', e.target.value)} />
        </Field>
        <Field label="Сортировка">
          <Select
            value={`${filter.sort}:${filter.order}`}
            onChange={(e) => {
              const [sort, order] = e.target.value.split(':')
              const next = new URLSearchParams(params)
              next.set('sort', sort ?? 'expires_at')
              next.set('order', order ?? 'asc')
              next.delete('page')
              setParams(next)
            }}
          >
            <option value="expires_at:asc">Срок оплаты ↑</option>
            <option value="expires_at:desc">Срок оплаты ↓</option>
            <option value="title:asc">Название</option>
            <option value="cost_amount:desc">Сумма ↓</option>
            <option value="updated_at:desc">Обновлено</option>
          </Select>
        </Field>
      </div>

      {items.isPending ? <PageState title="Загрузка списка…" /> : null}
      {items.isError ? (
        <PageState title="Ошибка списка" hint={items.error.message} onRetry={() => void items.refetch()} />
      ) : null}
      {items.data && items.data.items.length === 0 ? (
        <PageState
          title="Пусто"
          hint="Измените фильтр или создайте запись"
          action={
            isAdmin ? (
              <Link to="/items/new">
                <Button type="button">Создать</Button>
              </Link>
            ) : undefined
          }
        />
      ) : null}

      {items.data && items.data.items.length > 0 ? (
        <>
          <ul className="space-y-2 md:hidden">
            {items.data.items.map((it) => (
              <li key={it.id} className="rounded-xl border border-slate-800 bg-slate-900/40 p-3">
                <Link to={`/items/${it.id}`} className="font-medium text-teal-300">
                  {it.title}
                </Link>
                <p className="mt-1 text-xs text-slate-400">
                  {kindById.get(it.kind_id)?.name ?? '—'} · {formatDate(it.expires_at)}
                </p>
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  <StatusBadge status={it.status} />
                  <span className="text-xs text-slate-400">{formatMoney(it.cost_amount, it.currency)}</span>
                </div>
              </li>
            ))}
          </ul>
          <div className="hidden overflow-x-auto rounded-xl border border-slate-800 md:block">
            <table className="w-full min-w-[720px] text-left text-sm">
              <thead className="bg-slate-900 text-xs tracking-wide text-slate-400 uppercase">
                <tr>
                  <th className="px-3 py-2 font-medium">Запись</th>
                  <th className="px-3 py-2 font-medium">Тип</th>
                  <th className="px-3 py-2 font-medium">Срок оплаты</th>
                  <th className="px-3 py-2 font-medium">Сумма</th>
                  <th className="px-3 py-2 font-medium">Статус</th>
                </tr>
              </thead>
              <tbody>
                {items.data.items.map((it) => (
                  <tr key={it.id} className="border-t border-slate-800 hover:bg-slate-900/60">
                    <td className="px-3 py-2">
                      <Link to={`/items/${it.id}`} className="font-medium text-teal-300 hover:underline">
                        {it.title}
                      </Link>
                      {it.vendor ? <p className="text-xs text-slate-500">{it.vendor}</p> : null}
                    </td>
                    <td className="px-3 py-2 text-slate-300">{kindById.get(it.kind_id)?.name ?? '—'}</td>
                    <td className="px-3 py-2">{formatDate(it.expires_at)}</td>
                    <td className="px-3 py-2">{formatMoney(it.cost_amount, it.currency)}</td>
                    <td className="px-3 py-2">
                      <StatusBadge status={it.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      ) : null}

      {pages > 1 ? (
        <div className="mt-4 flex items-center justify-between text-sm text-slate-400">
          <span>
            стр. {page} / {pages}
          </span>
          <div className="flex gap-2">
            <Button
              variant="outline"
              type="button"
              disabled={page <= 1}
              onClick={() => {
                const next = new URLSearchParams(params)
                next.set('page', String(page - 1))
                setParams(next)
              }}
            >
              Назад
            </Button>
            <Button
              variant="outline"
              type="button"
              disabled={page >= pages}
              onClick={() => {
                const next = new URLSearchParams(params)
                next.set('page', String(page + 1))
                setParams(next)
              }}
            >
              Дальше
            </Button>
          </div>
        </div>
      ) : null}
    </div>
  )
}
