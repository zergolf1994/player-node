package handlers

import (
	"net/http"
	"strconv"

	"player-node/internal/assets"
)

// Favicon serves the embedded site icon.
func Favicon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "image/x-icon")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("CDN-Cache-Control", "max-age=604800")
	w.Header().Set("Content-Length", strconv.Itoa(len(assets.Favicon)))

	if r.Method == http.MethodHead {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Write(assets.Favicon)
}
