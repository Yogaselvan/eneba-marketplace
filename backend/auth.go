package main

import (
	"context"
	"net/http"
	"net/mail"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type ctxKey string

const userKey ctxKey = "uid"

func (a *App) issueToken(uid int64) (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": uid,
		"exp": time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	return t.SignedString(a.jwtSecret)
}

func (a *App) requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := r.Header.Get("Authorization")
		if !strings.HasPrefix(h, "Bearer ") {
			writeErr(w, 401, "sign in required")
			return
		}
		tok, err := jwt.Parse(strings.TrimPrefix(h, "Bearer "), func(t *jwt.Token) (any, error) { return a.jwtSecret, nil },
			jwt.WithValidMethods([]string{"HS256"}))
		if err != nil || !tok.Valid {
			writeErr(w, 401, "invalid or expired session")
			return
		}
		claims, _ := tok.Claims.(jwt.MapClaims)
		sub, ok := claims["sub"].(float64)
		if !ok {
			writeErr(w, 401, "invalid session")
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, int64(sub))))
	})
}

func uidFrom(r *http.Request) int64 { return r.Context().Value(userKey).(int64) }

type userDTO struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Name, Password string }
	if !decode(r, &in) {
		writeErr(w, 400, "invalid request")
		return
	}
	in.Email = strings.ToLower(strings.TrimSpace(in.Email))
	in.Name = strings.TrimSpace(in.Name)
	if _, err := mail.ParseAddress(in.Email); err != nil {
		writeErr(w, 422, "enter a valid email address")
		return
	}
	if len(in.Password) < 8 {
		writeErr(w, 422, "password must be at least 8 characters")
		return
	}
	if in.Name == "" {
		in.Name = strings.Split(in.Email, "@")[0]
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		writeErr(w, 500, "could not create account")
		return
	}
	u := userDTO{Email: in.Email, Name: in.Name}
	err = a.db.QueryRow(r.Context(), `INSERT INTO users(email,name,password_hash) VALUES($1,$2,$3) ON CONFLICT DO NOTHING RETURNING id`,
		in.Email, in.Name, string(hash)).Scan(&u.ID)
	if err != nil {
		writeErr(w, 409, "an account with this email already exists")
		return
	}
	a.respondSession(w, u)
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	var in struct{ Email, Password string }
	if !decode(r, &in) {
		writeErr(w, 400, "invalid request")
		return
	}
	var u userDTO
	var hash string
	err := a.db.QueryRow(r.Context(), `SELECT id,email,name,password_hash FROM users WHERE email=$1`, strings.ToLower(strings.TrimSpace(in.Email))).
		Scan(&u.ID, &u.Email, &u.Name, &hash)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)) != nil {
		writeErr(w, 401, "incorrect email or password")
		return
	}
	a.respondSession(w, u)
}

func (a *App) respondSession(w http.ResponseWriter, u userDTO) {
	tok, err := a.issueToken(u.ID)
	if err != nil {
		writeErr(w, 500, "could not start session")
		return
	}
	writeJSON(w, 200, map[string]any{"token": tok, "user": u})
}

func (a *App) me(w http.ResponseWriter, r *http.Request) {
	var u userDTO
	if err := a.db.QueryRow(r.Context(), `SELECT id,email,name FROM users WHERE id=$1`, uidFrom(r)).Scan(&u.ID, &u.Email, &u.Name); err != nil {
		writeErr(w, 401, "account not found")
		return
	}
	writeJSON(w, 200, u)
}
