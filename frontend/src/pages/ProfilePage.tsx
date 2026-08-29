import { Link } from 'react-router-dom'

import { Button, PageTitle } from '../components/ui.tsx'
import { useAuth } from '../hooks/useAuth.ts'

export function ProfilePage() {
  const { user, isAdmin, logout, logoutAll } = useAuth()

  return (
    <div>
      <PageTitle title="Профиль" subtitle="Выход. Установка PWA и пуши — следующий шаг спринта." />
      <div className="space-y-3 rounded-xl border border-slate-800 bg-slate-900/40 p-4">
        <p className="text-sm text-slate-400">Email</p>
        <p className="text-lg">{user?.email}</p>
        <p className="text-sm text-slate-400">Роль</p>
        <p>{isAdmin ? 'admin — полный доступ' : 'viewer — только просмотр и экспорт'}</p>
      </div>

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
