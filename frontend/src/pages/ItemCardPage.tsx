import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { ApiError } from '../api/client.ts'
import { deleteItem, getItem, listCategories, listKinds, renewItem } from '../api/endpoints.ts'
import { Button, ErrorBanner, Field, PageState, PageTitle, StatusBadge, TextInput } from '../components/ui.tsx'
import { useAuth } from '../hooks/useAuth.ts'
import { billingLabel, findCategoryName, formatDate, formatDateTime, formatMoney } from '../lib/format.ts'

export function ItemCardPage() {
  const { id = '' } = useParams()
  const { isAdmin } = useAuth()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const [renewError, setRenewError] = useState<string | null>(null)
  const [expires, setExpires] = useState('')
  const [cost, setCost] = useState('')
  const [comment, setComment] = useState('')

  const kinds = useQuery({ queryKey: ['kinds'], queryFn: listKinds })
  const cats = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const card = useQuery({ queryKey: ['item', id], queryFn: () => getItem(id), enabled: Boolean(id) })

  const remove = useMutation({
    mutationFn: () => deleteItem(id),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ['items'] })
      navigate('/items')
    },
  })

  const renew = useMutation({
    mutationFn: () =>
      renewItem(id, {
        new_expires_at: expires,
        new_cost: cost === '' ? undefined : Number(cost),
        comment: comment || undefined,
      }),
    onSuccess: async () => {
      setRenewError(null)
      setComment('')
      await qc.invalidateQueries({ queryKey: ['item', id] })
      await qc.invalidateQueries({ queryKey: ['dashboard'] })
    },
    onError: (err) => {
      setRenewError(err instanceof ApiError ? err.message : 'Не удалось продлить')
    },
  })

  if (card.isPending) {
    return <PageState title="Загрузка карточки…" />
  }
  if (card.isError) {
    return <PageState title="Запись не найдена" hint={card.error.message} onRetry={() => void card.refetch()} />
  }

  const it = card.data.item
  const kind = kinds.data?.items.find((k) => k.id === it.kind_id)
  const rows: { label: string; value: string }[] = [
    { label: 'Тип', value: kind?.name ?? '—' },
    { label: 'Категория', value: findCategoryName(cats.data?.items ?? [], it.category_id) },
    { label: 'Поставщик', value: it.vendor || '—' },
    { label: 'Сумма', value: formatMoney(it.cost_amount, it.currency) },
    { label: 'Период', value: billingLabel[it.billing_period] },
    { label: 'Начало', value: formatDate(it.started_at ?? '') },
    { label: 'Истекает', value: formatDate(it.expires_at) },
    { label: 'Напомнить за', value: `${it.notify_before_days} дн.` },
    { label: 'Теги', value: it.tags.length ? it.tags.join(', ') : '—' },
    { label: 'URL', value: it.url || '—' },
    { label: 'Подсказка', value: it.account_hint || '—' },
  ]

  return (
    <div>
      <PageTitle
        title={it.title}
        subtitle={it.description || undefined}
        actions={
          isAdmin ? (
            <>
              <Link to={`/items/${it.id}/edit`}>
                <Button type="button" variant="outline">
                  Изменить
                </Button>
              </Link>
              <Button
                type="button"
                variant="danger"
                onClick={() => {
                  if (window.confirm('Удалить запись?')) {
                    remove.mutate()
                  }
                }}
              >
                Удалить
              </Button>
            </>
          ) : undefined
        }
      />

      <div className="mb-4">
        <StatusBadge status={it.status} />
      </div>

      <dl className="grid gap-3 rounded-xl border border-slate-800 bg-slate-900/40 p-4 sm:grid-cols-2">
        {rows.map((row) => (
          <div key={row.label}>
            <dt className="text-xs tracking-wide text-slate-500 uppercase">{row.label}</dt>
            <dd className="mt-1 break-all text-sm text-slate-100">{row.value}</dd>
          </div>
        ))}
      </dl>

      {kind && kind.attr_schema.length > 0 ? (
        <section className="mt-6 rounded-xl border border-slate-800 p-4">
          <h2 className="mb-3 text-sm font-medium text-slate-300">Поля типа</h2>
          <dl className="grid gap-3 sm:grid-cols-2">
            {kind.attr_schema.map((f) => (
              <div key={f.key}>
                <dt className="text-xs text-slate-500 uppercase">{f.label}</dt>
                <dd className="mt-1 text-sm">{formatAttr(it.attrs[f.key])}</dd>
              </div>
            ))}
          </dl>
        </section>
      ) : null}

      {isAdmin ? (
        <section className="mt-6 rounded-xl border border-slate-800 p-4">
          <h2 className="mb-3 text-sm font-medium text-slate-300">Продление</h2>
          {renewError ? <ErrorBanner message={renewError} /> : null}
          <div className="mt-3 grid gap-3 sm:grid-cols-3">
            <Field label="Новая дата">
              <TextInput type="date" value={expires} onChange={(e) => setExpires(e.target.value)} />
            </Field>
            <Field label="Новая сумма">
              <TextInput type="number" min={0} value={cost} onChange={(e) => setCost(e.target.value)} placeholder="не менять" />
            </Field>
            <Field label="Комментарий">
              <TextInput value={comment} onChange={(e) => setComment(e.target.value)} />
            </Field>
          </div>
          <Button className="mt-4" type="button" disabled={!expires || renew.isPending} onClick={() => renew.mutate()}>
            Продлить
          </Button>
        </section>
      ) : null}

      <section className="mt-6">
        <h2 className="mb-3 text-sm font-medium text-slate-300">История продлений</h2>
        {card.data.renewals.length === 0 ? (
          <PageState title="Пока нет продлений" />
        ) : (
          <div className="overflow-x-auto rounded-xl border border-slate-800">
            <table className="w-full text-left text-sm">
              <thead className="bg-slate-900 text-xs text-slate-400 uppercase">
                <tr>
                  <th className="px-3 py-2 font-medium">Когда</th>
                  <th className="px-3 py-2 font-medium">Было</th>
                  <th className="px-3 py-2 font-medium">Стало</th>
                  <th className="px-3 py-2 font-medium">Цена</th>
                  <th className="px-3 py-2 font-medium">Комментарий</th>
                </tr>
              </thead>
              <tbody>
                {card.data.renewals.map((r) => (
                  <tr key={r.id} className="border-t border-slate-800">
                    <td className="px-3 py-2">{formatDateTime(r.created_at)}</td>
                    <td className="px-3 py-2">{formatDate(r.old_expires_at)}</td>
                    <td className="px-3 py-2">{formatDate(r.new_expires_at)}</td>
                    <td className="px-3 py-2">
                      {r.old_cost} → {r.new_cost}
                    </td>
                    <td className="px-3 py-2 text-slate-400">{r.comment || '—'}</td>
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

function formatAttr(value: unknown): string {
  if (value === undefined || value === null || value === '') {
    return '—'
  }
  if (typeof value === 'boolean') {
    return value ? 'да' : 'нет'
  }
  return String(value)
}
