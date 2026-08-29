import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'

import { ApiError } from '../api/client.ts'
import { createCategory, deleteCategory, listCategories, patchCategory } from '../api/endpoints.ts'
import type { Category } from '../api/types.ts'
import { Button, ErrorBanner, Field, PageState, PageTitle, TextInput } from '../components/ui.tsx'
import { useAuth } from '../hooks/useAuth.ts'

export function CategoriesPage() {
  const { isAdmin } = useAuth()
  const qc = useQueryClient()
  const [error, setError] = useState<string | null>(null)
  const [name, setName] = useState('')
  const [parentId, setParentId] = useState('')

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

      {cats.data.items.length === 0 ? <PageState title="Категорий нет" /> : null}
      <ul className="space-y-2">
        {cats.data.items.map((node) => (
          <TreeNode
            key={node.id}
            node={node}
            depth={0}
            isAdmin={isAdmin}
            onError={setError}
            onChanged={invalidate}
          />
        ))}
      </ul>
    </div>
  )
}

function walk(nodes: Category[], depth = 0): { id: string; name: string; depth: number }[] {
  return nodes.flatMap((n) => [{ id: n.id, name: n.name, depth }, ...walk(n.children ?? [], depth + 1)])
}

function TreeNode({
  node,
  depth,
  isAdmin,
  onError,
  onChanged,
}: {
  node: Category
  depth: number
  isAdmin: boolean
  onError: (msg: string | null) => void
  onChanged: () => Promise<void>
}) {
  const [editing, setEditing] = useState(false)
  const [name, setName] = useState(node.name)

  const save = useMutation({
    mutationFn: () => patchCategory(node.id, { name }),
    onSuccess: async () => {
      setEditing(false)
      onError(null)
      await onChanged()
    },
    onError: (err) => onError(err instanceof ApiError ? err.message : 'Не удалось сохранить'),
  })

  const remove = useMutation({
    mutationFn: () => deleteCategory(node.id),
    onSuccess: async () => {
      onError(null)
      await onChanged()
    },
    onError: (err) => onError(err instanceof ApiError ? err.message : 'Нельзя удалить (дети или записи)'),
  })

  return (
    <li>
      <div className="flex flex-wrap items-center gap-2 rounded-lg bg-slate-900/50 px-3 py-2" style={{ marginLeft: depth * 16 }}>
        {editing ? (
          <>
            <TextInput value={name} onChange={(e) => setName(e.target.value)} />
            <Button type="button" onClick={() => save.mutate()}>
              Ок
            </Button>
            <Button type="button" variant="ghost" onClick={() => setEditing(false)}>
              Отмена
            </Button>
          </>
        ) : (
          <>
            <span className="font-medium">{node.name}</span>
            {isAdmin ? (
              <>
                <Button type="button" variant="ghost" onClick={() => setEditing(true)}>
                  Переименовать
                </Button>
                <Button
                  type="button"
                  variant="danger"
                  onClick={() => {
                    if (window.confirm(`Удалить «${node.name}»?`)) {
                      remove.mutate()
                    }
                  }}
                >
                  Удалить
                </Button>
              </>
            ) : null}
          </>
        )}
      </div>
      {node.children.length > 0 ? (
        <ul className="mt-2 space-y-2">
          {node.children.map((child) => (
            <TreeNode
              key={child.id}
              node={child}
              depth={depth + 1}
              isAdmin={isAdmin}
              onError={onError}
              onChanged={onChanged}
            />
          ))}
        </ul>
      ) : null}
    </li>
  )
}
