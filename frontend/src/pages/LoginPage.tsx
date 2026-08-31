import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'

import { ApiError } from '../api/client.ts'
import { Button, ErrorBanner, Field, TextInput } from '../components/ui.tsx'
import { useAuth } from '../hooks/useAuth.ts'

const schema = z.object({
  email: z.string().email('Нужен email'),
  password: z.string().min(8, 'Минимум 8 символов'),
})

type FormValues = z.infer<typeof schema>

export function LoginPage() {
  const { login, register } = useAuth()
  const [mode, setMode] = useState<'login' | 'register'>('login')
  const [error, setError] = useState<string | null>(null)
  const form = useForm<FormValues>({
    resolver: zodResolver(schema),
    defaultValues: { email: '', password: '' },
  })

  const onSubmit = form.handleSubmit(async (values) => {
    setError(null)
    try {
      if (mode === 'login') {
        await login(values.email, values.password)
      } else {
        await register(values.email, values.password)
      }
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Не удалось войти')
    }
  })

  return (
    <main className="flex min-h-screen items-center justify-center bg-slate-950 px-6 text-slate-50">
      <div className="w-full max-w-md">
        <p className="text-sm tracking-[0.2em] text-slate-400 uppercase">календарь сроков оплаты</p>
        <h1 className="mt-3 text-4xl font-semibold tracking-tight">Duekeep</h1>
        <p className="mt-2 text-slate-400">
          {mode === 'login' ? 'Вход в аккаунт' : 'Регистрация — свой список, без общего каталога'}
        </p>

        <form onSubmit={onSubmit} className="mt-8 space-y-4 rounded-2xl border border-slate-800 bg-slate-900/60 p-6">
          {error ? <ErrorBanner message={error} /> : null}
          <Field label="Email">
            <TextInput type="email" autoComplete="email" {...form.register('email')} />
            {form.formState.errors.email ? (
              <p className="mt-1 text-xs text-rose-300">{form.formState.errors.email.message}</p>
            ) : null}
          </Field>
          <Field label="Пароль">
            <TextInput type="password" autoComplete={mode === 'login' ? 'current-password' : 'new-password'} {...form.register('password')} />
            {form.formState.errors.password ? (
              <p className="mt-1 text-xs text-rose-300">{form.formState.errors.password.message}</p>
            ) : null}
          </Field>
          <Button type="submit" className="w-full" disabled={form.formState.isSubmitting}>
            {form.formState.isSubmitting ? 'Секунду…' : mode === 'login' ? 'Войти' : 'Зарегистрироваться'}
          </Button>
          <button
            type="button"
            className="w-full text-center text-sm text-slate-400 hover:text-slate-200"
            onClick={() => {
              setError(null)
              setMode(mode === 'login' ? 'register' : 'login')
            }}
          >
            {mode === 'login' ? 'Нет аккаунта — регистрация' : 'Уже есть аккаунт — вход'}
          </button>
        </form>
        <p className="mt-4 text-center text-xs text-slate-500">
          Локальный стенд: admin@duekeep.local / admin1234
        </p>
      </div>
    </main>
  )
}
