import { createContext } from 'react'

import type { PublicUser } from '../api/types.ts'

export type AuthContextValue = {
  user: PublicUser | null
  ready: boolean
  isAdmin: boolean
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
  logoutAll: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)
