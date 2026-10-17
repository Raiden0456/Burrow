package wallpaper

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadServeRemove(t *testing.T) {
	mux := http.NewServeMux()
	New(t.TempDir()).Routes(mux)

	do := func(method, path string, body []byte) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, path, bytes.NewReader(body)))
		return rec
	}
	version := func() *int64 {
		var out struct{ Version *int64 }
		json.NewDecoder(do("GET", "/api/wallpaper", nil).Body).Decode(&out)
		return out.Version
	}

	if version() != nil {
		t.Fatal("want no wallpaper initially")
	}
	if rec := do("PUT", "/api/wallpaper/image", []byte("not an image")); rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("non-image upload: %d", rec.Code)
	}

	var img bytes.Buffer
	png.Encode(&img, image.NewRGBA(image.Rect(0, 0, 2, 2)))
	if rec := do("PUT", "/api/wallpaper/image", img.Bytes()); rec.Code != http.StatusOK {
		t.Fatalf("upload: %d %s", rec.Code, rec.Body)
	}
	if version() == nil {
		t.Fatal("want version after upload")
	}
	if rec := do("GET", "/api/wallpaper/image", nil); rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("serve: %d %s", rec.Code, rec.Header().Get("Content-Type"))
	}

	if rec := do("DELETE", "/api/wallpaper/image", nil); rec.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", rec.Code)
	}
	if rec := do("GET", "/api/wallpaper/image", nil); rec.Code != http.StatusNotFound {
		t.Fatalf("serve after delete: %d", rec.Code)
	}
}
