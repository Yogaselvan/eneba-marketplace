package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"
)

type CartLine struct {
	OfferID int64 `json:"offerId"`
	Qty     int   `json:"qty"`
}

type QuoteLine struct {
	OfferID    int64  `json:"offerId"`
	Qty        int    `json:"qty"`
	ProductID  int64  `json:"productId"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Platform   string `json:"platform"`
	Region     string `json:"region"`
	CoverURL   string `json:"coverUrl"`
	Seller     string `json:"seller"`
	PriceCents int    `json:"priceCents"`
	Stock      int    `json:"stock"`
	Available  bool   `json:"available"`
}

const maxQty = 5

// quote re-prices a client-held cart from the database. The client never
// supplies prices, so a tampered localStorage cart cannot change what is charged.
func (a *App) quote(w http.ResponseWriter, r *http.Request) {
	var in struct{ Items []CartLine }
	if !decode(r, &in) {
		writeErr(w, 400, "invalid request")
		return
	}
	lines, total, err := a.priceCart(r, a.db, in.Items)
	if err != nil {
		writeErr(w, 500, "could not price cart")
		return
	}
	writeJSON(w, 200, map[string]any{"items": lines, "totalCents": total})
}

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (a *App) priceCart(r *http.Request, q querier, items []CartLine) ([]QuoteLine, int, error) {
	lines := []QuoteLine{}
	total := 0
	for _, it := range items {
		if it.Qty < 1 {
			continue
		}
		if it.Qty > maxQty {
			it.Qty = maxQty
		}
		rows, err := q.Query(r.Context(), `
			SELECT o.id, p.id, p.slug, p.name, p.platform, p.region, s.name, o.price_cents, o.stock
			FROM offers o JOIN products p ON p.id=o.product_id JOIN sellers s ON s.id=o.seller_id WHERE o.id=$1`, it.OfferID)
		if err != nil {
			return nil, 0, err
		}
		if rows.Next() {
			l := QuoteLine{Qty: it.Qty}
			if err := rows.Scan(&l.OfferID, &l.ProductID, &l.Slug, &l.Name, &l.Platform, &l.Region, &l.Seller, &l.PriceCents, &l.Stock); err == nil {
				l.CoverURL = "/api/cover/" + l.Slug
				l.Available = l.Stock >= l.Qty
				if l.Available {
					total += l.PriceCents * l.Qty
				}
				lines = append(lines, l)
			}
		}
		rows.Close()
	}
	return lines, total, nil
}

// checkout runs in a single transaction: lock stock, simulate payment, assign keys.
// Keys are claimed with FOR UPDATE SKIP LOCKED so concurrent buyers never receive the same key.
func (a *App) checkout(w http.ResponseWriter, r *http.Request) {
	var in struct {
		Items []CartLine
		Card  struct{ Number, Expiry, CVC, Name string }
	}
	if !decode(r, &in) || len(in.Items) == 0 {
		writeErr(w, 400, "your cart is empty")
		return
	}
	// merge duplicate offers
	merged := map[int64]int{}
	order := []int64{}
	for _, it := range in.Items {
		if it.Qty < 1 {
			continue
		}
		if _, ok := merged[it.OfferID]; !ok {
			order = append(order, it.OfferID)
		}
		merged[it.OfferID] += it.Qty
	}
	items := []CartLine{}
	for _, id := range order {
		qty := merged[id]
		if qty > maxQty {
			writeErr(w, 422, fmt.Sprintf("you can buy at most %d copies of one offer", maxQty))
			return
		}
		items = append(items, CartLine{id, qty})
	}

	ref, perr := simulatePayment(in.Card.Number, in.Card.Expiry, in.Card.CVC)
	if perr != "" {
		writeErr(w, 402, perr) // payment failed before touching stock
		return
	}

	tx, err := a.db.Begin(r.Context())
	if err != nil {
		writeErr(w, 500, "checkout failed")
		return
	}
	defer tx.Rollback(r.Context())

	total := 0
	type bought struct {
		offerID, productID int64
		price              int
	}
	var purchases []bought
	for _, it := range items {
		var productID int64
		var price, stock int
		// lock the offer row so concurrent checkouts of the same offer serialise on stock
		err := tx.QueryRow(r.Context(), `SELECT product_id, price_cents, stock FROM offers WHERE id=$1 FOR UPDATE`, it.OfferID).Scan(&productID, &price, &stock)
		if err != nil {
			writeErr(w, 422, "an item in your cart no longer exists")
			return
		}
		if stock < it.Qty {
			writeErr(w, 409, "an item in your cart is out of stock — please review your cart")
			return
		}
		for i := 0; i < it.Qty; i++ {
			purchases = append(purchases, bought{it.OfferID, productID, price})
			total += price
		}
		if _, err := tx.Exec(r.Context(), `UPDATE offers SET stock = stock - $2 WHERE id=$1`, it.OfferID, it.Qty); err != nil {
			writeErr(w, 500, "checkout failed")
			return
		}
	}

	var orderID int64
	if err := tx.QueryRow(r.Context(), `INSERT INTO orders(user_id,total_cents,status,payment_ref) VALUES($1,$2,'paid',$3) RETURNING id`, uidFrom(r), total, ref).Scan(&orderID); err != nil {
		writeErr(w, 500, "checkout failed")
		return
	}
	for _, p := range purchases {
		var itemID int64
		if err := tx.QueryRow(r.Context(), `INSERT INTO order_items(order_id,offer_id,product_id,price_cents) VALUES($1,$2,$3,$4) RETURNING id`, orderID, p.offerID, p.productID, p.price).Scan(&itemID); err != nil {
			writeErr(w, 500, "checkout failed")
			return
		}
		tag, err := tx.Exec(r.Context(), `
			UPDATE keys SET order_item_id=$2
			WHERE id = (SELECT id FROM keys WHERE offer_id=$1 AND order_item_id IS NULL ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED)`, p.offerID, itemID)
		if err != nil || tag.RowsAffected() != 1 {
			writeErr(w, 409, "an item ran out of stock while paying — you were not charged")
			return
		}
	}
	if err := tx.Commit(r.Context()); err != nil {
		writeErr(w, 500, "checkout failed")
		return
	}
	writeJSON(w, 201, map[string]any{"orderId": orderID})
}

// simulatePayment stands in for a payment provider's test mode.
// Magic numbers mimic Stripe's: 4242… succeeds, 4000 0000 0000 0002 is declined.
func simulatePayment(number, expiry, cvc string) (string, string) {
	n := strings.NewReplacer(" ", "", "-", "").Replace(number)
	if len(n) < 13 || len(n) > 19 || strings.Trim(n, "0123456789") != "" {
		return "", "enter a valid card number"
	}
	if len(strings.TrimSpace(cvc)) < 3 || strings.TrimSpace(expiry) == "" {
		return "", "enter the card expiry and CVC"
	}
	if n == "4000000000000002" {
		return "", "your card was declined — try test card 4242 4242 4242 4242"
	}
	return "sim_" + strconv.FormatInt(int64(len(n))*7919+int64(n[len(n)-1]), 36) + "_" + randomRef(), ""
}

type OrderItemDTO struct {
	Name       string `json:"name"`
	Slug       string `json:"slug"`
	Platform   string `json:"platform"`
	Region     string `json:"region"`
	CoverURL   string `json:"coverUrl"`
	Seller     string `json:"seller"`
	PriceCents int    `json:"priceCents"`
	Key        string `json:"key"`
}

func (a *App) listOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := a.db.Query(r.Context(), `
		SELECT o.id, o.total_cents, o.status, o.created_at, (SELECT count(*) FROM order_items WHERE order_id=o.id)
		FROM orders o WHERE o.user_id=$1 ORDER BY o.id DESC`, uidFrom(r))
	if err != nil {
		writeErr(w, 500, "query failed")
		return
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var id int64
		var total, n int
		var status string
		var at any
		if rows.Scan(&id, &total, &status, &at, &n) == nil {
			out = append(out, map[string]any{"id": id, "totalCents": total, "status": status, "createdAt": at, "itemCount": n})
		}
	}
	writeJSON(w, 200, out)
}

func (a *App) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeErr(w, 404, "order not found")
		return
	}
	var total int
	var status, ref string
	var at any
	// scoping by user_id means one buyer can never read another buyer's keys
	err = a.db.QueryRow(r.Context(), `SELECT total_cents,status,payment_ref,created_at FROM orders WHERE id=$1 AND user_id=$2`, id, uidFrom(r)).Scan(&total, &status, &ref, &at)
	if err != nil {
		writeErr(w, 404, "order not found")
		return
	}
	rows, err := a.db.Query(r.Context(), `
		SELECT p.name, p.slug, p.platform, p.region, s.name, oi.price_cents, coalesce(k.code,'')
		FROM order_items oi
		JOIN products p ON p.id=oi.product_id JOIN offers o ON o.id=oi.offer_id JOIN sellers s ON s.id=o.seller_id
		LEFT JOIN keys k ON k.order_item_id=oi.id
		WHERE oi.order_id=$1 ORDER BY oi.id`, id)
	if err != nil {
		writeErr(w, 500, "query failed")
		return
	}
	defer rows.Close()
	items := []OrderItemDTO{}
	for rows.Next() {
		var it OrderItemDTO
		if rows.Scan(&it.Name, &it.Slug, &it.Platform, &it.Region, &it.Seller, &it.PriceCents, &it.Key) == nil {
			it.CoverURL = "/api/cover/" + it.Slug
			items = append(items, it)
		}
	}
	writeJSON(w, 200, map[string]any{"id": id, "totalCents": total, "status": status, "paymentRef": ref, "createdAt": at, "items": items})
}
