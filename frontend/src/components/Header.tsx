import { useEffect, useRef, useState, type FormEvent } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api, coverSrc, money, type Card } from '../api'
import { useAuth } from '../auth'
import { useCart } from '../cart'

function useDebounced<T>(value: T, ms = 200) {
  const [v, setV] = useState(value)
  useEffect(() => { const t = setTimeout(() => setV(value), ms); return () => clearTimeout(t) }, [value, ms])
  return v
}

export function Logo() {
  return <Link to="/" className="flex items-center gap-2 text-3xl font-extrabold tracking-tight lowercase" aria-label="Home">
    <span className="inline-block h-8 w-8 bg-gradient-to-br from-accent via-mint to-brand-light" style={{ clipPath: 'polygon(50% 0, 100% 40%, 80% 100%, 0 80%, 10% 20%)' }} />
    marketplace
  </Link>
}

function SearchBox() {
  const nav = useNavigate()
  const [params] = useSearchParams()
  const [q, setQ] = useState(params.get('q') ?? '')
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(-1)
  const dq = useDebounced(q.trim())
  const box = useRef<HTMLFormElement>(null)

  const { data } = useQuery({
    queryKey: ['suggest', dq],
    queryFn: () => api<Card[]>(`/api/suggest?q=${encodeURIComponent(dq)}`),
    enabled: dq.length >= 2,
    staleTime: 60_000,
  })
  useEffect(() => {
    const close = (e: MouseEvent) => { if (!box.current?.contains(e.target as Node)) setOpen(false) }
    document.addEventListener('mousedown', close)
    return () => document.removeEventListener('mousedown', close)
  }, [])

  const submit = (e: FormEvent) => {
    e.preventDefault()
    setOpen(false)
    if (active >= 0 && data?.[active]) return nav(`/product/${data[active].slug}`)
    nav(q.trim() ? `/store?q=${encodeURIComponent(q.trim())}` : '/store')
  }
  const onKey = (e: React.KeyboardEvent) => {
    if (!data?.length) return
    if (e.key === 'ArrowDown') { e.preventDefault(); setActive((a) => (a + 1) % data.length) }
    if (e.key === 'ArrowUp') { e.preventDefault(); setActive((a) => (a <= 0 ? data.length - 1 : a - 1)) }
    if (e.key === 'Escape') setOpen(false)
  }

  return (
    <form ref={box} onSubmit={submit} role="search" className="relative flex-1 min-w-0">
      <div className="flex items-center border border-white bg-transparent focus-within:bg-brand-deep">
        <svg className="ml-3 h-5 w-5 shrink-0" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="7" /><path d="m20 20-3.5-3.5" /></svg>
        <input value={q} onChange={(e) => { setQ(e.target.value); setOpen(true); setActive(-1) }} onFocus={() => setOpen(true)} onKeyDown={onKey}
          placeholder="Search for games…" aria-label="Search games" autoComplete="off"
          className="w-full bg-transparent px-3 py-2.5 text-lg font-bold outline-none placeholder:text-white/50" />
      </div>
      {open && dq.length >= 2 && data && (
        <ul className="absolute left-0 right-0 top-full z-50 border border-brand-light bg-brand-deep shadow-2xl" role="listbox">
          {data.length === 0 && <li className="px-4 py-3 text-muted">No games match “{dq}”</li>}
          {data.map((p, i) => (
            <li key={p.id} role="option" aria-selected={i === active}>
              <Link to={`/product/${p.slug}`} onClick={() => setOpen(false)} className={`flex items-center gap-3 px-3 py-2 ${i === active ? 'bg-brand-mid' : 'hover:bg-brand-mid'}`}>
                <img src={coverSrc(p.coverUrl)} alt="" className="h-12 w-9 object-cover" />
                <span className="min-w-0 flex-1"><span className="block truncate font-bold">{p.name}</span><span className="text-xs text-muted">{p.platform} · {p.region}</span></span>
                <span className="font-extrabold text-mint">{money(p.priceCents)}</span>
              </Link>
            </li>
          ))}
          <li><Link to={`/store?q=${encodeURIComponent(dq)}`} onClick={() => setOpen(false)} className="block border-t border-white/10 px-4 py-2 text-center text-sm font-bold text-accent hover:bg-brand-mid">See all results</Link></li>
        </ul>
      )}
    </form>
  )
}

export default function Header() {
  const { user, signOut } = useAuth()
  const { count } = useCart()
  const nav = useNavigate()
  return (
    <header>
      <div className="mx-auto flex max-w-[1400px] flex-wrap items-center gap-x-6 gap-y-3 px-4 py-4">
        <Logo />
        <div className="order-3 w-full md:order-none md:w-auto md:flex-1 md:flex"><SearchBox /></div>
        <nav className="ml-auto flex items-center gap-5 text-sm font-semibold md:ml-0" aria-label="Account">
          <Link to="/cart" className="relative flex items-center gap-1" aria-label={`Cart, ${count} items`}>
            <svg className="h-6 w-6" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M3 4h2l2.4 11h10.2L20 7H6" /><circle cx="9" cy="19" r="1.5" /><circle cx="17" cy="19" r="1.5" /></svg>
            {count > 0 && <span className="absolute -right-2 -top-2 flex h-5 min-w-5 items-center justify-center bg-accent px-1 text-xs font-extrabold text-black">{count}</span>}
          </Link>
          {user ? (
            <>
              <Link to="/orders" className="hover:underline">{user.name}</Link>
              <button onClick={() => { signOut(); nav('/') }} className="text-muted hover:text-white">Log out</button>
            </>
          ) : (
            <span><Link to="/login" className="hover:underline">Log in</Link> | <Link to="/register" className="hover:underline">Register</Link></span>
          )}
        </nav>
      </div>
      <div className="bg-black/25">
        <div className="no-scrollbar mx-auto flex max-w-[1400px] gap-6 overflow-x-auto px-4 py-3 text-[15px] font-medium whitespace-nowrap">
          <Link to="/store" className="hover:text-accent">All games</Link>
          <Link to="/store?sort=discount&onSale=1" className="hover:text-accent">Deals</Link>
          <Link to="/store?sort=newest" className="hover:text-accent">New releases</Link>
          {['Steam', 'Xbox', 'PlayStation', 'Nintendo', 'EA App'].map((p) => <Link key={p} to={`/store?platform=${encodeURIComponent(p)}`} className="hover:text-accent">{p}</Link>)}
        </div>
      </div>
    </header>
  )
}
