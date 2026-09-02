import { useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'

import { getCalendar } from '../api/endpoints.ts'
import type { Status } from '../api/types.ts'
import { Button, PageState, PageTitle, StatusBadge } from '../components/ui.tsx'

const weekdays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

const dot: Record<Status, string> = {
  active: 'bg-emerald-400',
  expiring: 'bg-amber-400',
  expired: 'bg-rose-400',
  cancelled: 'bg-slate-400',
  archived: 'bg-zinc-400',
  paid: 'bg-sky-400',
}

function monthTitle(year: number, month: number): string {
  return new Date(Date.UTC(year, month - 1, 1)).toLocaleDateString('ru-RU', {
    month: 'long',
    year: 'numeric',
    timeZone: 'UTC',
  })
}

function cells(year: number, month: number): (number | null)[] {
  const first = new Date(Date.UTC(year, month - 1, 1))
  const start = (first.getUTCDay() + 6) % 7
  const days = new Date(Date.UTC(year, month, 0)).getUTCDate()
  const out: (number | null)[] = Array.from({ length: start }, () => null)
  for (let d = 1; d <= days; d++) {
    out.push(d)
  }
  while (out.length % 7 !== 0) {
    out.push(null)
  }
  return out
}

export function CalendarPage() {
  const now = new Date()
  const [params, setParams] = useSearchParams()
  const year = Number(params.get('year') ?? now.getUTCFullYear())
  const month = Number(params.get('month') ?? now.getUTCMonth() + 1)
  const [selected, setSelected] = useState<string | null>(null)

  const cal = useQuery({
    queryKey: ['calendar', year, month],
    queryFn: () => getCalendar(year, month),
  })

  const days = cal.data?.days
  const byDate = useMemo(() => {
    const map = new Map<string, NonNullable<typeof days>[number]['items']>()
    for (const day of days ?? []) {
      map.set(day.date, day.items)
    }
    return map
  }, [days])

  const shift = (delta: number) => {
    const d = new Date(Date.UTC(year, month - 1 + delta, 1))
    setParams({ year: String(d.getUTCFullYear()), month: String(d.getUTCMonth() + 1) })
    setSelected(null)
  }

  const selectedItems = selected ? (byDate.get(selected) ?? []) : []

  return (
    <div>
      <PageTitle
        title="Календарь"
        subtitle={monthTitle(year, month)}
        actions={
          <>
            <Button type="button" variant="outline" onClick={() => shift(-1)}>
              ←
            </Button>
            <Button type="button" variant="outline" onClick={() => shift(1)}>
              →
            </Button>
          </>
        }
      />

      {cal.isPending ? <PageState title="Загрузка месяца…" /> : null}
      {cal.isError ? (
        <PageState title="Ошибка календаря" hint={cal.error.message} onRetry={() => void cal.refetch()} />
      ) : null}

      {cal.data ? (
        <div className="grid gap-6 lg:grid-cols-[1fr_280px]">
          <div className="rounded-xl border border-slate-800 p-3">
            <div className="grid grid-cols-7 text-center text-[10px] text-slate-500 sm:text-xs">
              {weekdays.map((d) => (
                <div key={d} className="py-1 sm:py-2">
                  <span className="sm:hidden">{d.slice(0, 1)}</span>
                  <span className="hidden sm:inline">{d}</span>
                </div>
              ))}
            </div>
            <div className="grid grid-cols-7 gap-0.5 sm:gap-1">
              {cells(year, month).map((day, i) => {
                if (!day) {
                  return <div key={`e-${i}`} className="min-h-10 sm:min-h-16" />
                }
                const date = `${year}-${String(month).padStart(2, '0')}-${String(day).padStart(2, '0')}`
                const items = byDate.get(date) ?? []
                const active = selected === date
                return (
                  <button
                    key={date}
                    type="button"
                    onClick={() => setSelected(items.length ? date : null)}
                    className={`min-h-10 rounded-lg border p-0.5 text-left text-xs sm:min-h-16 sm:p-1 sm:text-sm ${
                      active ? 'border-teal-500 bg-teal-500/10' : 'border-transparent hover:bg-slate-900'
                    }`}
                  >
                    <span className="text-slate-300">{day}</span>
                    {items.length > 0 ? (
                      <span className="mt-1 flex flex-wrap gap-1">
                        {items.slice(0, 3).map((it) => (
                          <span key={it.id} className={`h-1.5 w-1.5 rounded-full ${dot[it.status]}`} />
                        ))}
                      </span>
                    ) : null}
                  </button>
                )
              })}
            </div>
          </div>
          <aside className="rounded-xl border border-slate-800 p-4">
            <h2 className="text-sm font-medium text-slate-300">{selected ?? 'День не выбран'}</h2>
            {selected && selectedItems.length === 0 ? <p className="mt-3 text-sm text-slate-500">Пусто</p> : null}
            <ul className="mt-3 space-y-2">
              {selectedItems.map((it) => (
                <li key={it.id} className="rounded-lg bg-slate-900 px-3 py-2">
                  <Link to={`/items/${it.id}`} className="text-sm text-teal-300 hover:underline">
                    {it.title}
                  </Link>
                  <div className="mt-1">
                    <StatusBadge status={it.status} />
                  </div>
                </li>
              ))}
            </ul>
          </aside>
        </div>
      ) : null}
    </div>
  )
}
