package enums

// ─── File Types ──────────────────────────────────────────────────────
// Must match FileType in vdohide-service (file.enum.ts).

const (
	FileTypeFolder = "folder"
	FileTypeVideo  = "video"
	FileTypeImage  = "image"
	FileTypeOther  = "other"
)

// ─── File Statuses ───────────────────────────────────────────────────
// Must match FileStatus in vdohide-service (file.enum.ts).

const (
	FileStatusWaiting       = "waiting"
	FileStatusProcessing    = "processing"
	FileStatusReady         = "ready"
	FileStatusReadyOriginal = "ready_original"
	FileStatusError         = "error"
	FileStatusQueue         = "queue"
)
