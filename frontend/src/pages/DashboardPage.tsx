import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { Bar, BarChart, CartesianGrid, Cell, Pie, PieChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts'

import type { Dashboard } from '../api/types.ts'
import { getDashboard, listKinds } from '../api/endpoints.ts'
import { PageState, PageTitle, StatusBadge } from '../components/ui.tsx'
import { formatDate, formatMoney, monthLabel } from '../lib/format.ts'

const kpi = [
  { key: 'active' as const, label: 'Активные' },
  { key: 'expiring_7' as const, label: '7 дней' },
  { key: 'expiring_30' as const, label: '30 дней' },
  { key: 'expired' as const, label: 'Просрочены' },
]

export function DashboardPage() {
  const dash = useQuery({ queryKey: ['dashboard'], queryFn: getDashboard })
  const kinds = useQuery({ queryKey: ['kinds'], queryFn: listKinds })
  const [currency, setCurrency] = useState<string | null>(null)

  const kindName = useMemo(() => {
    const map = new Map<string, { name: string; color: string }>()
    for (const k of kinds.data?.items ?? []) {
      map.set(k.id, { name: k.name, color: k.color })
    }
    return map
  }, [kinds.data])

  if (dash.isPending) {
    return <PageState title="Загрузка обзора…" />
  }
  if (dash.isError) {
    return <PageState title="Не удалось загрузить дашборд" hint={dash.error.message} onRetry={() => void dash.refetch()} />
  }

  const data = dash.data
  const currencies = collectCurrencies(data)
  const selected = currency ?? (currencies.includes('RUB') ? 'RUB' : currencies[0] ?? 'RUB')
  const pie = data.cost_by_kind
    .filter((row) => row.currency === selected)
    .map((row) => ({
      name: kindName.get(row.kind_id)?.name ?? row.kind_id,
      value: row.amount,
      color: kindName.get(row.kind_id)?.color ?? '#14b8a6',
    }))
  const bars = data.expirations_by_month.map((row) => ({
    month: monthLabel(row.month),
    amount: row.amounts.find((a) => a.currency === selected)?.amount ?? 0,
  }))

  return (
    <div>
      <PageTitle title="Обзор" subtitle="Сроки оплаты и расходы без конвертации валют" />

      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
        {kpi.map((card) => (
          <div key={card.key} className="rounded-xl border border-slate-800 bg-slate-900/50 p-4">
            <p className="text-xs tracking-wide text-slate-400 uppercase">{card.label}</p>
            <p className="mt-2 text-3xl font-semibold">{data.counts[card.key]}</p>
          </div>
        ))}
      </div>

      <div className="mt-4 grid gap-3 lg:grid-cols-2">
        {data.upcoming_cost.length === 0 ? (
          <p className="text-sm text-slate-500">Нет регулярных расходов в открытых записях</p>
        ) : null}
        {data.upcoming_cost.map((row) => (
          <div key={row.currency} className="rounded-xl border border-slate-800 bg-slate-900/50 p-4">
            <p className="text-xs tracking-wide text-slate-400 uppercase">Расход · {row.currency}</p>
            <p className="mt-2 text-lg">
              {formatMoney(row.monthly, row.currency)}
              <span className="text-slate-500"> / мес</span>
            </p>
            <p className="text-sm text-slate-400">{formatMoney(row.yearly, row.currency)} / год</p>
          </div>
        ))}
      </div>

      <div className="mt-6 grid gap-6 lg:grid-cols-2">
        <section className="rounded-xl border border-slate-800 bg-slate-900/40 p-4">
          <div className="mb-4 flex items-center justify-between gap-2">
            <h2 className="text-sm font-medium text-slate-300">Суммы оплаты по месяцам</h2>
            {currencies.length > 1 ? (
              <select
                className="rounded-lg border border-slate-700 bg-slate-950 px-2 py-1 text-sm"
                value={selected}
                onChange={(e) => setCurrency(e.target.value)}
              >
                {currencies.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            ) : null}
          </div>
          <p className="mb-3 text-xs text-slate-500">Сумма оплат записей, у которых срок в этом месяце. Валюты не смешиваем.</p>
          <div className="h-64">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={bars}>
                <CartesianGrid stroke="#1e293b" vertical={false} />
                <XAxis dataKey="month" stroke="#94a3b8" fontSize={12} />
                <YAxis stroke="#94a3b8" fontSize={12} allowDecimals={false} />
                <Tooltip
                  contentStyle={{ background: '#0f172a', border: '1px solid #334155' }}
                  formatter={(value) => [formatMoney(Number(value), selected), selected]}
                />
                <Bar dataKey="amount" name="Сумма" fill="#14b8a6" radius={[6, 6, 0, 0]} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </section>

        <section className="rounded-xl border border-slate-800 bg-slate-900/40 p-4">
          <div className="mb-4 flex items-center justify-between gap-2">
            <h2 className="text-sm font-medium text-slate-300">Расход по типу</h2>
            {currencies.length > 1 ? (
              <select
                className="rounded-lg border border-slate-700 bg-slate-950 px-2 py-1 text-sm"
                value={selected}
                onChange={(e) => setCurrency(e.target.value)}
              >
                {currencies.map((c) => (
                  <option key={c} value={c}>
                    {c}
                  </option>
                ))}
              </select>
            ) : null}
          </div>
          {pie.length === 0 ? (
            <p className="py-16 text-center text-sm text-slate-500">Нет сумм в {selected}</p>
          ) : (
            <div className="h-64">
              <ResponsiveContainer width="100%" height="100%">
                <PieChart>
                  <Pie data={pie} dataKey="value" nameKey="name" innerRadius={50} outerRadius={80} paddingAngle={2}>
                    {pie.map((slice) => (
                      <Cell key={slice.name} fill={slice.color} />
                    ))}
                  </Pie>
                  <Tooltip contentStyle={{ background: '#0f172a', border: '1px solid #334155' }} />
                </PieChart>
              </ResponsiveContainer>
            </div>
          )}
        </section>
      </div>

      <section className="mt-6">
        <h2 className="mb-3 text-sm font-medium text-slate-300">Ближайшие 10</h2>
        {data.soonest.length === 0 ? (
          <PageState title="Нет открытых записей" hint="Список пуст — это ваши данные, не общий каталог." />
        ) : (
          <div className="overflow-x-auto rounded-xl border border-slate-800">
            <table className="w-full min-w-[480px] text-left text-sm">
              <thead className="bg-slate-900 text-xs tracking-wide text-slate-400 uppercase">
                <tr>
                  <th className="px-3 py-2 font-medium">Запись</th>
                  <th className="px-3 py-2 font-medium">Тип</th>
                  <th className="px-3 py-2 font-medium">Срок оплаты</th>
                  <th className="px-3 py-2 font-medium">Статус</th>
                </tr>
              </thead>
              <tbody>
                {data.soonest.map((row) => (
                  <tr key={row.id} className="border-t border-slate-800">
                    <td className="px-3 py-2">
                      <Link to={`/items/${row.id}`} className="text-teal-300 hover:underline">
                        {row.title}
                      </Link>
                    </td>
                    <td className="px-3 py-2 text-slate-400">{kindName.get(row.kind_id)?.name ?? '—'}</td>
                    <td className="px-3 py-2">{formatDate(row.expires_at)}</td>
                    <td className="px-3 py-2">
                      <StatusBadge status={row.status} />
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  )
}

function collectCurrencies(data: Dashboard): string[] {
  const set = new Set<string>()
  for (const row of data.upcoming_cost) {
    set.add(row.currency)
  }
  for (const row of data.cost_by_kind) {
    set.add(row.currency)
  }
  for (const row of data.expirations_by_month) {
    for (const a of row.amounts) {
      set.add(a.currency)
    }
  }
  return [...set].sort()
}
