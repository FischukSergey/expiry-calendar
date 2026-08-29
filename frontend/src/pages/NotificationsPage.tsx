import { Link, useSearchParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { listNotifications, readAllNotifications, readNotification } from '../api/endpoints.ts'
import { Button, PageState, PageTitle, StatusBadge } from '../components/ui.tsx'
import { formatDateTime } from '../lib/format.ts'

export function NotificationsPage() {
  const qc = useQueryClient()
  const [params, setParams] = useSearchParams()
  const unreadOnly = params.get('unread') === '1'
  const page = Number(params.get('page') ?? '1')

  const list = useQuery({
    queryKey: ['notifications', { unreadOnly, page }],
    queryFn: () => listNotifications({ unread: unreadOnly, page, per_page: 20 }),
  })

  const invalidate = async () => {
    await qc.invalidateQueries({ queryKey: ['notifications'] })
  }

  const readOne = useMutation({
    mutationFn: (id: string) => readNotification(id),
    onSuccess: invalidate,
  })
  const readAll = useMutation({
    mutationFn: readAllNotifications,
    onSuccess: invalidate,
  })

  const pages = Math.max(1, Math.ceil((list.data?.total ?? 0) / 20))

  return (
    <div>
      <PageTitle
        title="Уведомления"
        subtitle={`${list.data?.total ?? 0} в ленте`}
        actions={
          <>
            <Button
              type="button"
              variant={unreadOnly ? 'primary' : 'outline'}
              onClick={() => setParams(unreadOnly ? {} : { unread: '1' })}
            >
              Только непрочитанные
            </Button>
            <Button type="button" variant="outline" onClick={() => readAll.mutate()} disabled={readAll.isPending}>
              Прочитать все
            </Button>
          </>
        }
      />

      {list.isPending ? <PageState title="Загрузка ленты…" /> : null}
      {list.isError ? (
        <PageState title="Ошибка ленты" hint={list.error.message} onRetry={() => void list.refetch()} />
      ) : null}
      {list.data && list.data.items.length === 0 ? <PageState title="Пусто" hint="Новых событий нет" /> : null}

      <ul className="space-y-2">
        {(list.data?.items ?? []).map((n) => (
          <li
            key={n.id}
            className={`rounded-xl border px-4 py-3 ${n.read_at ? 'border-slate-800' : 'border-teal-800 bg-teal-950/20'}`}
          >
            <div className="flex flex-wrap items-start justify-between gap-2">
              <div>
                <Link to={`/items/${n.item_id}`} className="font-medium text-teal-300 hover:underline">
                  {n.title}
                </Link>
                <div className="mt-1 flex flex-wrap items-center gap-2 text-xs text-slate-400">
                  <StatusBadge status={n.to_status} />
                  <span>{formatDateTime(n.created_at)}</span>
                </div>
              </div>
              {!n.read_at ? (
                <Button type="button" variant="ghost" onClick={() => readOne.mutate(n.id)}>
                  Прочитано
                </Button>
              ) : null}
            </div>
          </li>
        ))}
      </ul>

      {pages > 1 ? (
        <div className="mt-4 flex justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            disabled={page <= 1}
            onClick={() => setParams({ ...(unreadOnly ? { unread: '1' } : {}), page: String(page - 1) })}
          >
            Назад
          </Button>
          <Button
            type="button"
            variant="outline"
            disabled={page >= pages}
            onClick={() => setParams({ ...(unreadOnly ? { unread: '1' } : {}), page: String(page + 1) })}
          >
            Дальше
          </Button>
        </div>
      ) : null}
    </div>
  )
}
