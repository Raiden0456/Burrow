package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

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

func TestDailyPaused(t *testing.T) {
	now := time.Date(2026, 10, 26, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		until string
		want  bool
	}{
		{"", false},
		{"2026-10-25", false},
		{"2026-10-26", true},
		{"2026-11-02", true},
	}
	for _, tt := range tests {
		if got := (Preferences{DailyPausedUntil: tt.until}).DailyPaused(now); got != tt.want {
			t.Errorf("until %q: got %v, want %v", tt.until, got, tt.want)
		}
	}
}
