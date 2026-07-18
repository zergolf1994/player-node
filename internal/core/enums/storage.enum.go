package enums

// ─── Storage Types ───────────────────────────────────────────────────
// Must match StorageType in vdohide-service (storage.enum.ts).

const (
	StorageTypeLocal = "local"
	StorageTypeS3    = "s3"
)

// ─── Storage Statuses ────────────────────────────────────────────────
// Must match StorageStatus in vdohide-service (storage.enum.ts).

const (
	StorageStatusOnline      = "online"
	StorageStatusOffline     = "offline"
	StorageStatusError       = "error"
	StorageStatusMaintenance = "maintenance"
)

// ─── Storage Accepts ─────────────────────────────────────────────────
// Must match StorageAccept in vdohide-service (storage.enum.ts).

const (
	StorageAcceptUpload  = "upload"
	StorageAcceptTemp    = "temp"
	StorageAcceptStorage = "storage"
	StorageAcceptVideo   = "video"
	StorageAcceptImage   = "image"
	StorageAcceptOther   = "other"
)
