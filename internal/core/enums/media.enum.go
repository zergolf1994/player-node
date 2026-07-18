package enums

// ─── Media Types ─────────────────────────────────────────────────────
// Must match MediaType in vdohide-service (file.enum.ts).

const (
	MediaTypeVideo     = "video"
	MediaTypeAudio     = "audio"
	MediaTypeSubtitle  = "subtitle"
	MediaTypeThumbnail = "thumbnail"
	MediaTypeImage     = "image"
	MediaTypeDocument  = "document"
	MediaTypeOther     = "other"
)

// ─── Resolution ──────────────────────────────────────────────────────

const (
	ResolutionOriginal = "original"
	Resolution1080     = "1080"
	Resolution720      = "720"
	Resolution480      = "480"
	Resolution360      = "360"
)

// ─── Image Resolution ────────────────────────────────────────────────

const (
	ResolutionPoster  = "poster"
	ResolutionGallery = "gallery"
)
