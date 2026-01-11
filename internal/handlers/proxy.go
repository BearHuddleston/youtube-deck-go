package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var cacheDir = "cache/images"

func init() {
	_ = os.MkdirAll(cacheDir, 0755)
}

func (h *Handlers) HandleImageProxy(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		http.Error(w, "missing url", http.StatusBadRequest)
		return
	}

	if !strings.HasPrefix(url, "https://yt") && !strings.HasPrefix(url, "https://i.ytimg.com") {
		http.Error(w, "invalid url", http.StatusBadRequest)
		return
	}

	hash := sha256.Sum256([]byte(url))
	filename := hex.EncodeToString(hash[:]) + ".jpg"
	cachePath := filepath.Join(cacheDir, filename)

	if data, err := os.ReadFile(cachePath); err == nil {
		w.Header().Set("Content-Type", "image/jpeg")
		w.Header().Set("Cache-Control", "public, max-age=604800")
		_, _ = w.Write(data)
		return
	}

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "fetch failed", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		http.Error(w, "upstream error", resp.StatusCode)
		return
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "read failed", http.StatusBadGateway)
		return
	}

	_ = os.WriteFile(cachePath, data, 0644)

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "public, max-age=604800")
	_, _ = w.Write(data)
}
