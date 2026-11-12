import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { api, money, type Card, type Facet, type Filters } from '../api'
import ProductCard, { CardSkeletons } from '../components/ProductCard'

const SORTS = [
  ['popular', 'Popularity'], ['price_asc', 'Price: low to high'], ['price_desc', 'Price: high to low'],
  ['discount', 'Biggest discount'], ['newest', 'Newest'], ['name', 'Name A–Z'],
] as const

function FacetGroup({ title, facets, selected, onToggle }: { title: string; facets: Facet[]; selected: string[]; onToggle: (v: string) => void }) {
  return (
    <fieldset className="border-b border-white/10 pb-4">
      <legend className="mb-2 font-bold">{title}</legend>
      <div className="flex flex-col gap-1.5">
        {facets.map((f) => (
          <label key={f.value} className="flex cursor-pointer items-center gap-2 text-sm">
            <input type="checkbox" checked={selected.includes(f.value)} onChange={() => onToggle(f.value)} className="h-4 w-4 accent-[#fad318]" />
            <span className="flex-1">{f.value}</span><span className="text-muted">{f.count}</span>
          </label>
        ))}
      </div>
    </fieldset>
  )
}

export default function Store() {
  const [params, setParams] = useSearchParams()
  const q = params.get('q') ?? ''
  const sort = params.get('sort') ?? (q ? 'relevance' : 'popular')
  const page = Number(params.get('page') ?? 1)
  const list = (k: string) => (params.get(k) ? params.get(k)!.split(',') : [])
  const platforms = list('platform'), regions = list('region'), genres = list('genre')
  const onSale = params.get('onSale') === '1'

  const [min, setMin] = useState(params.get('minPrice') ? String(Number(params.get('minPrice')) / 100) : '')
  const [max, setMax] = useState(params.get('maxPrice') ? String(Number(params.get('maxPrice')) / 100) : '')
  const [showFilters, setShowFilters] = useState(false)
  useEffect(() => { if (!params.get('minPrice')) setMin(''); if (!params.get('maxPrice')) setMax('') }, [params])

  const update = (patch: Record<string, string | null>) => {
    const next = new URLSearchParams(params)
    Object.entries(patch).forEach(([k, v]) => (v ? next.set(k, v) : next.delete(k)))
    if (!('page' in patch)) next.delete('page')
    setParams(next)
  }
  const toggle = (key: string, current: string[], v: string) =>
    update({ [key]: (current.includes(v) ? current.filter((x) => x !== v) : [...current, v]).join(',') || null })
  const applyPrice = () => update({ minPrice: min ? String(Math.round(Number(min) * 100)) : null, maxPrice: max ? String(Math.round(Number(max) * 100)) : null })

  const filters = useQuery({ queryKey: ['filters'], queryFn: () => api<Filters>('/api/filters'), staleTime: 300_000 })
  const products = useQuery({
    queryKey: ['products', params.toString()],
    queryFn: () => api<{ items: Card[]; total: number; pageSize: number }>(`/api/products?${params.toString()}`),
    placeholderData: keepPreviousData,
  })
  const total = products.data?.total ?? 0
  const pages = Math.max(1, Math.ceil(total / (products.data?.pageSize ?? 24)))
  const activeCount = platforms.length + regions.length + genres.length + (params.get('minPrice') || params.get('maxPrice') ? 1 : 0) + (onSale ? 1 : 0)

  const sidebar = (
    <div className="flex flex-col gap-4">
      <label className="flex cursor-pointer items-center gap-2 border-b border-white/10 pb-4 font-bold">
        <input type="checkbox" checked={onSale} onChange={() => update({ onSale: onSale ? null : '1' })} className="h-4 w-4 accent-[#fad318]" /> On sale only
      </label>
      <fieldset className="border-b border-white/10 pb-4">
        <legend className="mb-2 font-bold">Price (€)</legend>
        <div className="flex items-center gap-2">
          <input className="field !p-2" inputMode="decimal" placeholder="Min" aria-label="Minimum price" value={min} onChange={(e) => setMin(e.target.value)} onBlur={applyPrice} onKeyDown={(e) => e.key === 'Enter' && applyPrice()} />
          <span>–</span>
          <input className="field !p-2" inputMode="decimal" placeholder={filters.data ? String(Math.ceil(filters.data.maxPriceCents / 100)) : 'Max'} aria-label="Maximum price" value={max} onChange={(e) => setMax(e.target.value)} onBlur={applyPrice} onKeyDown={(e) => e.key === 'Enter' && applyPrice()} />
        </div>
      </fieldset>
      {filters.data && <>
        <FacetGroup title="Platform" facets={filters.data.platforms} selected={platforms} onToggle={(v) => toggle('platform', platforms, v)} />
        <FacetGroup title="Region" facets={filters.data.regions} selected={regions} onToggle={(v) => toggle('region', regions, v)} />
        <FacetGroup title="Genre" facets={filters.data.genres} selected={genres} onToggle={(v) => toggle('genre', genres, v)} />
      </>}
      {activeCount > 0 && <button className="btn-secondary" onClick={() => { setMin(''); setMax(''); setParams(q ? { q } : {}) }}>Clear all filters ({activeCount})</button>}
    </div>
  )

  return (
    <div className="mx-auto max-w-[1400px] px-4 py-6">
      <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-extrabold md:text-3xl">{q ? <>Results for “{q}”</> : 'All games'}</h1>
          <p className="text-sm text-muted" aria-live="polite">{products.data ? `${total} ${total === 1 ? 'product' : 'products'}` : 'Loading…'}</p>
        </div>
        <div className="flex items-center gap-2">
          <button className="btn-secondary lg:hidden" onClick={() => setShowFilters((s) => !s)} aria-expanded={showFilters}>Filters{activeCount ? ` (${activeCount})` : ''}</button>
          <label className="flex items-center gap-2 text-sm"><span className="text-muted">Sort by</span>
            <select value={sort} onChange={(e) => update({ sort: e.target.value })} className="field !w-auto !py-2">
              {q && <option value="relevance">Best match</option>}
              {SORTS.map(([v, l]) => <option key={v} value={v}>{l}</option>)}
            </select>
          </label>
        </div>
      </div>

      <div className="grid gap-6 lg:grid-cols-[250px_1fr]">
        <aside className={`${showFilters ? 'block' : 'hidden'} bg-brand-mid p-4 lg:block lg:self-start`} aria-label="Filters">{sidebar}</aside>
        <div>
          {products.isError ? (
            <p className="bg-brand-mid p-6">Something went wrong loading products. <button className="font-bold text-accent underline" onClick={() => products.refetch()}>Try again</button></p>
          ) : products.isLoading ? (
            <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 xl:grid-cols-4"><CardSkeletons n={8} /></div>
          ) : products.data!.items.length === 0 ? (
            <div className="bg-brand-mid p-10 text-center">
              <p className="text-xl font-bold">No games found</p>
              <p className="mt-1 text-muted">Try a different spelling or remove some filters.</p>
            </div>
          ) : (
            <div className={`grid grid-cols-2 gap-4 sm:grid-cols-3 xl:grid-cols-4 ${products.isPlaceholderData ? 'opacity-60' : ''}`}>
              {products.data!.items.map((p) => <ProductCard key={p.id} p={p} />)}
            </div>
          )}
          {pages > 1 && (
            <nav className="mt-8 flex items-center justify-center gap-3" aria-label="Pagination">
              <button className="btn-secondary" disabled={page <= 1} onClick={() => { update({ page: String(page - 1) }); window.scrollTo(0, 0) }}>Previous</button>
              <span className="text-sm">Page {page} of {pages}</span>
              <button className="btn-secondary" disabled={page >= pages} onClick={() => { update({ page: String(page + 1) }); window.scrollTo(0, 0) }}>Next</button>
            </nav>
          )}
          {filters.data && <p className="mt-6 text-xs text-muted">Prices from {money(filters.data.minPriceCents)}</p>}
        </div>
      </div>
    </div>
  )
}
