# Marketplace — Eneba-style game key store

A full-stack marketplace where several sellers compete on price for the same digital game key.
Built for the Eneba Intern Assignment 2027.

- **Frontend:** React 19 + TypeScript + Vite, React Router, TanStack Query, Tailwind v4
- **Backend:** Go (chi, pgx), REST
- **Database:** PostgreSQL (uses `pg_trgm` for typo-tolerant search)
- **Look & feel:** colours, type and component shapes sampled from the live eneba.com computed styles
  (`#4618AC` page, `#351183` panels, `#1F0A4D` inputs, `#FAD318` CTA with black 800-weight text, square corners).

## Run locally

```bash
createdb eneba_clone
cd backend && go run .            # creates the schema, seeds ~120 products on first boot, serves :8080
cd frontend && npm i && npm run dev   # http://localhost:5173 (Vite proxies /api to :8080)
```

Env vars (backend): `DATABASE_URL`, `JWT_SECRET`, `CORS_ORIGINS`, `PORT`. Frontend: `VITE_API_URL`.
Tests: `cd backend && go test ./...`. Reseed: `go run . seed`.

## What is built (and why)

The brief says depth beats breadth, so every feature in the purchase funnel works end to end:
homepage/search → product → cart → sign-in → pay → order + key.

| Area | Decision |
|---|---|
| **Homepage** | Banners, best sellers, deals and new releases each have their **own endpoint** (`/api/home/*`) and own loading/error state. The frontend hardcodes no content. |
| **Store** | Server-side filters (platform, region, genre, price range, on-sale), 6 sort orders, pagination, facet counts from `/api/filters`. Filter state lives in the URL, so views are shareable and the back button works. |
| **Search** | Header search with debounced autocomplete and keyboard navigation, available on every page. Matching = `ILIKE` substring **or** trigram `similarity` **or** `word_similarity`, so `witcer`, `cyberpnk` and `witcher 3` all work. Results rank prefix matches first, then similarity. |
| **Product page** | Product details plus every seller's offer, cheapest first, with rating, stock and a "best price" tag. |
| **Cart** | Held in `localStorage` as `{offerId, qty}` only, so guests can shop and keep the cart through sign-in. **Prices are never stored client-side**: `/api/cart/quote` re-prices every line from the DB. |
| **Checkout** | Requires sign-in (cart survives the redirect). One DB transaction locks the offer rows, decrements stock, creates the order and claims one key per unit with `SELECT … FOR UPDATE SKIP LOCKED`, so two buyers can never receive the same key and an out-of-stock race fails cleanly instead of overselling. |
| **Payment** | Simulated provider in test-mode style: `4242 4242 4242 4242` succeeds, `4000 0000 0000 0002` is declined. Payment is validated before stock is touched. |
| **Auth** | bcrypt passwords, HS256 JWT (7 days) in the `Authorization` header. Orders are always queried with `user_id = <token user>`, so one buyer cannot read another's keys. |
| **Covers** | Generated as SVG by the API (`/api/cover/{slug}`), so there are no copyrighted images and no asset hosting. |

## Trade-offs and what I'd do next

- JWT in `localStorage` is simple and works across Vercel↔Render origins, but is exposed to XSS. In production I'd use an `HttpOnly; SameSite` cookie behind a shared parent domain.
- The cart is client-side; a server-side cart would enable cross-device carts and abandoned-cart features.
- No seller dashboard, reviews, wishlist or real payment provider. Stripe test mode would slot in at `simulatePayment`.
- Search uses Postgres trigrams rather than a search engine: fine for thousands of products, and I'd move to Meilisearch/Elasticsearch beyond that.
- Keys are stored in plaintext for the demo; real keys should be encrypted at rest.
- Seed data is generated (`seed.go`, deterministic seed) from a list of real game titles with invented sellers and prices.

## Deploy (free tier)

1. **Database:** create a free Postgres on [Neon](https://neon.tech); copy the connection string.
2. **API on Render:** New → Blueprint → this repo (`render.yaml`). Set `DATABASE_URL` and `CORS_ORIGINS=https://<your-vercel-app>.vercel.app`. The API migrates and seeds itself on first boot.
3. **Frontend on Vercel:** import the repo, root directory `frontend`, env `VITE_API_URL=https://<your-render-service>.onrender.com`.
4. Free Render services sleep when idle; the first request can take ~30s.

## AI usage

The full agent conversation is in [`AI_LOG.md`](./AI_LOG.md).
