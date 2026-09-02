import { NavLink, Outlet } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'

import { listNotifications } from '../api/endpoints.ts'
import { useAuth } from '../hooks/useAuth.ts'
import { usePush } from '../hooks/usePush.ts'
import { useSSE } from '../hooks/useSSE.ts'
import { InstallBanner } from './InstallBanner.tsx'

type NavItem = { to: string; label: string; admin?: boolean; badge?: boolean }

const mainNav: NavItem[] = [
  { to: '/', label: 'Обзор' },
  { to: '/items', label: 'Записи' },
  { to: '/calendar', label: 'Календарь' },
  { to: '/categories', label: 'Категории' },
  { to: '/notifications', label: 'Уведомления', badge: true },
  { to: '/import', label: 'Импорт', admin: true },
  { to: '/audit', label: 'Аудит', admin: true },
]

const mobileTabs: NavItem[] = [
  { to: '/', label: 'Обзор' },
  { to: '/items', label: 'Записи' },
  { to: '/calendar', label: 'Календарь' },
  { to: '/notifications', label: 'Лента', badge: true },
  { to: '/profile', label: 'Ещё' },
]

function linkClass(active: boolean): string {
  return [
    'block rounded-lg px-3 py-2 text-sm transition',
    active ? 'bg-slate-800 text-teal-300' : 'text-slate-300 hover:bg-slate-800/70 hover:text-slate-50',
  ].join(' ')
}

export function Layout() {
  const { user, isAdmin } = useAuth()
  useSSE()
  usePush()
  const unread = useQuery({
    queryKey: ['notifications', 'unread-count'],
    queryFn: () => listNotifications({ unread: true, page: 1, per_page: 1 }),
  })
  const unreadTotal = unread.data?.total ?? 0

  const visible = mainNav.filter((item) => !item.admin || isAdmin)

  return (
    <div className="min-h-screen bg-slate-950 text-slate-50 lg:flex">
      <aside className="hidden w-56 shrink-0 border-r border-slate-800 p-4 lg:flex lg:flex-col">
        <div className="px-3 pb-6">
          <p className="text-[11px] tracking-[0.2em] text-slate-500 uppercase">Duekeep</p>
          <p className="mt-1 text-lg font-semibold">Обязательства</p>
        </div>
        <nav className="flex flex-1 flex-col gap-1">
          {visible.map((item) => (
            <NavLink key={item.to} to={item.to} end={item.to === '/'} className={({ isActive }) => linkClass(isActive)}>
              <span className="flex items-center justify-between gap-2">
                {item.label}
                {item.badge && unreadTotal > 0 ? (
                  <span className="rounded-full bg-teal-500 px-1.5 text-[11px] font-semibold text-slate-950">
                    {unreadTotal > 99 ? '99+' : unreadTotal}
                  </span>
                ) : null}
              </span>
            </NavLink>
          ))}
        </nav>
        <NavLink to="/profile" className={({ isActive }) => `${linkClass(isActive)} mt-4`}>
          <span className="block truncate text-sm">{user?.email}</span>
        </NavLink>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col pb-[calc(5.25rem+env(safe-area-inset-bottom))] lg:pb-0">
        <InstallBanner />
        <header className="flex items-center justify-between border-b border-slate-800 px-4 py-3 lg:hidden">
          <span className="text-sm font-semibold tracking-wide">Duekeep</span>
          <NavLink to="/notifications" className="relative text-sm text-slate-300">
            Колокольчик
            {unreadTotal > 0 ? (
              <span className="ml-2 rounded-full bg-teal-500 px-1.5 text-[11px] font-semibold text-slate-950">
                {unreadTotal}
              </span>
            ) : null}
          </NavLink>
        </header>
        <main className="mx-auto w-full max-w-6xl flex-1 px-4 py-6">
          <Outlet />
        </main>
      </div>

      <nav className="fixed inset-x-0 bottom-0 z-10 grid grid-cols-5 border-t border-slate-800 bg-slate-950/95 px-0.5 py-1.5 pb-[max(0.35rem,env(safe-area-inset-bottom))] backdrop-blur lg:hidden">
        {mobileTabs.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) =>
              `flex flex-col items-center justify-center gap-0.5 rounded-lg px-0.5 py-1.5 text-center text-sm leading-tight ${
                isActive ? 'text-teal-300' : 'text-slate-400'
              }`
            }
          >
            <span className="relative max-w-full whitespace-normal">
              {item.label}
              {item.badge && unreadTotal > 0 ? (
                <span className="absolute -top-2 -right-3 rounded-full bg-teal-500 px-1 text-[10px] font-semibold text-slate-950">
                  {unreadTotal > 9 ? '9+' : unreadTotal}
                </span>
              ) : null}
            </span>
          </NavLink>
        ))}
      </nav>
    </div>
  )
}
