import { useState, type FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'
import { api, ApiError, money } from '../api'
import { useCart } from '../cart'
import { useQuote } from './Cart'

export default function Checkout() {
  const { items, clear } = useCart()
  const nav = useNavigate()
  const qc = useQueryClient()
  const { data, isLoading } = useQuote()
  const [card, setCard] = useState({ Number: '4242 4242 4242 4242', Expiry: '12/30', CVC: '123', Name: '' })
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)

  if (items.length === 0) return <div className="mx-auto max-w-xl px-4 py-20 text-center"><h1 className="text-2xl font-bold">Nothing to check out</h1><Link to="/store" className="btn-primary mt-6">Browse games</Link></div>
  const blocked = data?.items.some((l) => !l.available)

  const pay = async (e: FormEvent) => {
    e.preventDefault()
    setError(''); setBusy(true)
    try {
      const { orderId } = await api<{ orderId: number }>('/api/checkout', { body: { items, card } })
      clear()
      qc.invalidateQueries()
      nav(`/orders/${orderId}?new=1`, { replace: true })
    } catch (err) {
      setError(err instanceof ApiError ? err.message : 'Payment failed. Please try again.')
      setBusy(false)
    }
  }
  const set = (k: keyof typeof card) => (e: React.ChangeEvent<HTMLInputElement>) => setCard({ ...card, [k]: e.target.value })

  return (
    <div className="mx-auto max-w-[1000px] px-4 py-8">
      <h1 className="mb-6 text-3xl font-extrabold">Checkout</h1>
      <div className="grid gap-6 md:grid-cols-[1fr_340px]">
        <form onSubmit={pay} className="flex flex-col gap-4 bg-brand-mid p-5">
          <h2 className="text-xl font-bold">Payment</h2>
          <p className="bg-black/25 p-3 text-sm text-muted">Test mode — no real money is charged. Use <b className="text-white">4242 4242 4242 4242</b> to succeed or <b className="text-white">4000 0000 0000 0002</b> to see a declined payment.</p>
          <label className="flex flex-col gap-1 text-sm font-semibold">Name on card<input className="field" value={card.Name} onChange={set('Name')} autoComplete="cc-name" /></label>
          <label className="flex flex-col gap-1 text-sm font-semibold">Card number<input className="field" inputMode="numeric" required value={card.Number} onChange={set('Number')} autoComplete="cc-number" /></label>
          <div className="grid grid-cols-2 gap-4">
            <label className="flex flex-col gap-1 text-sm font-semibold">Expiry<input className="field" required placeholder="MM/YY" value={card.Expiry} onChange={set('Expiry')} autoComplete="cc-exp" /></label>
            <label className="flex flex-col gap-1 text-sm font-semibold">CVC<input className="field" required inputMode="numeric" maxLength={4} value={card.CVC} onChange={set('CVC')} autoComplete="cc-csc" /></label>
          </div>
          {error && <p role="alert" className="bg-danger/20 p-3 font-semibold text-danger">{error}</p>}
          <button className="btn-primary text-lg" disabled={busy || isLoading || blocked}>{busy ? 'Processing…' : `Pay ${money(data?.totalCents ?? 0)}`}</button>
          {blocked && <p className="text-sm text-danger">Some items are out of stock. <Link to="/cart" className="underline">Review your cart</Link>.</p>}
        </form>
        <aside className="self-start bg-brand-mid p-5">
          <h2 className="mb-3 text-xl font-bold">Order summary</h2>
          <ul className="flex flex-col gap-2 text-sm">
            {data?.items.map((l) => <li key={l.offerId} className="flex justify-between gap-3"><span>{l.qty}× {l.name} <span className="text-muted">({l.platform})</span></span><span className="font-bold">{money(l.priceCents * l.qty)}</span></li>)}
          </ul>
          <p className="mt-4 flex justify-between border-t border-white/10 pt-3 text-lg font-extrabold"><span>Total</span><span>{money(data?.totalCents ?? 0)}</span></p>
        </aside>
      </div>
    </div>
  )
}
