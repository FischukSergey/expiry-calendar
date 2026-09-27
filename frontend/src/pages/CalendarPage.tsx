import { useMemo, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type ColumnDef } from '@tanstack/react-table'

import { getCalendar, payItemOccurrence, unpayItemOccurrence } from '../api/endpoints.ts'
import type { CalendarItem, OccurrenceStatus } from '../api/types.ts'
import { DataTable } from '../components/DataTable.tsx'
import { Button, OccurrenceBadge, PageState, PageTitle } from '../components/ui.tsx'
import { useAuth } from '../hooks/useAuth.ts'
import { formatDate, formatMoney } from '../lib/format.ts'

const weekdays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

const occDot: Record<OccurrenceStatus, string> = {
  open: 'bg-amber-400',
  paid: 'bg-sky-400',
}

type CalendarRow = { date: string; item: CalendarItem }

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
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const [params, setParams] = useSearchParams()
  const year = Number(params.get('year') ?? now.getUTCFullYear())
  const month = Number(params.get('month') ?? now.getUTCMonth() + 1)
  const [selected, setSelected] = useState<string | null>(null)
  const [payError, setPayError] = useState<string | null>(null)

  const cal = useQuery({
    queryKey: ['calendar', year, month],
    queryFn: () => getCalendar(year, month),
  })

  const days = cal.data?.days
  const byDate = useMemo(() => {
    const map = new Map<string, CalendarItem[]>()
    for (const day of days ?? []) {
      map.set(day.date, day.items)
    }
    return map
  }, [days])

  const monthRows = useMemo(() => {
    const rows: CalendarRow[] = []
    for (const day of days ?? []) {
      for (const item of day.items) {
        rows.push({ date: day.date, item })
      }
    }
    return rows
  }, [days])

  const invalidate = async () => {
    await qc.invalidateQueries({ queryKey: ['calendar'] })
    await qc.invalidateQueries({ queryKey: ['dashboard'] })
    await qc.invalidateQueries({ queryKey: ['item'] })
    await qc.invalidateQueries({ queryKey: ['items'] })
  }

  const pay = useMutation({
    mutationFn: ({ id, date }: { id: string; date: string }) => payItemOccurrence(id, date),
    onSuccess: async () => {
      setPayError(null)
      await invalidate()
    },
    onError: (err) => {
      setPayError(err.message)
    },
  })

  const unpay = useMutation({
    mutationFn: ({ id, date }: { id: string; date: string }) => unpayItemOccurrence(id, date),
    onSuccess: async () => {
      setPayError(null)
      await invalidate()
    },
    onError: (err) => {
      setPayError(err.message)
    },
  })

  const shift = (delta: number) => {
    const d = new Date(Date.UTC(year, month - 1 + delta, 1))
    setParams({ year: String(d.getUTCFullYear()), month: String(d.getUTCMonth() + 1) })
    setSelected(null)
    setPayError(null)
  }

  const visibleRows = selected ? monthRows.filter((row) => row.date === selected) : monthRows
  const busy = pay.isPending || unpay.isPending
  const columns = useMemo<ColumnDef<CalendarRow>[]>(() => {
    const defs: ColumnDef<CalendarRow>[] = [
      {
        accessorKey: 'date',
        header: 'Дата',
        cell: ({ row }) => <span className="whitespace-nowrap tabular-nums">{formatDate(row.original.date)}</span>,
        meta: { className: 'w-36' },
      },
      {
        id: 'title',
        header: 'Запись',
        accessorFn: (row) => row.item.title,
        cell: ({ row }) => (
          <Link to={`/items/${row.original.item.id}`} className="font-medium text-teal-300 hover:underline">
            {row.original.item.title}
          </Link>
        ),
      },
      {
        id: 'amount',
        header: 'Сумма',
        accessorFn: (row) => row.item.cost_amount,
        cell: ({ row }) => (
          <span className="whitespace-nowrap tabular-nums">
            {formatMoney(row.original.item.cost_amount, row.original.item.currency)}
          </span>
        ),
        meta: { align: 'right' },
      },
      {
        id: 'status',
        header: 'Статус',
        cell: ({ row }) => <OccurrenceBadge status={row.original.item.occurrence_status} />,
        meta: { className: 'w-36' },
      },
    ]
    if (isAdmin) {
      defs.push({
        id: 'pay',
        header: '',
        cell: ({ row }) =>
          row.original.item.occurrence_status === 'open' ? (
            <Button
              type="button"
              disabled={busy}
              onClick={() => pay.mutate({ id: row.original.item.id, date: row.original.date })}
            >
              Оплачено
            </Button>
          ) : (
            <Button
              type="button"
              variant="outline"
              disabled={busy}
              onClick={() => unpay.mutate({ id: row.original.item.id, date: row.original.date })}
            >
              Снять оплату
            </Button>
          ),
        meta: { align: 'right', className: 'w-40' },
      })
    }
    return defs
  }, [busy, isAdmin, pay, unpay])

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
        <div className="space-y-6">
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
                    onClick={() => {
                      if (items.length === 0) {
                        setSelected(null)
                        return
                      }
                      setSelected((cur) => (cur === date ? null : date))
                    }}
                    className={`min-h-10 rounded-lg border p-0.5 text-left text-xs sm:min-h-16 sm:p-1 sm:text-sm ${
                      active ? 'border-teal-500 bg-teal-500/10' : 'border-transparent hover:bg-slate-900'
                    }`}
                  >
                    <span className="text-slate-300">{day}</span>
                    {items.length > 0 ? (
                      <span className="mt-1 flex flex-wrap gap-1">
                        {items.slice(0, 3).map((it) => (
                          <span key={it.id} className={`h-1.5 w-1.5 rounded-full ${occDot[it.occurrence_status]}`} />
                        ))}
                      </span>
                    ) : null}
                  </button>
                )
              })}
            </div>
          </div>
          <section className="rounded-xl border border-slate-800">
            <div className={visibleRows.length > 0 ? 'border-b border-slate-800 px-4 py-3' : 'px-4 py-3'}>
              <h2 className="text-sm font-medium text-slate-300">{selected ? formatDate(selected) : 'Весь месяц'}</h2>
              {payError ? <p className="mt-2 text-sm text-rose-300">{payError}</p> : null}
              {visibleRows.length === 0 ? (
                <p className="mt-2 text-sm text-slate-500">{selected ? 'Пусто' : 'В этом месяце нет вхождений'}</p>
              ) : null}
            </div>
            {visibleRows.length > 0 ? (
              <DataTable
                columns={columns}
                data={visibleRows}
                getRowId={(row) => `${row.date}-${row.item.id}`}
                minWidthClass="min-w-[40rem]"
              />
            ) : null}
          </section>
        </div>
      ) : null}
    </div>
  )
}
