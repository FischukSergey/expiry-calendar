import { useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'

import { listAudit } from '../api/endpoints.ts'
import { Button, PageState, PageTitle } from '../components/ui.tsx'
import { formatDateTime } from '../lib/format.ts'

export function AuditPage() {
  const [params, setParams] = useSearchParams()
  const page = Number(params.get('page') ?? '1')
  const [open, setOpen] = useState<string | null>(null)

  const list = useQuery({
    queryKey: ['audit', page],
    queryFn: () => listAudit(page, 20),
  })

  const pages = Math.max(1, Math.ceil((list.data?.total ?? 0) / 20))

  return (
    <div>
      <PageTitle title="Журнал аудита" subtitle="Только admin. Снимки без url и подсказок аккаунта." />
      {list.isPending ? <PageState title="Загрузка журнала…" /> : null}
      {list.isError ? (
        <PageState title="Ошибка журнала" hint={list.error.message} onRetry={() => void list.refetch()} />
      ) : null}
      {list.data && list.data.items.length === 0 ? <PageState title="Пока пусто" /> : null}

      <ul className="space-y-2">
        {(list.data?.items ?? []).map((row) => (
          <li key={row.id} className="rounded-xl border border-slate-800">
            <button
              type="button"
              className="flex w-full flex-wrap items-center justify-between gap-2 px-4 py-3 text-left"
              onClick={() => setOpen(open === row.id ? null : row.id)}
            >
              <span className="font-medium text-slate-100">{row.action}</span>
              <span className="text-sm text-slate-400">{row.entity}</span>
              <span className="text-sm text-slate-500">{formatDateTime(row.created_at)}</span>
            </button>
            {open === row.id ? (
              <pre className="overflow-x-auto border-t border-slate-800 p-4 text-xs text-slate-300">
                {JSON.stringify({ before: row.before_json, after: row.after_json, entity_id: row.entity_id }, null, 2)}
              </pre>
            ) : null}
          </li>
        ))}
      </ul>

      {pages > 1 ? (
        <div className="mt-4 flex justify-end gap-2">
          <Button type="button" variant="outline" disabled={page <= 1} onClick={() => setParams({ page: String(page - 1) })}>
            Назад
          </Button>
          <Button type="button" variant="outline" disabled={page >= pages} onClick={() => setParams({ page: String(page + 1) })}>
            Дальше
          </Button>
        </div>
      ) : null}
    </div>
  )
}
