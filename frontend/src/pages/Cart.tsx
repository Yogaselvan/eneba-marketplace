import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api, coverSrc, money, type QuoteLine } from '../api'
import { useCart } from '../cart'

export function useQuote() {
  const { items } = useCart()
  return useQuery({
    queryKey: ['quote', items],
    queryFn: () => api<{ items: QuoteLine[]; totalCents: number }>('/api/cart/quote', { body: { items } }),
    enabled: items.length > 0,
    placeholderData: (prev) => prev,
  })
}

export default function Cart() {
  const { items, setQty, remove } = useCart()
  const { data, isLoading, isError } = useQuote()

  if (items.length === 0) return (
    <div className="mx-auto max-w-xl px-4 py-20 text-center">
      <h1 className="text-3xl font-extrabold">Your cart is empty</h1>
      <p className="mt-2 text-muted">Find a game you love and compare offers from several sellers.</p>
      <Link to="/store" className="btn-primary mt-6">Browse games</Link>
    </div>
  )
  const lines = data?.items ?? []
  const blocked = lines.some((l) => !l.available)

  return (
    <div className="mx-auto max-w-[1100px] px-4 py-8">
      <h1 className="mb-6 text-3xl font-extrabold">Cart</h1>
      {isError && <p className="mb-4 bg-danger/20 p-3">Couldn’t refresh prices. Check your connection and try again.</p>}
      <div className="grid gap-6 lg:grid-cols-[1fr_320px]">
        <ul className="flex flex-col gap-3">
          {isLoading && items.map((i) => <li key={i.offerId} className="skeleton h-28" />)}
          {lines.map((l) => (
            <li key={l.offerId} className="flex gap-4 bg-brand-mid p-3">
              <img src={coverSrc(l.coverUrl)} alt="" className="h-24 w-[72px] object-cover" />
              <div className="min-w-0 flex-1">
                <Link to={`/product/${l.slug}`} className="block truncate text-lg font-bold hover:underline">{l.name}</Link>
                <p className="text-sm text-muted">{l.platform} · {l.region} · sold by {l.seller}</p>
                {!l.available && <p className="mt-1 text-sm font-bold text-danger">Only {l.stock} in stock — lower the quantity or remove this item.</p>}
                <div className="mt-2 flex items-center gap-3">
                  <div className="flex items-center border border-brand-light">
                    <button className="px-3 py-1 text-lg" aria-label={`Decrease quantity of ${l.name}`} onClick={() => setQty(l.offerId, l.qty - 1)} disabled={l.qty <= 1}>−</button>
                    <span className="w-8 text-center font-bold" aria-live="polite">{l.qty}</span>
                    <button className="px-3 py-1 text-lg" aria-label={`Increase quantity of ${l.name}`} onClick={() => setQty(l.offerId, l.qty + 1)} disabled={l.qty >= 5}>+</button>
                  </div>
                  <button className="text-sm text-muted underline hover:text-white" onClick={() => remove(l.offerId)}>Remove</button>
                </div>
              </div>
              <p className="text-lg font-extrabold">{money(l.priceCents * l.qty)}</p>
            </li>
          ))}
        </ul>
        <aside className="self-start bg-brand-mid p-5">
          <h2 className="text-xl font-bold">Summary</h2>
          <p className="mt-4 flex justify-between text-lg"><span>Total</span><span className="font-extrabold">{money(data?.totalCents ?? 0)}</span></p>
          <p className="mt-1 text-xs text-muted">Keys are delivered instantly after payment.</p>
          {blocked ? <button className="btn-primary mt-5 w-full" disabled>Fix stock issues to continue</button>
            : <Link to="/checkout" className="btn-primary mt-5 w-full">Go to checkout</Link>}
        </aside>
      </div>
    </div>
  )
}
