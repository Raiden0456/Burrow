package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

const (
	cookieName = "gonteractive_session"
	sessionTTL = 90 * 24 * time.Hour
)

type Auth struct {
	q            *db.Queries
	passwordHash [32]byte
	secure       *bool
}

// New creates the authenticator. A nil secure marks the session cookie Secure only for requests that
// arrived over HTTPS, so plain-HTTP LAN access keeps working.
func New(q *db.Queries, password string, secure *bool) *Auth {
	return &Auth{q: q, passwordHash: sha256.Sum256([]byte(password)), secure: secure}
}

func (a *Auth) Routes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/auth/login", httpx.Handle(a.login))
	mux.HandleFunc("POST /api/auth/logout", httpx.Handle(a.logout))
}

func (a *Auth) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err == nil {
			n, err := a.q.CountActiveSessions(r.Context(), db.CountActiveSessionsParams{
				TokenHash: hashToken(cookie.Value),
				ExpiresAt: time.Now().Unix(),
			})
			if err != nil {
				httpx.WriteError(w, r, err)
				return
			}
			if n > 0 {
				next.ServeHTTP(w, r)
				return
			}
		}
		httpx.JSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	})
}

func (a *Auth) login(w http.ResponseWriter, r *http.Request) error {
	var body struct {
		Password string `json:"password"`
	}
	if err := httpx.Decode(r, &body); err != nil {
		return err
	}
	given := sha256.Sum256([]byte(body.Password))
	if subtle.ConstantTimeCompare(given[:], a.passwordHash[:]) != 1 {
		time.Sleep(500 * time.Millisecond)
		return httpx.Coded(http.StatusUnauthorized, "wrong_password", "wrong password")
	}

	raw := make([]byte, 32)
	rand.Read(raw)
	token := base64.RawURLEncoding.EncodeToString(raw)
	expires := time.Now().Add(sessionTTL)

	ctx := r.Context()
	if err := a.q.DeleteExpiredSessions(ctx, time.Now().Unix()); err != nil {
		return err
	}
	if err := a.q.CreateSession(ctx, db.CreateSessionParams{TokenHash: hashToken(token), ExpiresAt: expires.Unix()}); err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		Expires:  expires,
		HttpOnly: true,
		Secure:   a.isSecure(r),
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (a *Auth) logout(w http.ResponseWriter, r *http.Request) error {
	if cookie, err := r.Cookie(cookieName); err == nil {
		if err := a.q.DeleteSession(r.Context(), hashToken(cookie.Value)); err != nil {
			return err
		}
	}
	http.SetCookie(w, &http.Cookie{Name: cookieName, Path: "/", MaxAge: -1, HttpOnly: true, Secure: a.isSecure(r)})
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (a *Auth) isSecure(r *http.Request) bool {
	if a.secure != nil {
		return *a.secure
	}
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
