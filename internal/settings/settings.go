package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/raiden0456/gonteractive/internal/httpx"
	"github.com/raiden0456/gonteractive/internal/store/db"
)

const (
	preferencesKey  = "preferences"
	interfaceKey    = "interface"
	maxWallpaperDim = 90
	dateLayout      = "2006-01-02"
)

type Interface struct {
	Language     string `json:"language"`
	WallpaperDim int    `json:"wallpaperDim"`
}

func DefaultInterface() Interface {
	return Interface{WallpaperDim: 55}
}

func (ui Interface) Validate() error {
	if ui.Language != "en" && ui.Language != "ru" {
		return errors.New("language must be en or ru")
	}
	if ui.WallpaperDim < 0 || ui.WallpaperDim > maxWallpaperDim {
		return fmt.Errorf("wallpaperDim must be between 0 and %d", maxWallpaperDim)
	}
	return nil
}

type Preferences struct {
	Language         string `json:"language"`
	Profile          string `json:"profile"`
	DailyEnabled     bool   `json:"dailyEnabled"`
	DailyTime        string `json:"dailyTime"`
	DailySize        int    `json:"dailySize"`
	DailyPausedUntil string `json:"dailyPausedUntil"`
}

func DefaultPreferences() Preferences {
	return Preferences{
		Language:     "Russian",
		DailyEnabled: true,
		DailyTime:    "09:00",
		DailySize:    8,
	}
}

func (p Preferences) Validate() error {
	if strings.TrimSpace(p.Language) == "" {
		return errors.New("language is required")
	}
	if _, err := time.Parse("15:04", p.DailyTime); err != nil {
		return errors.New("dailyTime must be HH:MM")
	}
	if p.DailySize < 3 || p.DailySize > 30 {
		return errors.New("dailySize must be between 3 and 30")
	}
	if p.DailyPausedUntil != "" {
		if _, err := time.Parse(dateLayout, p.DailyPausedUntil); err != nil {
			return errors.New("dailyPausedUntil must be YYYY-MM-DD")
		}
	}
	return nil
}

func (p Preferences) DailyAt(day time.Time) time.Time {
	t, _ := time.Parse("15:04", p.DailyTime)
	y, m, d := day.Date()
	return time.Date(y, m, d, t.Hour(), t.Minute(), 0, 0, day.Location())
}

func (p Preferences) DailyPaused(now time.Time) bool {
	return p.DailyPausedUntil != "" && now.Format(dateLayout) <= p.DailyPausedUntil
}

type Store struct {
	q *db.Queries
}

func New(q *db.Queries) *Store {
	return &Store{q: q}
}

func (s *Store) Get(ctx context.Context, key string, out any) (bool, error) {
	raw, err := s.q.GetSetting(ctx, key)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if err := json.Unmarshal([]byte(raw), out); err != nil {
		return false, fmt.Errorf("decode setting %s: %w", key, err)
	}
	return true, nil
}

func (s *Store) Put(ctx context.Context, key string, v any) error {
	raw, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return s.q.PutSetting(ctx, db.PutSettingParams{Key: key, Value: string(raw)})
}

func (s *Store) Preferences(ctx context.Context) (Preferences, error) {
	p := DefaultPreferences()
	_, err := s.Get(ctx, preferencesKey, &p)
	return p, err
}

func (s *Store) Interface(ctx context.Context) (Interface, error) {
	ui := DefaultInterface()
	_, err := s.Get(ctx, interfaceKey, &ui)
	return ui, err
}

func (s *Store) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/settings/interface", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		ui, err := s.Interface(r.Context())
		if err != nil {
			return err
		}
		httpx.JSON(w, http.StatusOK, ui)
		return nil
	}))

	mux.HandleFunc("PUT /api/settings/interface", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var ui Interface
		if err := httpx.Decode(r, &ui); err != nil {
			return err
		}
		if err := ui.Validate(); err != nil {
			return httpx.Errorf(http.StatusBadRequest, "%v", err)
		}
		if err := s.Put(r.Context(), interfaceKey, ui); err != nil {
			return err
		}
		httpx.JSON(w, http.StatusOK, ui)
		return nil
	}))

	mux.HandleFunc("GET /api/settings/preferences", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		p, err := s.Preferences(r.Context())
		if err != nil {
			return err
		}
		httpx.JSON(w, http.StatusOK, p)
		return nil
	}))

	mux.HandleFunc("PUT /api/settings/preferences", httpx.Handle(func(w http.ResponseWriter, r *http.Request) error {
		var p Preferences
		if err := httpx.Decode(r, &p); err != nil {
			return err
		}
		if err := p.Validate(); err != nil {
			return httpx.Errorf(http.StatusBadRequest, "%v", err)
		}
		if err := s.Put(r.Context(), preferencesKey, p); err != nil {
			return err
		}
		httpx.JSON(w, http.StatusOK, p)
		return nil
	}))
}
