import { createContext, useContext, useEffect, useState, type ReactNode } from 'react'
import { api, tokenStore, type User } from './api'

interface AuthCtx {
  user: User | null
  loading: boolean
  signIn: (mode: 'login' | 'register', body: Record<string, string>) => Promise<void>
  signOut: () => void
}
const Ctx = createContext<AuthCtx>(null!)
export const useAuth = () => useContext(Ctx)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(!!tokenStore.get())

  useEffect(() => {
    if (!tokenStore.get()) return
    api<User>('/api/me').then(setUser).catch(() => tokenStore.clear()).finally(() => setLoading(false))
  }, [])

  const signIn: AuthCtx['signIn'] = async (mode, body) => {
    const res = await api<{ token: string; user: User }>(`/api/auth/${mode}`, { body })
    tokenStore.set(res.token)
    setUser(res.user)
  }
  const signOut = () => { tokenStore.clear(); setUser(null) }

  return <Ctx.Provider value={{ user, loading, signIn, signOut }}>{children}</Ctx.Provider>
}
