import { useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api, coverSrc, money, type Order, type OrderSummary } from '../api'

const fmt = (s: string) => new Date(s).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })

export function OrderList() {
  const { data, isLoading } = useQuery({ queryKey: ['orders'], queryFn: () => api<OrderSummary[]>('/api/orders') })
  return (
    <div className="mx-auto max-w-[900px] px-4 py-8">
      <h1 className="mb-6 text-3xl font-extrabold">Order history</h1>
      {isLoading ? <div className="skeleton h-24" /> : !data?.length ? (
        <div className="bg-brand-mid p-10 text-center"><p className="text-xl font-bold">No orders yet</p><Link to="/store" className="btn-primary mt-4">Find a game</Link></div>
      ) : (
        <ul className="flex flex-col gap-3">
          {data.map((o) => (
            <li key={o.id}><Link to={`/orders/${o.id}`} className="flex items-center justify-between bg-brand-mid p-4 hover:bg-brand-mid/70">
              <div><p className="font-bold">Order #{o.id}</p><p className="text-sm text-muted">{fmt(o.createdAt)} · {o.itemCount} {o.itemCount === 1 ? 'key' : 'keys'}</p></div>
              <div className="text-right"><p className="text-lg font-extrabold">{money(o.totalCents)}</p><p className="text-xs font-bold uppercase text-mint">{o.status}</p></div>
            </Link></li>
          ))}
        </ul>
      )}
    </div>
  )
}

function KeyRow({ code }: { code: string }) {
  const [shown, setShown] = useState(false)
  const [copied, setCopied] = useState(false)
  return (
    <div className="mt-2 flex flex-wrap items-center gap-2">
      <code className="bg-brand-deep px-3 py-2 font-mono text-lg tracking-widest">{shown ? code : '•••••-•••••-•••••'}</code>
      <button className="btn-secondary !px-3 !py-2 text-sm" onClick={() => setShown((s) => !s)}>{shown ? 'Hide' : 'Reveal key'}</button>
      <button className="btn-secondary !px-3 !py-2 text-sm" onClick={() => { navigator.clipboard.writeText(code); setCopied(true); setTimeout(() => setCopied(false), 1500) }}>{copied ? 'Copied ✓' : 'Copy'}</button>
    </div>
  )
}

export function OrderDetail() {
  const { id } = useParams()
  const [sp] = useSearchParams()
  const { data, isLoading, isError } = useQuery({ queryKey: ['order', id], queryFn: () => api<Order>(`/api/orders/${id}`), retry: false })
  if (isLoading) return <div className="mx-auto max-w-[900px] p-4"><div className="skeleton h-60" /></div>
  if (isError || !data) return <div className="mx-auto max-w-xl px-4 py-20 text-center"><h1 className="text-2xl font-bold">Order not found</h1><Link to="/orders" className="btn-primary mt-6">Your orders</Link></div>
  return (
    <div className="mx-auto max-w-[900px] px-4 py-8">
      {sp.get('new') && (
        <div role="status" className="mb-6 bg-mint p-5 text-black"><p className="text-2xl font-extrabold">Payment successful 🎉</p><p className="font-semibold">Your keys are ready below. We’ve saved them to your order history too.</p></div>
      )}
      <h1 className="text-3xl font-extrabold">Order #{data.id}</h1>
      <p className="text-sm text-muted">{fmt(data.createdAt)} · Payment ref {data.paymentRef} · Total {money(data.totalCents)}</p>
      <ul className="mt-6 flex flex-col gap-3">
        {data.items.map((it, i) => (
          <li key={i} className="flex gap-4 bg-brand-mid p-4">
            <img src={coverSrc(it.coverUrl)} alt="" className="h-24 w-[72px] object-cover" />
            <div className="min-w-0 flex-1">
              <Link to={`/product/${it.slug}`} className="text-lg font-bold hover:underline">{it.name}</Link>
              <p className="text-sm text-muted">{it.platform} · {it.region} · sold by {it.seller} · {money(it.priceCents)}</p>
              <KeyRow code={it.key} />
            </div>
          </li>
        ))}
      </ul>
    </div>
  )
}
