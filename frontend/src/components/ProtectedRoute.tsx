import { Navigate, Outlet, useLocation } from 'react-router-dom'

import { useAuth } from '../hooks/useAuth.ts'

export function ProtectedRoute() {
  const { user, ready } = useAuth()
  const location = useLocation()

  if (!ready) {
    return <div className="grid min-h-screen place-items-center text-slate-400">Загрузка…</div>
  }
  if (!user) {
    return <Navigate to="/login" replace state={{ from: location.pathname }} />
  }
  return <Outlet />
}

export function AdminRoute() {
  const { isAdmin, ready } = useAuth()
  if (!ready) {
    return <div className="grid min-h-screen place-items-center text-slate-400">Загрузка…</div>
  }
  if (!isAdmin) {
    return <Navigate to="/" replace />
  }
  return <Outlet />
}

export function GuestRoute() {
  const { user, ready, afterAuthPath } = useAuth()
  const location = useLocation()
  if (!ready) {
    return <div className="grid min-h-screen place-items-center text-slate-400">Загрузка…</div>
  }
  if (user) {
    const from = (location.state as { from?: string } | null)?.from
    // Register всегда на свой список; login чтит исходную страницу.
    const dest = afterAuthPath === '/items' ? '/items' : from && from !== '/login' ? from : afterAuthPath
    return <Navigate to={dest} replace />
  }
  return <Outlet />
}
