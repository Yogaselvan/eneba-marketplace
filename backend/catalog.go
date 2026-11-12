package main

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type ProductCard struct {
	ID              int64  `json:"id"`
	Slug            string `json:"slug"`
	Name            string `json:"name"`
	Platform        string `json:"platform"`
	Region          string `json:"region"`
	Genre           string `json:"genre"`
	CoverURL        string `json:"coverUrl"`
	PriceCents      int    `json:"priceCents"`
	ListPriceCents  int    `json:"listPriceCents"`
	DiscountPercent int    `json:"discountPercent"`
	OfferCount      int    `json:"offerCount"`
	ReleaseYear     int    `json:"releaseYear"`
}

// cardSelect joins each product with the aggregate of its in-stock offers.
// Products with no stock are excluded by the join condition.
const cardSelect = `
SELECT p.id, p.slug, p.name, p.platform, p.region, p.genre, p.list_price_cents, p.release_year,
       o.min_price, o.cnt
FROM products p
JOIN LATERAL (
  SELECT min(price_cents) AS min_price, count(*) AS cnt
  FROM offers WHERE product_id = p.id AND stock > 0
) o ON o.cnt > 0`

func scanCard(rows pgx.Rows) (ProductCard, error) {
	var c ProductCard
	err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.Platform, &c.Region, &c.Genre, &c.ListPriceCents, &c.ReleaseYear, &c.PriceCents, &c.OfferCount)
	c.CoverURL = "/api/cover/" + c.Slug
	if c.ListPriceCents > 0 && c.PriceCents < c.ListPriceCents {
		c.DiscountPercent = 100 - c.PriceCents*100/c.ListPriceCents
	}
	return c, err
}

func (a *App) queryCards(w http.ResponseWriter, r *http.Request, q string, args ...any) {
	rows, err := a.db.Query(r.Context(), q, args...)
	if err != nil {
		writeErr(w, 500, "query failed")
		return
	}
	defer rows.Close()
	out := []ProductCard{}
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			writeErr(w, 500, "scan failed")
			return
		}
		out = append(out, c)
	}
	writeJSON(w, 200, out)
}

// ---- Homepage blocks: each is an independent endpoint ----

func (a *App) homeBanners(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), `SELECT id,title,subtitle,cta,link,color_a,color_b FROM banners ORDER BY position`)
	if err != nil {
		writeErr(w, 500, "query failed")
		return
	}
	defer rows.Close()
	type B struct {
		ID                                         int64
		Title, Subtitle, Cta, Link, ColorA, ColorB string
	}
	out := []map[string]any{}
	for rows.Next() {
		var b B
		if rows.Scan(&b.ID, &b.Title, &b.Subtitle, &b.Cta, &b.Link, &b.ColorA, &b.ColorB) == nil {
			out = append(out, map[string]any{"id": b.ID, "title": b.Title, "subtitle": b.Subtitle, "cta": b.Cta, "link": b.Link, "colorA": b.ColorA, "colorB": b.ColorB})
		}
	}
	writeJSON(w, 200, out)
}

func (a *App) homeBestSellers(w http.ResponseWriter, r *http.Request) {
	a.queryCards(w, r, cardSelect+` ORDER BY p.sales DESC LIMIT 12`)
}

func (a *App) homeDeals(w http.ResponseWriter, r *http.Request) {
	a.queryCards(w, r, cardSelect+` ORDER BY (1 - o.min_price::float / p.list_price_cents) DESC LIMIT 12`)
}

func (a *App) homeNewReleases(w http.ResponseWriter, r *http.Request) {
	a.queryCards(w, r, cardSelect+` ORDER BY p.release_year DESC, p.sales DESC LIMIT 12`)
}

// ---- Store: filters, sorting, search, pagination ----

// searchClause matches substrings, then falls back to trigram similarity so typos
// ("cyberpnk") and partial names ("witcher 3") still return results.
const searchClause = `(p.name ILIKE '%%' || $%[1]d || '%%' OR similarity(p.name, $%[1]d) > 0.2 OR word_similarity($%[1]d, p.name) > 0.45)`

func (a *App) listProducts(w http.ResponseWriter, r *http.Request) {
	qs := r.URL.Query()
	where := []string{}
	args := []any{}
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}
	q := strings.TrimSpace(qs.Get("q"))
	if q != "" {
		add(searchClause, q)
	}
	if v := qs.Get("platform"); v != "" {
		add("p.platform = ANY(string_to_array($%d, ','))", v)
	}
	if v := qs.Get("region"); v != "" {
		add("p.region = ANY(string_to_array($%d, ','))", v)
	}
	if v := qs.Get("genre"); v != "" {
		add("p.genre = ANY(string_to_array($%d, ','))", v)
	}
	if v, err := strconv.Atoi(qs.Get("minPrice")); err == nil {
		add("o.min_price >= $%d", v)
	}
	if v, err := strconv.Atoi(qs.Get("maxPrice")); err == nil {
		add("o.min_price <= $%d", v)
	}
	if qs.Get("onSale") == "1" {
		where = append(where, "o.min_price < p.list_price_cents")
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	order := "p.sales DESC"
	switch qs.Get("sort") {
	case "price_asc":
		order = "o.min_price ASC, p.sales DESC"
	case "price_desc":
		order = "o.min_price DESC, p.sales DESC"
	case "discount":
		order = "(1 - o.min_price::float / p.list_price_cents) DESC"
	case "newest":
		order = "p.release_year DESC, p.sales DESC"
	case "name":
		order = "p.name ASC"
	case "relevance":
		if q != "" {
			order = fmt.Sprintf("(p.name ILIKE $%d || '%%')::int DESC, greatest(similarity(p.name,$%d), word_similarity($%d,p.name)) DESC, p.sales DESC", 1, 1, 1)
		}
	}
	// when searching, default to relevance (q is always $1 because it is added first)
	if q != "" && qs.Get("sort") == "" {
		order = "(p.name ILIKE $1 || '%') DESC, greatest(similarity(p.name,$1), word_similarity($1,p.name)) DESC, p.sales DESC"
	}

	page, _ := strconv.Atoi(qs.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(qs.Get("limit"))
	if limit < 1 || limit > 60 {
		limit = 24
	}

	var total int
	countSQL := `SELECT count(*) FROM products p JOIN LATERAL (SELECT min(price_cents) AS min_price, count(*) AS cnt FROM offers WHERE product_id=p.id AND stock>0) o ON o.cnt>0` + whereSQL
	if err := a.db.QueryRow(r.Context(), countSQL, args...).Scan(&total); err != nil {
		writeErr(w, 500, "count failed")
		return
	}
	args = append(args, limit, (page-1)*limit)
	sql := cardSelect + whereSQL + " ORDER BY " + order + fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)-1, len(args))
	rows, err := a.db.Query(r.Context(), sql, args...)
	if err != nil {
		writeErr(w, 500, "query failed")
		return
	}
	defer rows.Close()
	items := []ProductCard{}
	for rows.Next() {
		c, err := scanCard(rows)
		if err != nil {
			writeErr(w, 500, "scan failed")
			return
		}
		items = append(items, c)
	}
	writeJSON(w, 200, map[string]any{"items": items, "total": total, "page": page, "pageSize": limit})
}

func (a *App) suggest(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if len(q) < 2 {
		writeJSON(w, 200, []ProductCard{})
		return
	}
	a.queryCards(w, r, cardSelect+" WHERE "+fmt.Sprintf(searchClause, 1)+
		` ORDER BY (p.name ILIKE $1 || '%') DESC, greatest(similarity(p.name,$1), word_similarity($1,p.name)) DESC, p.sales DESC LIMIT 6`, q)
}

func (a *App) filters(w http.ResponseWriter, r *http.Request) {
	type facet struct {
		Value string `json:"value"`
		Count int    `json:"count"`
	}
	facetQ := func(col string) []facet {
		rows, err := a.db.Query(r.Context(), `SELECT `+col+`, count(*) FROM products p WHERE EXISTS (SELECT 1 FROM offers WHERE product_id=p.id AND stock>0) GROUP BY 1 ORDER BY 2 DESC`)
		out := []facet{}
		if err != nil {
			return out
		}
		defer rows.Close()
		for rows.Next() {
			var f facet
			if rows.Scan(&f.Value, &f.Count) == nil {
				out = append(out, f)
			}
		}
		return out
	}
	var lo, hi int
	_ = a.db.QueryRow(r.Context(), `SELECT coalesce(min(price_cents),0), coalesce(max(price_cents),0) FROM offers WHERE stock>0`).Scan(&lo, &hi)
	writeJSON(w, 200, map[string]any{
		"platforms": facetQ("platform"), "regions": facetQ("region"), "genres": facetQ("genre"),
		"minPriceCents": lo, "maxPriceCents": hi,
	})
}

// ---- Product page ----

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var p struct {
		ID                                  int64  `json:"id"`
		Slug, Name, Platform, Region, Genre string `json:"-"`
		Description                         string `json:"-"`
		ListPrice, Year                     int    `json:"-"`
	}
	err := a.db.QueryRow(r.Context(), `SELECT id,slug,name,platform,region,genre,description,list_price_cents,release_year FROM products WHERE slug=$1`, slug).
		Scan(&p.ID, &p.Slug, &p.Name, &p.Platform, &p.Region, &p.Genre, &p.Description, &p.ListPrice, &p.Year)
	if err != nil {
		writeErr(w, 404, "product not found")
		return
	}
	rows, err := a.db.Query(r.Context(), `
		SELECT o.id, o.price_cents, o.stock, s.name, s.rating, s.sales_count
		FROM offers o JOIN sellers s ON s.id=o.seller_id
		WHERE o.product_id=$1 AND o.stock>0 ORDER BY o.price_cents ASC, s.rating DESC`, p.ID)
	if err != nil {
		writeErr(w, 500, "query failed")
		return
	}
	defer rows.Close()
	offers := []map[string]any{}
	for rows.Next() {
		var id int64
		var price, stock, sales int
		var seller string
		var rating float64
		if rows.Scan(&id, &price, &stock, &seller, &rating, &sales) == nil {
			offers = append(offers, map[string]any{"id": id, "priceCents": price, "stock": stock, "seller": seller, "sellerRating": rating, "sellerSales": sales})
		}
	}
	best := 0
	if len(offers) > 0 {
		best = offers[0]["priceCents"].(int)
	}
	disc := 0
	if p.ListPrice > 0 && best > 0 && best < p.ListPrice {
		disc = 100 - best*100/p.ListPrice
	}
	writeJSON(w, 200, map[string]any{
		"id": p.ID, "slug": p.Slug, "name": p.Name, "platform": p.Platform, "region": p.Region, "genre": p.Genre,
		"description": p.Description, "releaseYear": p.Year, "coverUrl": "/api/cover/" + p.Slug,
		"listPriceCents": p.ListPrice, "priceCents": best, "discountPercent": disc, "offers": offers,
	})
}

// ---- Generated cover art (no external assets needed) ----

func (a *App) cover(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	var name, platform string
	if err := a.db.QueryRow(r.Context(), `SELECT name, platform FROM products WHERE slug=$1`, slug).Scan(&name, &platform); err != nil {
		http.NotFound(w, r)
		return
	}
	h := fnv.New32a()
	h.Write([]byte(strings.SplitN(name, ":", 2)[0]))
	hue := int(h.Sum32() % 360)
	words := strings.Fields(strings.SplitN(name, ":", 2)[0])
	lines := []string{}
	cur := ""
	for _, wd := range words {
		if len(cur)+len(wd) > 11 && cur != "" {
			lines = append(lines, cur)
			cur = wd
		} else {
			cur = strings.TrimSpace(cur + " " + wd)
		}
	}
	if cur != "" {
		lines = append(lines, cur)
	}
	if len(lines) > 4 {
		lines = lines[:4]
	}
	longest := 1
	for _, l := range lines {
		if len(l) > longest {
			longest = len(l)
		}
	}
	fs := 34
	if fit := int(270 / (float64(longest) * 0.9)); fit < fs {
		fs = fit
	}
	var text strings.Builder
	startY := 190 - (len(lines)-1)*22
	for i, l := range lines {
		text.WriteString(fmt.Sprintf(`<text x="150" y="%d" text-anchor="middle" font-family="Arial Black,Arial,sans-serif" font-weight="900" font-size="%d" fill="#fff">%s</text>`, startY+i*(fs+10), fs, escapeXML(strings.ToUpper(l))))
	}
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 400"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="hsl(%d,70%%,42%%)"/><stop offset="1" stop-color="hsl(%d,75%%,16%%)"/></linearGradient></defs><rect width="300" height="400" fill="url(#g)"/><circle cx="260" cy="40" r="120" fill="#fff" opacity=".07"/><circle cx="20" cy="380" r="90" fill="#000" opacity=".15"/>%s<text x="150" y="376" text-anchor="middle" font-family="Arial,sans-serif" font-size="14" letter-spacing="3" fill="#fff" opacity=".8">%s</text></svg>`,
		hue, (hue+40)%360, text.String(), escapeXML(strings.ToUpper(platform)))
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write([]byte(svg))
}

func escapeXML(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;").Replace(s)
}
