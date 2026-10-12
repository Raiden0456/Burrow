package auth

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/raiden0456/gonteractive/internal/store"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

func TestLoginGuardsAPI(t *testing.T) {
	conn, err := store.Open(t.Context(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	a := New(db.New(conn), "secret", nil)
	mux := http.NewServeMux()
	a.Routes(mux)
	mux.Handle("/api/private", a.Require(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	get := func(cookie *http.Cookie) int {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/api/private", nil)
		if cookie != nil {
			req.AddCookie(cookie)
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res.StatusCode
	}
	login := func(password string) *http.Response {
		res, err := http.Post(srv.URL+"/api/auth/login", "application/json", strings.NewReader(`{"password":"`+password+`"}`))
		if err != nil {
			t.Fatal(err)
		}
		res.Body.Close()
		return res
	}

	if got := get(nil); got != http.StatusUnauthorized {
		t.Fatalf("anonymous request: got %d", got)
	}
	if res := login("wrong"); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password: got %d", res.StatusCode)
	}
	res := login("secret")
	if res.StatusCode != http.StatusNoContent || len(res.Cookies()) != 1 {
		t.Fatalf("login: got %d with %d cookies", res.StatusCode, len(res.Cookies()))
	}
	if got := get(res.Cookies()[0]); got != http.StatusTeapot {
		t.Fatalf("authenticated request: got %d", got)
	}
	if got := get(&http.Cookie{Name: cookieName, Value: "forged"}); got != http.StatusUnauthorized {
		t.Fatalf("forged cookie: got %d", got)
	}
}
