package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"player-node/internal/services"
)

type EmbedData struct {
	Title        string
	PlayerConfig template.JS
}

// Embed handles GET /embed/{fileSlug} — Video player embed page
func (h *Handler) Embed(w http.ResponseWriter, r *http.Request) {
	if services.IsMaintenanceMode() {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("CDN-Cache-Control", "max-age=60")
		templates.ExecuteTemplate(w, "maintenance.html", nil)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, "/embed/")
	slug := strings.TrimSuffix(path, "/")
	if slug == "" {
		RenderError(w, "File not found", http.StatusNotFound)
		return
	}

	resolved, resolveErr := h.resolveEmbed(r, slug)
	if resolveErr != nil {
		if resolveErr.Processing != nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			w.Header().Set("CDN-Cache-Control", "max-age=60")
			if err := templates.ExecuteTemplate(w, "processing.html", resolveErr.Processing); err != nil {
				log.Printf("⚠️ Template error: %v", err)
				RenderError(w, "Processing...", http.StatusInternalServerError)
			}
			return
		}
		switch resolveErr.Status {
		case http.StatusForbidden:
			RenderError(w, resolveErr.Message, http.StatusForbidden)
		default:
			RenderError(w, resolveErr.Message, http.StatusNotFound)
		}
		return
	}

	configJSON, _ := json.Marshal(resolved.EmbedConfig)

	data := EmbedData{
		Title:        resolved.File.Name,
		PlayerConfig: template.JS(configJSON),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("CDN-Cache-Control", "max-age=14400")
	if err := templates.ExecuteTemplate(w, "embed.html", data); err != nil {
		log.Printf("⚠️ Template error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
