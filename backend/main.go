package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

type App struct {
	db        *pgxpool.Pool
	jwtSecret []byte
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func main() {
	ctx := context.Background()
	dsn := env("DATABASE_URL", "postgres://localhost:5432/eneba_clone?sslmode=disable")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("database unreachable: %v", err)
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	app := &App{db: pool, jwtSecret: []byte(env("JWT_SECRET", "dev-secret-change-me"))}

	// `go run . seed` (re)populates the catalogue; the server also seeds on first boot.
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		if err := app.seed(ctx, true); err != nil {
			log.Fatal(err)
		}
		log.Println("seeded")
		return
	}
	if err := app.seed(ctx, false); err != nil {
		log.Fatal(err)
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP, middleware.Logger, middleware.Recoverer, middleware.Timeout(20*time.Second))
	origins := strings.Split(env("CORS_ORIGINS", "*"), ",")
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: origins,
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders: []string{"Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]string{"status": "ok"}) })
	r.Route("/api", func(r chi.Router) {
		r.Get("/home/banners", app.homeBanners)
		r.Get("/home/best-sellers", app.homeBestSellers)
		r.Get("/home/deals", app.homeDeals)
		r.Get("/home/new-releases", app.homeNewReleases)

		r.Get("/products", app.listProducts)
		r.Get("/products/{slug}", app.getProduct)
		r.Get("/filters", app.filters)
		r.Get("/suggest", app.suggest)
		r.Get("/cover/{slug}", app.cover)

		r.Post("/auth/register", app.register)
		r.Post("/auth/login", app.login)
		r.Post("/cart/quote", app.quote)

		r.Group(func(r chi.Router) {
			r.Use(app.requireAuth)
			r.Get("/me", app.me)
			r.Post("/checkout", app.checkout)
			r.Get("/orders", app.listOrders)
			r.Get("/orders/{id}", app.getOrder)
		})
	})

	addr := ":" + env("PORT", "8080")
	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, r))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(r *http.Request, v any) bool {
	r.Body = http.MaxBytesReader(nil, r.Body, 1<<20)
	return json.NewDecoder(r.Body).Decode(v) == nil
}
