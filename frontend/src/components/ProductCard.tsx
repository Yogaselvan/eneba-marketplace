import { Link } from 'react-router-dom'
import { coverSrc, money, type Card } from '../api'

export default function ProductCard({ p }: { p: Card }) {
  return (
    <Link to={`/product/${p.slug}`} className="group flex flex-col bg-brand-mid transition hover:-translate-y-1 hover:shadow-xl hover:shadow-black/30">
      <div className="relative aspect-[3/4] overflow-hidden bg-brand-deep">
        <img src={coverSrc(p.coverUrl)} alt={`${p.name} cover`} loading="lazy" className="h-full w-full object-cover transition group-hover:scale-105" />
        <span className="absolute left-0 top-0 bg-black/60 px-2 py-1 text-[11px] font-bold uppercase tracking-wide">{p.platform}</span>
      </div>
      <div className="flex flex-1 flex-col gap-1 p-3">
        <h3 className="line-clamp-2 text-[15px] font-bold leading-snug">{p.name}</h3>
        <p className="text-xs text-muted">{p.region}</p>
        <div className="mt-auto pt-2">
          {p.discountPercent > 0 && (
            <p className="flex items-center gap-2 text-xs"><span className="text-muted line-through">{money(p.listPriceCents)}</span><span className="bg-mint px-1.5 font-extrabold text-black">-{p.discountPercent}%</span></p>
          )}
          <p className="text-lg font-extrabold">{money(p.priceCents)}</p>
          <p className="text-xs text-muted">{p.offerCount} {p.offerCount === 1 ? 'offer' : 'offers'}</p>
        </div>
      </div>
    </Link>
  )
}

export function CardSkeletons({ n = 6 }: { n?: number }) {
  return <>{Array.from({ length: n }).map((_, i) => <div key={i} className="skeleton aspect-[3/5]" />)}</>
}
