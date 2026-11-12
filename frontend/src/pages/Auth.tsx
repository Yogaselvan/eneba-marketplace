import { useState, type FormEvent } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { ApiError } from '../api'
import { useAuth } from '../auth'

export default function Auth({ mode }: { mode: 'login' | 'register' }) {
  const { signIn } = useAuth()
  const nav = useNavigate()
  const loc = useLocation()
  const from = (loc.state as { from?: string } | null)?.from ?? '/'
  const [f, setF] = useState({ name: '', email: '', password: '' })
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const reg = mode === 'register'

  const submit = async (e: FormEvent) => {
    e.preventDefault()
    setError(''); setBusy(true)
    try {
      await signIn(mode, reg ? f : { email: f.email, password: f.password })
      nav(from, { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Something went wrong')
      setBusy(false)
    }
  }
  return (
    <div className="mx-auto max-w-md px-4 py-12">
      <form onSubmit={submit} className="flex flex-col gap-4 bg-brand-mid p-6">
        <h1 className="text-3xl font-extrabold">{reg ? 'Create account' : 'Log in'}</h1>
        {from !== '/' && <p className="text-sm text-muted">Sign in to continue to checkout — your cart is saved.</p>}
        {reg && <label className="flex flex-col gap-1 text-sm font-semibold">Name<input className="field" value={f.name} onChange={(e) => setF({ ...f, name: e.target.value })} autoComplete="name" /></label>}
        <label className="flex flex-col gap-1 text-sm font-semibold">Email<input className="field" type="email" required value={f.email} onChange={(e) => setF({ ...f, email: e.target.value })} autoComplete="email" /></label>
        <label className="flex flex-col gap-1 text-sm font-semibold">Password<input className="field" type="password" required minLength={reg ? 8 : undefined} value={f.password} onChange={(e) => setF({ ...f, password: e.target.value })} autoComplete={reg ? 'new-password' : 'current-password'} />{reg && <span className="text-xs font-normal text-muted">At least 8 characters</span>}</label>
        {error && <p role="alert" className="bg-danger/20 p-3 font-semibold text-danger">{error}</p>}
        <button className="btn-primary" disabled={busy}>{busy ? 'Please wait…' : reg ? 'Register' : 'Log in'}</button>
        <p className="text-center text-sm text-muted">{reg ? 'Already have an account?' : 'New here?'} <Link state={loc.state} to={reg ? '/login' : '/register'} className="font-bold text-accent">{reg ? 'Log in' : 'Create an account'}</Link></p>
      </form>
    </div>
  )
}
