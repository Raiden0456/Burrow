package wallpaper

import (
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/raiden0456/gonteractive/internal/httpx"
)

const maxSize = 15 << 20

type Service struct {
	path string
}

func New(dataDir string) *Service {
	return &Service{path: filepath.Join(dataDir, "wallpaper")}
}

func (s *Service) Routes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/wallpaper", httpx.Handle(s.info))
	mux.HandleFunc("GET /api/wallpaper/image", s.image)
	mux.HandleFunc("PUT /api/wallpaper/image", httpx.Handle(s.upload))
	mux.HandleFunc("DELETE /api/wallpaper/image", httpx.Handle(s.remove))
}

func (s *Service) info(w http.ResponseWriter, r *http.Request) error {
	var version *int64
	st, err := os.Stat(s.path)
	switch {
	case err == nil:
		v := st.ModTime().UnixMilli()
		version = &v
	case !errors.Is(err, fs.ErrNotExist):
		return err
	}
	httpx.JSON(w, http.StatusOK, map[string]*int64{"version": version})
	return nil
}

func (s *Service) image(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat(s.path); err != nil {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	http.ServeFile(w, r, s.path)
}

func (s *Service) upload(w http.ResponseWriter, r *http.Request) error {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxSize))
	if err != nil {
		return httpx.Coded(http.StatusRequestEntityTooLarge, "wallpaper_too_large", "image is larger than 15 MB")
	}
	if !strings.HasPrefix(http.DetectContentType(data), "image/") {
		return httpx.Coded(http.StatusUnsupportedMediaType, "wallpaper_not_image", "file is not an image")
	}

	tmp, err := os.CreateTemp(filepath.Dir(s.path), "wallpaper-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp.Name(), s.path); err != nil {
		return err
	}
	return s.info(w, r)
}

func (s *Service) remove(w http.ResponseWriter, r *http.Request) error {
	if err := os.Remove(s.path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
