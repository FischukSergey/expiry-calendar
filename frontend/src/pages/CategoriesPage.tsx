import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type ColumnDef } from '@tanstack/react-table'

import { ApiError } from '../api/client.ts'
import { createCategory, deleteCategory, listCategories, patchCategory } from '../api/endpoints.ts'
import type { Category } from '../api/types.ts'
import { DataTable } from '../components/DataTable.tsx'
import { Button, ErrorBanner, Field, PageState, PageTitle, TextInput } from '../components/ui.tsx'
import { useAuth } from '../hooks/useAuth.ts'

type FlatCategory = {
  id: string
  name: string
  depth: number
  parentName: string | null
}

export function CategoriesPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [parentId, setParentId] = useState('')
  const [editId, setEditId] = useState<string | null>(null)
  const [draft, setDraft] = useState('')

  const cats = useQuery({ queryKey: ['categories'], queryFn: listCategories })

  const invalidate = async () => {
    await qc.invalidateQueries({ queryKey: ['categories'] })
  }

  const add = useMutation({
    mutationFn: () => createCategory({ name, parent_id: parentId || null }),
    onSuccess: async () => {
      setName('')
      setError(null)
      await invalidate()
    },
    onError: (err) => setError(err instanceof ApiError ? err.message : 'Не удалось создать'),
  })

  const save = useMutation({
    mutationFn: (input: { id: string; name: string }) => patchCategory(input.id, { name: input.name }),
    onSuccess: async () => {
      setEditId(null)
      setError(null)
      await invalidate()
    },
    onError: (err) => setError(err instanceof ApiError ? err.message : 'Не удалось сохранить'),
  })

  const remove = useMutation({
    mutationFn: (id: string) => deleteCategory(id),
    onSuccess: async () => {
      setError(null)
      await invalidate()
    },
    onError: (err) => setError(err instanceof ApiError ? err.message : 'Нельзя удалить (дети или записи)'),
  })

  const rows = useMemo(() => flattenCategories(cats.data?.items ?? []), [cats.data])
  const columns = useMemo<ColumnDef<FlatCategory>[]>(() => {
    const defs: ColumnDef<FlatCategory>[] = [
      {
        accessorKey: 'name',
        header: 'Категория',
        cell: ({ row }) => (
          <span className="flex min-w-0 items-center gap-2" style={{ paddingLeft: `${row.original.depth * 1.25}rem` }}>
            {row.original.depth > 0 ? <span className="text-slate-600">└</span> : null}
            {editId === row.original.id ? (
              <TextInput value={draft} onChange={(e) => setDraft(e.target.value)} />
            ) : (
              <span className="font-medium">{row.original.name}</span>
            )}
          </span>
        ),
      },
      {
        accessorKey: 'parentName',
        header: 'Родитель',
        cell: ({ row }) => <span className="text-slate-400">{row.original.parentName ?? '—'}</span>,
      },
    ]
    if (isAdmin) {
      defs.push({
        id: 'actions',
        header: '',
        cell: ({ row }) =>
          editId === row.original.id ? (
            <span className="inline-flex gap-2">
              <Button type="button" disabled={save.isPending} onClick={() => save.mutate({ id: row.original.id, name: draft })}>
                Ок
              </Button>
              <Button type="button" variant="ghost" onClick={() => setEditId(null)}>
                Отмена
              </Button>
            </span>
          ) : (
            <span className="inline-flex gap-2">
              <Button
                type="button"
                variant="ghost"
                onClick={() => {
                  setEditId(row.original.id)
                  setDraft(row.original.name)
                }}
              >
                Переименовать
              </Button>
              <Button
                type="button"
                variant="danger"
                disabled={remove.isPending}
                onClick={() => {
                  if (window.confirm(`Удалить «${row.original.name}»?`)) {
                    remove.mutate(row.original.id)
                  }
                }}
              >
                Удалить
              </Button>
            </span>
          ),
        meta: { align: 'right', className: 'whitespace-nowrap' },
      })
    }
    return defs
  }, [draft, editId, isAdmin, remove, save])

  if (cats.isPending) {
    return <PageState title="Загрузка категорий…" />
  }
  if (cats.isError) {
    return <PageState title="Ошибка категорий" hint={cats.error.message} onRetry={() => void cats.refetch()} />
  }

  return (
    <div>
      <PageTitle title="Категории" subtitle="Дерево, глубина до 3" />
      {error ? <ErrorBanner message={error} /> : null}

      {isAdmin ? (
        <form
          className="mb-6 flex flex-wrap items-end gap-3 rounded-xl border border-slate-800 p-4"
          onSubmit={(e) => {
            e.preventDefault()
            if (name.trim()) {
              add.mutate()
            }
          }}
        >
          <Field label="Новая категория">
            <TextInput value={name} onChange={(e) => setName(e.target.value)} />
          </Field>
          <Field label="Родитель">
            <select
              className="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm"
              value={parentId}
              onChange={(e) => setParentId(e.target.value)}
            >
              <option value="">Корень</option>
              {walk(cats.data.items).map((c) => (
                <option key={c.id} value={c.id}>
                  {'· '.repeat(c.depth)}
                  {c.name}
                </option>
              ))}
            </select>
          </Field>
          <Button type="submit" disabled={add.isPending}>
            Добавить
          </Button>
        </form>
      ) : null}

      {rows.length === 0 ? (
        <PageState title="Категорий нет" />
      ) : (
        <div className="overflow-hidden rounded-xl border border-slate-800">
          <DataTable columns={columns} data={rows} getRowId={(row) => row.id} />
        </div>
      )}
    </div>
  )
}

function walk(nodes: Category[], depth = 0): { id: string; name: string; depth: number }[] {
  return nodes.flatMap((n) => [{ id: n.id, name: n.name, depth }, ...walk(n.children ?? [], depth + 1)])
}

function flattenCategories(nodes: Category[], depth = 0, parentName: string | null = null): FlatCategory[] {
  return nodes.flatMap((node) => [
    { id: node.id, name: node.name, depth, parentName },
    ...flattenCategories(node.children ?? [], depth + 1, node.name),
  ])
}
