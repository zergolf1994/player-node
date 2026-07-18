package handlers

import (
	"embed"
	"html/template"
	"log"
	"net/http"

	"player-node/internal/services"
)

//go:embed templates/*.html
var templatesFS embed.FS

var templates *template.Template

// Handler holds dependencies for HTTP handlers
type Handler struct{}

// NewHandler creates a new Handler instance
func NewHandler(h Handler) *Handler {
	return &h
}

// InitTemplates loads HTML templates from embedded FS
func InitTemplates() error {
	var err error
	templates, err = template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return err
	}
	log.Printf("✅ Templates loaded from embedded filesystem")
	return nil
}

// ─── Domain/Space Validation ──────────────────────────────────────────────────

// CheckDomainSpace validates if the request domain allows accessing the target space.
func CheckDomainSpace(r *http.Request, targetSpaceID *string) bool {
	domain, isDomainRequest := services.FindDomain(r.Host)
	if isDomainRequest {
		if domain == nil || domain.Status != "active" || !domain.Enable {
			return false
		}

		hasSpace := domain.SpaceID != nil && *domain.SpaceID != ""
		hasCreator := domain.CreatorID != nil && *domain.CreatorID != ""

		if !hasSpace && !hasCreator {
			return true
		}

		if hasSpace {
			tSpace := ""
			if targetSpaceID != nil {
				tSpace = *targetSpaceID
			}
			if tSpace != *domain.SpaceID {
				return false
			}
			return true
		}

		return false
	}
	return true
}

// ─── Router ───────────────────────────────────────────────────────────────────

// Home — player-node เสิร์ฟเฉพาะตัว player (embed/playlist json/advert json)
// content จริง (stream/รูป/sprite/m3u8) เป็นหน้าที่ของ content-node ทั้งหมด
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	HandleNotFound(w, r)
}
