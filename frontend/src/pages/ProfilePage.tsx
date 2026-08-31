import { Link } from 'react-router-dom'

import { Button, PageTitle } from '../components/ui.tsx'
import { useAuth } from '../hooks/useAuth.ts'
import { useInstallPrompt } from '../hooks/useInstallPrompt.ts'
import { disablePush, usePush } from '../hooks/usePush.ts'

export function ProfilePage() {
  const { user, isAdmin, logout, logoutAll } = useAuth()
  const { canInstall, install } = useInstallPrompt()
  const push = usePush()

  return (
    <div>
      <PageTitle title="Профиль" subtitle="Выход, пуши и установка PWA" />
      <div className="space-y-3 rounded-xl border border-slate-800 bg-slate-900/40 p-4">
        <p className="text-sm text-slate-400">Email</p>
        <p className="text-lg">{user?.email}</p>
        {!isAdmin ? (
          <p className="mt-3 text-sm text-slate-500">Только просмотр — локальный стенд, не продуктовая роль.</p>
        ) : null}
      </div>

      <section className="mt-6 space-y-3 rounded-xl border border-slate-800 p-4">
        <h2 className="text-sm font-medium text-slate-300">Уведомления в ОС</h2>
        {!push.supported ? (
          <p className="text-sm text-slate-500">Этот браузер не умеет Web Push. Демо — Chromium.</p>
        ) : (
          <>
            <p className="text-sm text-slate-400">
              Разрешение: {push.permission === 'granted' ? 'есть' : push.permission === 'denied' ? 'запрещено' : 'не спрашивали'}
            </p>
            <div className="flex flex-wrap gap-2">
              <Button type="button" disabled={push.busy} onClick={() => void push.request()}>
                Разрешить пуши
              </Button>
              <Button type="button" variant="outline" onClick={() => void disablePush()}>
                Отписаться
              </Button>
            </div>
          </>
        )}
      </section>

      <section className="mt-6 space-y-3 rounded-xl border border-slate-800 p-4">
        <h2 className="text-sm font-medium text-slate-300">Установка</h2>
        {canInstall ? (
          <Button type="button" onClick={() => void install()}>
            Установить
          </Button>
        ) : (
          <p className="text-sm text-slate-500">
            Подсказка появится в Chrome, если приложение ещё не установлено. Или меню браузера → «Установить».
          </p>
        )}
      </section>

      <div className="mt-6 grid gap-2 lg:hidden">
        <Link className="rounded-lg border border-slate-800 px-3 py-2 text-sm" to="/categories">
          Категории
        </Link>
        {isAdmin ? (
          <>
            <Link className="rounded-lg border border-slate-800 px-3 py-2 text-sm" to="/import">
              Импорт CSV
            </Link>
            <Link className="rounded-lg border border-slate-800 px-3 py-2 text-sm" to="/audit">
              Журнал аудита
            </Link>
          </>
        ) : null}
      </div>

      <div className="mt-6 flex flex-wrap gap-2">
        <Button type="button" variant="outline" onClick={() => void logout()}>
          Выйти
        </Button>
        <Button type="button" variant="danger" onClick={() => void logoutAll()}>
          Выйти везде
        </Button>
      </div>
    </div>
  )
}
