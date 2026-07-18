package handlers

import (
	"fmt"
	"net/http"
)

// RenderError renders an HTML error page using the 404 template
func RenderError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("CDN-Cache-Control", "max-age=60")
	w.WriteHeader(statusCode)
	if templates != nil {
		templates.ExecuteTemplate(w, "404.html", map[string]string{
			"Message": message,
		})
		return
	}
	// Fallback if templates not loaded
	fmt.Fprintf(w, "<html><body><h1>%d</h1><p>%s</p></body></html>", statusCode, message)
}
