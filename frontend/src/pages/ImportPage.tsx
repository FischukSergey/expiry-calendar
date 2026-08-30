import { useMemo, useState } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'

import { ApiError } from '../api/client.ts'
import { importItems, listKinds } from '../api/endpoints.ts'
import type { CSVImportPreview, CSVImportResult } from '../api/types.ts'
import { Button, ErrorBanner, Field, PageState, PageTitle, Select } from '../components/ui.tsx'
import { csvFields, parseCSVHeaders } from '../lib/format.ts'

function isPreview(v: CSVImportPreview | CSVImportResult): v is CSVImportPreview {
  return 'rows' in v && 'valid' in v
}

export function ImportPage() {
  const qc = useQueryClient()
  const kinds = useQuery({ queryKey: ['kinds'], queryFn: listKinds })
  const [file, setFile] = useState<File | null>(null)
  const [headers, setHeaders] = useState<string[]>([])
  const [mapping, setMapping] = useState<Record<string, string>>({})
  const [attrKey, setAttrKey] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [preview, setPreview] = useState<CSVImportPreview | null>(null)
  const [created, setCreated] = useState<number | null>(null)
  const [busy, setBusy] = useState(false)

  const attrOptions = useMemo(() => {
    const keys = new Set<string>()
    for (const k of kinds.data?.items ?? []) {
      for (const f of k.attr_schema) {
        keys.add(f.key)
      }
    }
    return [...keys]
  }, [kinds.data])

  const onFile = async (next: File | null) => {
    setFile(next)
    setPreview(null)
    setCreated(null)
    if (!next) {
      setHeaders([])
      return
    }
    const text = await next.text()
    const cols = parseCSVHeaders(text)
    setHeaders(cols)
    const auto: Record<string, string> = {}
    for (const field of csvFields) {
      if (cols.includes(field.key)) {
        auto[field.key] = field.key
      }
    }
    for (const col of cols) {
      if (col.startsWith('attrs.')) {
        auto[col] = col
      }
    }
    setMapping(auto)
  }

  const setMap = (field: string, col: string) => {
    setMapping((prev) => {
      const next = { ...prev }
      if (!col) {
        delete next[field]
      } else {
        next[field] = col
      }
      return next
    })
  }

  const run = async (dry: boolean) => {
    if (!file) {
      setError('Выберите файл')
      return
    }
    setError(null)
    setBusy(true)
    try {
      const res = await importItems(file, mapping, dry)
      if (dry && isPreview(res)) {
        setPreview(res)
        setCreated(null)
      } else if (!dry && 'created' in res) {
        setCreated(res.created)
        await qc.invalidateQueries({ queryKey: ['items'] })
        await qc.invalidateQueries({ queryKey: ['dashboard'] })
        await qc.invalidateQueries({ queryKey: ['audit'] })
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Импорт не удался')
    } finally {
      setBusy(false)
    }
  }

  return (
    <div>
      <PageTitle title="Импорт CSV" subtitle="Маппинг колонок, сначала dry_run, затем запись одной пачкой" />
      {error ? <ErrorBanner message={error} /> : null}

      <div className="space-y-4 rounded-xl border border-slate-800 p-4">
        <Field label="Файл">
          <input
            type="file"
            accept=".csv,text/csv"
            className="text-sm"
            onChange={(e) => void onFile(e.target.files?.[0] ?? null)}
          />
        </Field>

        {!file ? <PageState title="Файл не выбран" hint="CSV с заголовком в первой строке" /> : null}

        {headers.length > 0 ? (
          <div className="grid gap-3 sm:grid-cols-2">
            {csvFields.map((f) => (
              <Field key={f.key} label={f.label}>
                <Select value={mapping[f.key] ?? ''} onChange={(e) => setMap(f.key, e.target.value)}>
                  <option value="">—</option>
                  {headers.map((h) => (
                    <option key={h} value={h}>
                      {h}
                    </option>
                  ))}
                </Select>
              </Field>
            ))}
          </div>
        ) : null}

        {headers.length > 0 ? (
          <div className="flex flex-wrap items-end gap-2">
            <Field label="attrs.*">
              <Select value={attrKey} onChange={(e) => setAttrKey(e.target.value)}>
                <option value="">ключ атрибута</option>
                {attrOptions.map((k) => (
                  <option key={k} value={k}>
                    {k}
                  </option>
                ))}
              </Select>
            </Field>
            <Field label="Колонка CSV">
              <Select
                value={attrKey ? (mapping[`attrs.${attrKey}`] ?? '') : ''}
                onChange={(e) => attrKey && setMap(`attrs.${attrKey}`, e.target.value)}
                disabled={!attrKey}
              >
                <option value="">—</option>
                {headers.map((h) => (
                  <option key={h} value={h}>
                    {h}
                  </option>
                ))}
              </Select>
            </Field>
          </div>
        ) : null}

        <div className="flex gap-2">
          <Button type="button" variant="outline" disabled={!file || busy} onClick={() => void run(true)}>
            Проверка
          </Button>
          <Button type="button" disabled={!file || busy} onClick={() => void run(false)}>
            Записать
          </Button>
        </div>
      </div>

      {preview ? (
        <section className="mt-6 rounded-xl border border-slate-800 p-4 text-sm">
          <p>
            Строк: {preview.rows}, валидных: {preview.valid}, ошибок: {preview.errors.length}
          </p>
          {preview.errors.length > 0 ? (
            <ul className="mt-3 space-y-1 text-rose-300">
              {preview.errors.map((e) => (
                <li key={`${e.line}-${e.message}`}>
                  строка {e.line}: {e.message}
                </li>
              ))}
            </ul>
          ) : null}
          {preview.preview.length > 0 ? (
            <ul className="mt-3 space-y-1 text-slate-300">
              {preview.preview.map((row, i) => (
                <li key={`${row.title}-${i}`}>
                  {row.title} · {row.kind_slug} · {row.expires_at}
                </li>
              ))}
            </ul>
          ) : null}
        </section>
      ) : null}

      {created !== null ? <p className="mt-4 text-teal-300">Создано записей: {created}</p> : null}
    </div>
  )
}
