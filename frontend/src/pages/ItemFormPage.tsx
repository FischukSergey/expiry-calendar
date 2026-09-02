import { useEffect, useMemo, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

import { ApiError } from '../api/client.ts'
import { createItem, getItem, listCategories, listKinds, patchItem } from '../api/endpoints.ts'
import type { AttrField, BillingPeriod, ItemWrite } from '../api/types.ts'
import { Button, ErrorBanner, Field, PageState, PageTitle, Select, TextArea, TextInput } from '../components/ui.tsx'
import { flattenCategories, suggestCategoryId } from '../lib/format.ts'

const schema = z.object({
  title: z.string().min(1, 'Название обязательно'),
  kind_id: z.string().min(1, 'Выберите тип'),
  expires_at: z.string().min(1, 'Укажите срок оплаты'),
  description: z.string(),
  category_id: z.string(),
  vendor: z.string(),
  tags: z.string(),
  cost_amount: z.number({ error: 'Сумма — целое' }).int().min(0),
  currency: z.string().length(3),
  billing_period: z.enum(['one_time', 'monthly', 'yearly']),
  started_at: z.string(),
  notify_off: z.boolean(),
  notify_before_days: z.number({ error: 'Дни — целое' }).int().min(0),
  url: z.string(),
  account_hint: z.string(),
  status: z.enum(['', 'cancelled', 'archived', 'paid']),
})

type FormValues = z.infer<typeof schema>

function emptyAttrs(fields: AttrField[], current: Record<string, unknown>): Record<string, string> {
  const out: Record<string, string> = {}
  for (const f of fields) {
    const v = current[f.key]
    if (v === undefined || v === null) {
      out[f.key] = f.type === 'boolean' ? 'false' : ''
    } else if (typeof v === 'boolean') {
      out[f.key] = v ? 'true' : 'false'
    } else {
      out[f.key] = String(v)
    }
  }
  return out
}

function toAttrs(fields: AttrField[], raw: Record<string, string>): Record<string, unknown> {
  const attrs: Record<string, unknown> = {}
  for (const f of fields) {
    const v = raw[f.key] ?? ''
    if (f.type === 'boolean') {
      attrs[f.key] = v === 'true'
      continue
    }
    if (v === '') {
      continue
    }
    if (f.type === 'number') {
      attrs[f.key] = Number(v)
      continue
    }
    attrs[f.key] = v
  }
  return attrs
}

export function ItemFormPage() {
  const { id } = useParams()
  const isEdit = Boolean(id)
  const navigate = useNavigate()
  const [formError, setFormError] = useState<string | null>(null)
  const [attrs, setAttrs] = useState<Record<string, string>>({})

  const kinds = useQuery({ queryKey: ['kinds'], queryFn: listKinds })
  const cats = useQuery({ queryKey: ['categories'], queryFn: listCategories })
  const card = useQuery({
    queryKey: ['item', id],
    queryFn: () => getItem(id!),
    enabled: isEdit && Boolean(id),
  })

  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: {
      title: '',
      kind_id: '',
      expires_at: '',
      description: '',
      category_id: '',
      vendor: '',
      tags: '',
      cost_amount: 0,
      currency: 'RUB',
      billing_period: 'one_time',
      started_at: '',
      notify_off: false,
      notify_before_days: 30,
      url: '',
      account_hint: '',
      status: '',
    },
  })

  const kindId = form.watch('kind_id')
  const notifyOff = form.watch('notify_off')
  const schemaFields = useMemo(
    () => kinds.data?.items.find((k) => k.id === kindId)?.attr_schema ?? [],
    [kinds.data, kindId],
  )
  const flatCats = flattenCategories(cats.data?.items ?? [])

  useEffect(() => {
    if (isEdit || !kindId) {
      return
    }
    const slug = kinds.data?.items.find((k) => k.id === kindId)?.slug
    if (!slug) {
      return
    }
    form.setValue('category_id', suggestCategoryId(slug, flattenCategories(cats.data?.items ?? [])))
  }, [cats.data, form, isEdit, kindId, kinds.data])

  useEffect(() => {
    if (!card.data) {
      return
    }
    const it = card.data.item
    form.reset({
      title: it.title,
      kind_id: it.kind_id,
      expires_at: it.expires_at,
      description: it.description,
      category_id: it.category_id ?? '',
      vendor: it.vendor,
      tags: it.tags.join(', '),
      cost_amount: it.cost_amount,
      currency: it.currency,
      billing_period: it.billing_period,
      started_at: it.started_at ?? '',
      notify_off: it.notify_before_days === null,
      notify_before_days: it.notify_before_days ?? 30,
      url: it.url,
      account_hint: it.account_hint,
      status: it.status === 'cancelled' || it.status === 'archived' || it.status === 'paid' ? it.status : '',
    })
    setAttrs(emptyAttrs(it.kind_id ? (kinds.data?.items.find((k) => k.id === it.kind_id)?.attr_schema ?? []) : [], it.attrs))
  }, [card.data, form, kinds.data])

  useEffect(() => {
    setAttrs((prev) => emptyAttrs(schemaFields, Object.fromEntries(Object.entries(prev))))
  }, [schemaFields])

  const onSubmit = form.handleSubmit(async (values) => {
    setFormError(null)
    const tags = values.tags
      .split(',')
      .map((tag: string) => tag.trim())
      .filter(Boolean)
    const body: ItemWrite = {
      title: values.title,
      kind_id: values.kind_id,
      expires_at: values.expires_at,
      description: values.description,
      vendor: values.vendor,
      tags,
      cost_amount: values.cost_amount,
      currency: values.currency,
      billing_period: values.billing_period as BillingPeriod,
      notify_before_days: values.notify_off ? null : values.notify_before_days,
      url: values.url,
      account_hint: values.account_hint,
      category_id: values.category_id || null,
      attrs: toAttrs(schemaFields, attrs),
    }
    if (values.started_at) {
      body.started_at = values.started_at
    }
    if (values.status) {
      body.status = values.status
    } else if (isEdit) {
      body.status = 'active'
    }
    try {
      if (isEdit && id) {
        await patchItem(id, body)
        navigate(`/items/${id}`)
      } else {
        const created = await createItem(body)
        navigate(`/items/${created.id}`)
      }
    } catch (err) {
      setFormError(err instanceof ApiError ? err.message : 'Не удалось сохранить')
    }
  })

  if (isEdit && card.isPending) {
    return <PageState title="Загрузка формы…" />
  }
  if (isEdit && card.isError) {
    return <PageState title="Запись не найдена" hint={card.error.message} onRetry={() => void card.refetch()} />
  }

  return (
    <div>
      <PageTitle title={isEdit ? 'Редактирование' : 'Новая запись'} subtitle="Поля со звёздочкой обязательны" />
      <form onSubmit={onSubmit} className="space-y-6">
        {formError ? <ErrorBanner message={formError} /> : null}
        <div className="grid gap-4 lg:grid-cols-2">
          <Field label="Название" required error={form.formState.errors.title?.message}>
            <TextInput {...form.register('title')} />
          </Field>
          <Field
            label="Тип записи"
            required
            hint="Что это за платёж: домен, подписка, налог. От типа зависят доп. поля."
            error={form.formState.errors.kind_id?.message}
          >
            <Select {...form.register('kind_id')}>
              <option value="">Выберите</option>
              {(kinds.data?.items ?? []).map((k) => (
                <option key={k.id} value={k.id}>
                  {k.name}
                </option>
              ))}
            </Select>
          </Field>
          <Field label="Срок оплаты" required error={form.formState.errors.expires_at?.message}>
            <TextInput type="date" {...form.register('expires_at')} />
          </Field>
          <Field label="Начало периода" hint="С какого дня идёт текущий период (покупка, договор). Необязательно.">
            <TextInput type="date" {...form.register('started_at')} />
          </Field>
          <Field
            label="Раздел"
            hint="Папка в дереве «Категории». Подставляется по типу, можно сменить."
          >
            <Select {...form.register('category_id')}>
              <option value="">Без раздела</option>
              {flatCats.map((c) => (
                <option key={c.id} value={c.id}>
                  {'· '.repeat(c.depth)}
                  {c.name}
                </option>
              ))}
            </Select>
          </Field>
          <Field label="Поставщик">
            <TextInput {...form.register('vendor')} />
          </Field>
          <Field label="Сумма">
            <TextInput type="number" min={0} step={1} {...form.register('cost_amount', { valueAsNumber: true })} />
          </Field>
          <Field label="Валюта">
            <Select {...form.register('currency')}>
              <option value="RUB">RUB</option>
              <option value="USD">USD</option>
              <option value="EUR">EUR</option>
            </Select>
          </Field>
          <Field label="Период">
            <Select {...form.register('billing_period')}>
              <option value="one_time">Разово</option>
              <option value="monthly">Ежемесячно</option>
              <option value="yearly">Ежегодно</option>
            </Select>
          </Field>
          <Field
            label="Напомнить за, дней"
            hint="Ноль — в день срока. «Не уведомлять» выключает колокольчик и окно «скоро срок»."
          >
            <label className="mb-2 flex items-center gap-2 text-sm text-slate-200">
              <input type="checkbox" className="size-4 accent-teal-500" {...form.register('notify_off')} />
              Не уведомлять
            </label>
            <div className="flex gap-2">
              <TextInput
                type="number"
                min={0}
                disabled={notifyOff}
                {...form.register('notify_before_days', { valueAsNumber: true })}
              />
              {[7, 14, 30].map((n) => (
                <Button
                  key={n}
                  type="button"
                  variant="outline"
                  disabled={notifyOff}
                  onClick={() => form.setValue('notify_before_days', n)}
                >
                  {n}
                </Button>
              ))}
            </div>
          </Field>
          <Field label="Теги">
            <TextInput placeholder="через запятую" {...form.register('tags')} />
          </Field>
          <Field label="Статус вручную">
            <Select {...form.register('status')}>
              <option value="">Считает сервер</option>
              <option value="paid">Оплачено</option>
              <option value="cancelled">Отменено</option>
              <option value="archived">Архив</option>
            </Select>
          </Field>
          <Field label="URL">
            <TextInput {...form.register('url')} />
          </Field>
          <Field label="Подсказка аккаунта">
            <TextInput {...form.register('account_hint')} />
          </Field>
        </div>
        <Field label="Описание">
          <TextArea {...form.register('description')} />
        </Field>

        {schemaFields.length > 0 ? (
          <section className="rounded-xl border border-slate-800 p-4">
            <h2 className="mb-3 text-sm font-medium text-slate-300">Поля типа</h2>
            <div className="grid gap-4 lg:grid-cols-2">
              {schemaFields.map((f) => (
                <Field key={f.key} label={f.label} required={f.required}>
                  {f.type === 'boolean' ? (
                    <Select
                      value={attrs[f.key] ?? 'false'}
                      onChange={(e) => setAttrs((prev) => ({ ...prev, [f.key]: e.target.value }))}
                    >
                      <option value="false">Нет</option>
                      <option value="true">Да</option>
                    </Select>
                  ) : (
                    <TextInput
                      type={f.type === 'number' ? 'number' : 'text'}
                      required={f.required}
                      value={attrs[f.key] ?? ''}
                      onChange={(e) => setAttrs((prev) => ({ ...prev, [f.key]: e.target.value }))}
                    />
                  )}
                </Field>
              ))}
            </div>
          </section>
        ) : null}

        <div className="flex gap-2">
          <Button type="submit" disabled={form.formState.isSubmitting}>
            {form.formState.isSubmitting ? 'Сохранение…' : 'Сохранить'}
          </Button>
          <Button type="button" variant="ghost" onClick={() => navigate(-1)}>
            Отмена
          </Button>
        </div>
      </form>
    </div>
  )
}
