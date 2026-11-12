import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'

export interface CartItem { offerId: number; qty: number }
interface CartCtx {
  items: CartItem[]
  count: number
  add: (offerId: number) => void
  setQty: (offerId: number, qty: number) => void
  remove: (offerId: number) => void
  clear: () => void
}
const Ctx = createContext<CartCtx>(null!)
export const useCart = () => useContext(Ctx)
const MAX_QTY = 5

// The cart lives in localStorage so guests can shop without an account and keep
// their cart through sign-in. Prices are never stored: the server re-quotes every
// line (POST /api/cart/quote) and again at checkout.
export function CartProvider({ children }: { children: ReactNode }) {
  const [items, setItems] = useState<CartItem[]>(() => {
    try { return JSON.parse(localStorage.getItem('cart') ?? '[]') } catch { return [] }
  })
  useEffect(() => localStorage.setItem('cart', JSON.stringify(items)), [items])

  const value = useMemo<CartCtx>(() => ({
    items,
    count: items.reduce((n, i) => n + i.qty, 0),
    add: (offerId) => setItems((cur) => {
      const found = cur.find((i) => i.offerId === offerId)
      return found ? cur.map((i) => i.offerId === offerId ? { ...i, qty: Math.min(MAX_QTY, i.qty + 1) } : i) : [...cur, { offerId, qty: 1 }]
    }),
    setQty: (offerId, qty) => setItems((cur) => cur.map((i) => i.offerId === offerId ? { ...i, qty: Math.max(1, Math.min(MAX_QTY, qty)) } : i)),
    remove: (offerId) => setItems((cur) => cur.filter((i) => i.offerId !== offerId)),
    clear: () => setItems([]),
  }), [items])

  return <Ctx.Provider value={value}>{children}</Ctx.Provider>
}
