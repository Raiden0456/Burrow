package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/raiden0456/gonteractive/internal/store"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

func TestInterfaceRoundTrip(t *testing.T) {
	conn, err := store.Open(t.Context(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	mux := http.NewServeMux()
	New(db.New(conn)).Routes(mux)

	do := func(method, body string) (int, Interface) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, "/api/settings/interface", bytes.NewBufferString(body)))
		var ui Interface
		json.NewDecoder(rec.Body).Decode(&ui)
		return rec.Code, ui
	}

	if _, ui := do("GET", ""); ui != DefaultInterface() {
		t.Fatalf("default = %+v", ui)
	}
	if code, _ := do("PUT", `{"language":"de","wallpaperDim":10}`); code != http.StatusBadRequest {
		t.Fatalf("bad language accepted: %d", code)
	}
	if code, _ := do("PUT", `{"language":"en","wallpaperDim":95}`); code != http.StatusBadRequest {
		t.Fatalf("dim above max accepted: %d", code)
	}
	if code, _ := do("PUT", `{"language":"en","wallpaperDim":30}`); code != http.StatusOK {
		t.Fatalf("put: %d", code)
	}
	if _, ui := do("GET", ""); ui.Language != "en" || ui.WallpaperDim != 30 {
		t.Fatalf("after put = %+v", ui)
	}
}
