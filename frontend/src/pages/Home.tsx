import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api, type Banner, type Card } from '../api'
import ProductCard, { CardSkeletons } from '../components/ProductCard'

// Every block fetches from its own endpoint and renders exactly what it receives.
function Banners() {
  const { data, isError } = useQuery({ queryKey: ['banners'], queryFn: () => api<Banner[]>('/api/home/banners') })
  const [i, setI] = useState(0)
  useEffect(() => {
    if (!data?.length) return
    const t = setInterval(() => setI((x) => (x + 1) % data.length), 6000)
    return () => clearInterval(t)
  }, [data])
  if (isError) return null
  if (!data) return <div className="skeleton h-64 md:h-80" />
  const b = data[i % data.length]
  return (
    <section aria-roledescription="carousel" aria-label="Promotions" className="relative">
      <div className="flex h-64 items-center px-6 transition-colors md:h-80 md:px-16" style={{ background: `linear-gradient(110deg, ${b.colorA}, ${b.colorB})` }}>
        <div className="max-w-xl">
          <h1 className="text-3xl font-extrabold leading-tight md:text-5xl">{b.title}</h1>
          <p className="mt-3 text-lg text-white/85">{b.subtitle}</p>
          <Link to={b.link} className="btn-primary mt-6">{b.cta}</Link>
        </div>
      </div>
      <div className="absolute bottom-3 left-1/2 flex -translate-x-1/2 gap-2">
        {data.map((x, n) => <button key={x.id} aria-label={`Show promotion ${n + 1}`} onClick={() => setI(n)} className={`h-2 w-6 ${n === i ? 'bg-accent' : 'bg-white/40'}`} />)}
      </div>
    </section>
  )
}

function Block({ title, endpoint, more }: { title: string; endpoint: string; more: string }) {
  const { data, isLoading, isError, refetch } = useQuery({ queryKey: [endpoint], queryFn: () => api<Card[]>(endpoint) })
  return (
    <section className="mt-10" aria-label={title}>
      <div className="mb-4 flex items-end justify-between">
        <h2 className="text-2xl font-bold md:text-3xl">{title}</h2>
        <Link to={more} className="text-sm font-bold text-accent hover:underline">See all</Link>
      </div>
      {isError ? (
        <p className="bg-brand-mid p-4 text-muted">Couldn’t load this section. <button onClick={() => refetch()} className="font-bold text-accent underline">Retry</button></p>
      ) : (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-6">
          {isLoading ? <CardSkeletons /> : data!.slice(0, 6).map((p) => <ProductCard key={p.id} p={p} />)}
        </div>
      )}
    </section>
  )
}

export default function Home() {
  return (
    <>
      <Banners />
      <div className="mx-auto max-w-[1400px] px-4 pb-12">
        <Block title="Best sellers" endpoint="/api/home/best-sellers" more="/store?sort=popular" />
        <Block title="Hottest deals" endpoint="/api/home/deals" more="/store?sort=discount&onSale=1" />
        <Block title="New releases" endpoint="/api/home/new-releases" more="/store?sort=newest" />
      </div>
    </>
  )
}
