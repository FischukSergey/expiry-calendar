import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'

import { getMe, login as loginReq, logout as logoutReq, logoutAll as logoutAllReq, register as registerReq } from '../api/endpoints.ts'
import { refreshSession, setAccessToken, setOnAuthLost } from '../api/client.ts'
import type { PublicUser } from '../api/types.ts'
import { AuthContext, type AuthContextValue } from './context.ts'

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<PublicUser | null>(null)
  const [ready, setReady] = useState(false)

  useEffect(() => {
    setOnAuthLost(() => setUser(null))
    return () => setOnAuthLost(null)
  }, [])

  useEffect(() => {
    let cancelled = false
    void (async () => {
      const ok = await refreshSession()
      if (ok && !cancelled) {
        try {
          const me = await getMe()
          if (!cancelled) {
            setUser(me)
          }
        } catch {
          setAccessToken(null)
        }
      }
      if (!cancelled) {
        setReady(true)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [])

  const applyPair = useCallback(async (access: string) => {
    setAccessToken(access)
    setUser(await getMe())
  }, [])

  const login = useCallback(
    async (email: string, password: string) => {
      const pair = await loginReq(email, password)
      await applyPair(pair.access_token)
    },
    [applyPair],
  )

  const register = useCallback(
    async (email: string, password: string) => {
      const pair = await registerReq(email, password)
      await applyPair(pair.access_token)
    },
    [applyPair],
  )

  const logout = useCallback(async () => {
    try {
      await logoutReq()
    } finally {
      setAccessToken(null)
      setUser(null)
    }
  }, [])

  const logoutAll = useCallback(async () => {
    try {
      await logoutAllReq()
    } finally {
      setAccessToken(null)
      setUser(null)
    }
  }, [])

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      ready,
      isAdmin: user?.role === 'admin',
      login,
      register,
      logout,
      logoutAll,
    }),
    [user, ready, login, register, logout, logoutAll],
  )

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
