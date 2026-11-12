const BASE = import.meta.env.VITE_API_URL ?? ''

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

export const tokenStore = {
  get: () => localStorage.getItem('token'),
  set: (t: string) => localStorage.setItem('token', t),
  clear: () => localStorage.removeItem('token'),
}

export async function api<T>(path: string, opts: { method?: string; body?: unknown } = {}): Promise<T> {
  const token = tokenStore.get()
  const res = await fetch(BASE + path, {
    method: opts.method ?? (opts.body ? 'POST' : 'GET'),
    headers: { ...(opts.body ? { 'Content-Type': 'application/json' } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: opts.body ? JSON.stringify(opts.body) : undefined,
  })
  const data = await res.json().catch(() => ({}))
  if (!res.ok) throw new ApiError(res.status, data.error ?? 'Something went wrong')
  return data as T
}

export const coverSrc = (url: string) => BASE + url

export const money = (cents: number) => `€${(cents / 100).toFixed(2)}`

export interface Card {
  id: number; slug: string; name: string; platform: string; region: string; genre: string
  coverUrl: string; priceCents: number; listPriceCents: number; discountPercent: number
  offerCount: number; releaseYear: number
}
export interface Offer { id: number; priceCents: number; stock: number; seller: string; sellerRating: number; sellerSales: number }
export interface ProductDetail extends Omit<Card, 'offerCount'> { description: string; offers: Offer[] }
export interface Banner { id: number; title: string; subtitle: string; cta: string; link: string; colorA: string; colorB: string }
export interface Facet { value: string; count: number }
export interface Filters { platforms: Facet[]; regions: Facet[]; genres: Facet[]; minPriceCents: number; maxPriceCents: number }
export interface User { id: number; email: string; name: string }
export interface QuoteLine { offerId: number; qty: number; productId: number; slug: string; name: string; platform: string; region: string; coverUrl: string; seller: string; priceCents: number; stock: number; available: boolean }
export interface OrderItem { name: string; slug: string; platform: string; region: string; coverUrl: string; seller: string; priceCents: number; key: string }
export interface Order { id: number; totalCents: number; status: string; paymentRef: string; createdAt: string; items: OrderItem[] }
export interface OrderSummary { id: number; totalCents: number; status: string; createdAt: string; itemCount: number }
