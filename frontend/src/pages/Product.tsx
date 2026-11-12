import { useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api, ApiError, coverSrc, money, type ProductDetail } from '../api'
import { useCart } from '../cart'

export default function Product() {
  const { slug } = useParams()
  const nav = useNavigate()
  const { add } = useCart()
  const [added, setAdded] = useState<number | null>(null)
  const { data: p, isLoading, error } = useQuery({ queryKey: ['product', slug], queryFn: () => api<ProductDetail>(`/api/products/${slug}`), retry: false })

  if (isLoading) return <div className="mx-auto max-w-[1100px] p-4"><div className="skeleton h-96" /></div>
  if (error || !p) return (
    <div className="mx-auto max-w-xl p-10 text-center">
      <h1 className="text-2xl font-bold">{(error as ApiError)?.status === 404 ? 'Product not found' : 'Could not load product'}</h1>
      <Link to="/store" className="btn-primary mt-6">Browse the store</Link>
    </div>
  )

  const best = p.offers[0]
  const buy = (offerId: number, goToCart: boolean) => { add(offerId); setAdded(offerId); if (goToCart) nav('/cart') }

  return (
    <div className="mx-auto max-w-[1100px] px-4 py-8">
      <nav className="mb-4 text-sm text-muted" aria-label="Breadcrumb"><Link to="/store" className="hover:text-white">Store</Link> / <Link to={`/store?platform=${encodeURIComponent(p.platform)}`} className="hover:text-white">{p.platform}</Link> / {p.name}</nav>
      <div className="grid gap-8 md:grid-cols-[300px_1fr]">
        <img src={coverSrc(p.coverUrl)} alt={`${p.name} cover`} className="w-full max-w-[300px] justify-self-center shadow-2xl shadow-black/40" />
        <div>
          <h1 className="text-3xl font-extrabold md:text-4xl">{p.name} <span className="text-muted">({p.platform})</span></h1>
          <ul className="mt-3 flex flex-wrap gap-2 text-sm">
            {[p.platform, p.region, p.genre, String(p.releaseYear)].map((t) => <li key={t} className="bg-brand-mid px-3 py-1 font-semibold">{t}</li>)}
            <li className="bg-mint px-3 py-1 font-bold text-black">Instant key delivery</li>
          </ul>
          <p className="mt-5 max-w-prose leading-relaxed text-white/85">{p.description}</p>

          <div className="mt-6 bg-brand-mid p-5">
            <div className="flex flex-wrap items-end justify-between gap-4">
              <div>
                <p className="text-sm text-muted">Best price from {p.offers.length} {p.offers.length === 1 ? 'seller' : 'sellers'}</p>
                <p className="flex items-center gap-3">
                  <span className="text-4xl font-extrabold">{money(p.priceCents)}</span>
                  {p.discountPercent > 0 && <><span className="text-muted line-through">{money(p.listPriceCents)}</span><span className="bg-mint px-2 py-0.5 font-extrabold text-black">-{p.discountPercent}%</span></>}
                </p>
              </div>
              <button className="btn-primary text-lg" onClick={() => buy(best.id, true)}>Buy now</button>
            </div>
          </div>
        </div>
      </div>

      <section className="mt-10" aria-labelledby="offers-h">
        <h2 id="offers-h" className="mb-4 text-2xl font-bold">Compare offers</h2>
        <div className="overflow-x-auto"><table className="w-full min-w-[560px] text-left">
          <thead className="text-sm text-muted"><tr><th className="p-3 font-semibold">Seller</th><th className="p-3 font-semibold">Rating</th><th className="p-3 font-semibold">In stock</th><th className="p-3 text-right font-semibold">Price</th><th className="p-3"><span className="sr-only">Action</span></th></tr></thead>
          <tbody>
            {p.offers.map((o, i) => (
              <tr key={o.id} className={`border-t border-white/10 ${i === 0 ? 'bg-brand-mid' : ''}`}>
                <td className="p-3 font-bold">{o.seller}{i === 0 && <span className="ml-2 bg-accent px-1.5 py-0.5 text-[11px] font-extrabold text-black">BEST PRICE</span>}</td>
                <td className="p-3">★ {o.sellerRating.toFixed(1)} <span className="text-xs text-muted">({o.sellerSales.toLocaleString()} sales)</span></td>
                <td className="p-3">{o.stock <= 3 ? <span className="text-accent">Only {o.stock} left</span> : o.stock}</td>
                <td className="p-3 text-right text-lg font-extrabold">{money(o.priceCents)}</td>
                <td className="p-3 text-right">
                  <button className={added === o.id ? 'btn-secondary' : 'btn-primary'} onClick={() => buy(o.id, false)}>{added === o.id ? '✓ Added' : 'Add to cart'}</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table></div>
        {added && <p className="mt-4" role="status"><Link to="/cart" className="font-bold text-accent underline">Go to cart →</Link></p>}
      </section>
    </div>
  )
}
