package enums

// ─── Domain Statuses ─────────────────────────────────────────────────
// Must match DomainStatus in vdohide-service (domain.enum.ts).

const (
	DomainStatusPending = "pending"
	DomainStatusActive  = "active"
	DomainStatusFailed  = "failed"
	DomainStatusExpired = "expired"
)

// ─── Ads Image Show On ───────────────────────────────────────────────
// Must match AdsImageShowOn in vdohide-service (domain.enum.ts).

const (
	AdsImageShowOnReady = "ready"
	AdsImageShowOnEnd   = "end"
	AdsImageShowOnPause = "pause"
)
