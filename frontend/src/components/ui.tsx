import { forwardRef, type ButtonHTMLAttributes, type InputHTMLAttributes, type ReactNode, type SelectHTMLAttributes, type TextareaHTMLAttributes } from 'react'
import { Link } from 'react-router-dom'

import type { OccurrenceStatus, Status } from '../api/types.ts'
import { occurrenceLabel, statusLabel } from '../lib/format.ts'

export const fieldClass =
  'w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-100 placeholder:text-slate-500 outline-none focus:border-teal-500'

const btnBase = 'inline-flex items-center justify-center gap-2 rounded-lg px-3 py-2 text-sm font-medium transition disabled:opacity-50'

export function Button({
  variant = 'primary',
  className = '',
  ...props
}: ButtonHTMLAttributes<HTMLButtonElement> & { variant?: 'primary' | 'ghost' | 'danger' | 'outline' }) {
  const styles = {
    primary: 'bg-teal-500 text-slate-950 hover:bg-teal-400',
    ghost: 'text-slate-200 hover:bg-slate-800',
    danger: 'bg-rose-600 text-white hover:bg-rose-500',
    outline: 'border border-slate-600 text-slate-100 hover:bg-slate-800',
  }
  return <button className={`${btnBase} ${styles[variant]} ${className}`} {...props} />
}

export const TextInput = forwardRef<HTMLInputElement, InputHTMLAttributes<HTMLInputElement>>(
  function TextInput(props, ref) {
    return <input ref={ref} className={fieldClass} {...props} />
  },
)

export const TextArea = forwardRef<HTMLTextAreaElement, TextareaHTMLAttributes<HTMLTextAreaElement>>(
  function TextArea(props, ref) {
    return <textarea ref={ref} className={`${fieldClass} min-h-24`} {...props} />
  },
)

export const Select = forwardRef<HTMLSelectElement, SelectHTMLAttributes<HTMLSelectElement>>(
  function Select(props, ref) {
    return <select ref={ref} className={fieldClass} {...props} />
  },
)

export function Label({ children }: { children: ReactNode }) {
  return <label className="mb-1 block text-xs font-medium tracking-wide text-slate-400 uppercase">{children}</label>
}

export function Field({
  label,
  required,
  hint,
  error,
  children,
}: {
  label: string
  required?: boolean
  hint?: string
  error?: string
  children: ReactNode
}) {
  return (
    <div>
      <Label>
        {label}
        {required ? (
          <span className="ml-0.5 text-rose-400" aria-hidden="true">
            *
          </span>
        ) : null}
      </Label>
      {children}
      {error ? <p className="mt-1 text-xs text-rose-300">{error}</p> : null}
      {!error && hint ? <p className="mt-1 text-xs text-slate-500">{hint}</p> : null}
    </div>
  )
}

const statusClass: Record<Status, string> = {
  active: 'bg-emerald-500/15 text-emerald-300',
  expiring: 'bg-amber-500/15 text-amber-300',
  expired: 'bg-rose-500/15 text-rose-300',
  cancelled: 'bg-slate-500/20 text-slate-300',
  archived: 'bg-zinc-500/20 text-zinc-300',
  paid: 'bg-sky-500/15 text-sky-300',
}

export function StatusBadge({ status }: { status: Status }) {
  return (
    <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${statusClass[status]}`}>
      {statusLabel[status]}
    </span>
  )
}

const occurrenceClass: Record<OccurrenceStatus, string> = {
  open: 'bg-amber-500/15 text-amber-300',
  paid: 'bg-sky-500/15 text-sky-300',
}

export function OccurrenceBadge({ status }: { status: OccurrenceStatus }) {
  return (
    <span className={`inline-flex rounded-full px-2 py-0.5 text-xs font-medium ${occurrenceClass[status]}`}>
      {occurrenceLabel[status]}
    </span>
  )
}

export function PageState({
  title,
  hint,
  action,
  onRetry,
}: {
  title: string
  hint?: string
  action?: ReactNode
  onRetry?: () => void
}) {
  return (
    <div className="rounded-xl border border-dashed border-slate-700 px-4 py-12 text-center sm:px-6 sm:py-16">
      <p className="text-lg font-medium text-slate-100">{title}</p>
      {hint ? <p className="mt-2 text-sm text-slate-400">{hint}</p> : null}
      {onRetry ? (
        <div className="mt-4">
          <Button type="button" variant="outline" onClick={onRetry}>
            Повторить
          </Button>
        </div>
      ) : null}
      {action ? <div className="mt-4">{action}</div> : null}
    </div>
  )
}

export function ErrorBanner({ message }: { message: string }) {
  return (
    <div className="rounded-lg border border-rose-800 bg-rose-950/50 px-3 py-2 text-sm text-rose-200">{message}</div>
  )
}

export function PageTitle({
  title,
  subtitle,
  actions,
}: {
  title: string
  subtitle?: string
  actions?: ReactNode
}) {
  return (
    <div className="mb-6 flex flex-wrap items-start justify-between gap-3">
      <div>
        <h1 className="text-2xl font-semibold tracking-tight text-slate-50">{title}</h1>
        {subtitle ? <p className="mt-1 text-sm text-slate-400">{subtitle}</p> : null}
      </div>
      {actions ? <div className="flex flex-wrap gap-2">{actions}</div> : null}
    </div>
  )
}

export function TextLink({ to, children }: { to: string; children: ReactNode }) {
  return (
    <Link to={to} className="text-teal-400 hover:text-teal-300">
      {children}
    </Link>
  )
}
