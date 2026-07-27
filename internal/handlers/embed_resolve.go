package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"player-node/internal/cache"
	"player-node/internal/core/enums"
	"player-node/internal/db/models"
	"player-node/internal/services"

	"go.mongodb.org/mongo-driver/bson"
)

// ProcessingData is passed to processing.html when video is not ready.
type ProcessingData struct {
	State   string
	Message string
	Percent float64
}

// EmbedContent holds resolved media URLs for playlist feed.
type EmbedContent struct {
	PosterURL    string
	PlaylistM3U8 string
	SpriteVttURL string
}

// EmbedResolveResult is the shared embed / playlist feed resolution output.
type EmbedResolveResult struct {
	File        models.File
	Slug        string
	Content     EmbedContent
	EmbedConfig services.EmbedPlayerConfig
}

// EmbedResolveError describes a failed embed resolution.
type EmbedResolveError struct {
	Status     int
	Message    string
	Processing *ProcessingData
}

func requestHost(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-Host"); h != "" {
		return strings.TrimSpace(strings.Split(h, ",")[0])
	}
	return r.Host
}

func isLocalRequest(r *http.Request) bool {
	host := strings.ToLower(requestHost(r))
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	return host == "localhost" || host == "127.0.0.1" || host == "0.0.0.0"
}

func cfVisitorScheme(r *http.Request) string {
	raw := r.Header.Get("CF-Visitor")
	if raw == "" {
		return ""
	}
	var v struct {
		Scheme string `json:"scheme"`
	}
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		raw = strings.ToLower(raw)
		if strings.Contains(raw, `"scheme":"https"`) || strings.Contains(raw, `"scheme": "https"`) {
			return "https"
		}
		if strings.Contains(raw, `"scheme":"http"`) || strings.Contains(raw, `"scheme": "http"`) {
			return "http"
		}
		return ""
	}
	return strings.ToLower(strings.TrimSpace(v.Scheme))
}

func forwardedProto(r *http.Request) string {
	p := r.Header.Get("X-Forwarded-Proto")
	if p == "" {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(strings.Split(p, ",")[0]))
}

func requestProtocol(r *http.Request) string {
	if isLocalRequest(r) {
		return "http"
	}

	if scheme := cfVisitorScheme(r); scheme == "https" || scheme == "http" {
		return scheme
	}

	if p := forwardedProto(r); p == "https" || p == "http" {
		return p
	}

	// Cloudflare edge — visitor HTTPS even when origin connection is plain HTTP
	if r.Header.Get("CF-Ray") != "" {
		return "https"
	}

	if r.Header.Get("X-Forwarded-Ssl") == "on" ||
		r.Header.Get("X-Forwarded-Scheme") == "https" ||
		r.Header.Get("X-Url-Scheme") == "https" {
		return "https"
	}

	if r.TLS != nil {
		return "https"
	}

	return "http"
}

// embedLookup — ส่วนที่มาจาก DB ล้วนๆ (ไม่ขึ้นกับ request) เก็บลง Redis ได้
// cache เฉพาะไฟล์ที่ ready แล้ว — ไฟล์กำลัง processing ห้าม cache
// ไม่งั้นวิดีโอเสร็จแล้วแต่หน้า embed ยังค้างสถานะเดิมจนหมด TTL
type embedLookup struct {
	File         models.File `json:"file"`
	PosterURL    string      `json:"posterUrl"` // จาก storage publicUrl ("" = ใช้ thumb fallback)
	SpriteExists bool        `json:"spriteExists"`
}

// resolveEmbedData — งาน DB ทั้งหมดของ resolveEmbed (ผ่าน Redis lookup cache)
func resolveEmbedData(ctx context.Context, slug string) (*embedLookup, *EmbedResolveError) {
	cacheKey := "embed_resolve:" + slug
	var lk embedLookup
	if cache.GetJSON(cacheKey, &lk) {
		return &lk, nil
	}

	var file models.File
	err := models.FileModel.Col().FindOne(ctx, bson.M{"slug": slug}).Decode(&file)
	if err != nil {
		return nil, &EmbedResolveError{Status: http.StatusNotFound, Message: "File not found"}
	}

	if file.IsTrashed() || file.IsDeleted() {
		return nil, &EmbedResolveError{Status: http.StatusNotFound, Message: "File not found"}
	}

	cursor, err := models.MediaModel.Col().Find(ctx, bson.M{
		"fileId":     file.ID,
		"type":       enums.MediaTypeVideo,
		"resolution": bson.M{"$in": []string{"original", "1080", "720", "480", "360"}},
		"deletedAt":  bson.M{"$eq": nil},
	})
	if err != nil {
		return nil, &EmbedResolveError{Status: http.StatusInternalServerError, Message: "Error loading video"}
	}
	defer cursor.Close(ctx)

	mediaCount := 0
	for cursor.Next(ctx) {
		var media models.Media
		if err := cursor.Decode(&media); err != nil {
			continue
		}
		if media.Resolution != nil && *media.Resolution != "" {
			mediaCount++
		}
	}

	if mediaCount == 0 {
		var vp models.VideoProcess
		vpErr := models.VideoProcessModel.Col().FindOne(ctx, bson.M{"fileId": file.ID}).Decode(&vp)

		pd := &ProcessingData{State: "queue"}
		if vpErr == nil {
			status := ""
			if vp.Status != nil {
				status = *vp.Status
			}
			if status == "failed" {
				errMsg := "เกิดข้อผิดพลาดในการประมวลผล"
				if vp.Error != nil && *vp.Error != "" {
					errMsg = *vp.Error
				}
				pd = &ProcessingData{State: "error", Message: errMsg}
			} else {
				pct := 0.0
				if vp.OverallPercent != nil {
					pct = *vp.OverallPercent
				}
				pd = &ProcessingData{State: "processing", Percent: pct}
			}
		}

		return nil, &EmbedResolveError{
			Status:     http.StatusNotFound,
			Message:    "Video not ready",
			Processing: pd,
		}
	}

	// poster จาก storage (URL เต็ม ไม่ขึ้นกับ request) — ไม่มีค่อย fallback thumb
	posterURL := ""
	var posterMedia models.Media
	err = models.MediaModel.Col().FindOne(ctx, bson.M{
		"fileId":     file.ID,
		"type":       enums.MediaTypeImage,
		"resolution": enums.ResolutionPoster,
		"deletedAt":  bson.M{"$eq": nil},
	}).Decode(&posterMedia)
	if err == nil && posterMedia.StorageID != nil && *posterMedia.StorageID != "" {
		var storage models.Storage
		if sErr := models.StorageModel.Col().FindOne(ctx, bson.M{"_id": *posterMedia.StorageID}).Decode(&storage); sErr == nil {
			if storage.PublicURL != nil && *storage.PublicURL != "" {
				posterURL = strings.TrimRight(*storage.PublicURL, "/") + "/" + posterMedia.Slug + "/poster.jpg"
			}
		}
	}

	spriteExists := false
	var spriteMedia models.Media
	if err := models.MediaModel.Col().FindOne(ctx, bson.M{
		"fileId":    file.ID,
		"type":      enums.MediaTypeThumbnail,
		"fileName":  "sprite.vtt",
		"deletedAt": bson.M{"$eq": nil},
	}).Decode(&spriteMedia); err == nil {
		spriteExists = true
	}

	lk = embedLookup{File: file, PosterURL: posterURL, SpriteExists: spriteExists}
	cache.SetJSON(cacheKey, &lk) // ถึงจุดนี้ = ready แล้วเท่านั้น
	return &lk, nil
}

// overrideBool ทับค่าเดิมเฉพาะเมื่อมีการตั้งค่ามาจริง (ไม่ใช่ nil)
func overrideBool(dst *bool, v *bool) {
	if v != nil {
		*dst = *v
	}
}

func (h *Handler) resolveEmbed(r *http.Request, slug string) (*EmbedResolveResult, *EmbedResolveError) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	lk, rErr := resolveEmbedData(ctx, slug)
	if rErr != nil {
		return nil, rErr
	}
	file := lk.File

	domain, isDomainRequest := services.FindDomain(r.Host)
	if isDomainRequest {
		if domain == nil {
			return nil, &EmbedResolveError{Status: http.StatusNotFound, Message: "Domain not found"}
		}
		if domain.Status != "active" {
			return nil, &EmbedResolveError{Status: http.StatusForbidden, Message: "Domain is not verified"}
		}
		if !domain.Enable {
			return nil, &EmbedResolveError{Status: http.StatusNotFound, Message: "Domain is disabled"}
		}
	}

	if !CheckDomainSpace(r, file.SpaceID) {
		return nil, &EmbedResolveError{Status: http.StatusNotFound, Message: "File not found"}
	}

	// workspace ที่ถูกปิด/ลบ ห้ามเล่น
	// ⚠ ของเดิมเทียบกับ "error" ซึ่งไม่มีใน WorkspaceStatus (pending/active/
	// inactive/deleted) เงื่อนไขจึงไม่เคยเป็นจริง = ไม่ได้กันอะไรเลย
	if file.SpaceID != nil && *file.SpaceID != "" {
		space := services.FindSpace(*file.SpaceID)
		if space != nil &&
			(space.Status == enums.WorkspaceStatusInactive ||
				space.Status == enums.WorkspaceStatusDeleted) {
			return nil, &EmbedResolveError{Status: http.StatusNotFound, Message: "This content is currently unavailable"}
		}
	}

	reqProto := requestProtocol(r)
	playlistHost := services.GetDomainPlaylist(r.Host)
	previewHost := services.GetDomainPreview()
	staticHost := services.GetDomainStatic()

	playlistM3U8 := reqProto + "://" + playlistHost + "/" + slug + "/playlist.m3u8"

	posterURL := lk.PosterURL
	if posterURL == "" {
		thumbTime := 0
		if file.Metadata != nil && file.Metadata.Duration != nil {
			thumbTime = int(*file.Metadata.Duration / 2)
		}
		if staticHost != "" {
			posterURL = reqProto + "://" + staticHost + "/thumb/" + slug + "/" + fmt.Sprintf("%d", thumbTime) + ".jpg"
		} else if previewHost != "" {
			posterURL = reqProto + "://" + previewHost + "/thumb/" + slug + "/" + fmt.Sprintf("%d", thumbTime) + ".jpg"
		} else {
			posterURL = "/thumb/" + slug + "/" + fmt.Sprintf("%d", thumbTime) + ".jpg"
		}
	}

	spriteVttURL := ""
	if lk.SpriteExists {
		if staticHost != "" {
			spriteVttURL = reqProto + "://" + staticHost + "/" + slug + "/sprite/sprite.vtt"
		} else if previewHost != "" {
			spriteVttURL = reqProto + "://" + previewHost + "/" + slug + "/sprite/sprite.vtt"
		} else {
			spriteVttURL = "/" + slug + "/sprite/sprite.vtt"
		}
	}

	planType := "hobby"
	if file.SpaceID != nil && *file.SpaceID != "" {
		if plan := services.GetSpacePlan(*file.SpaceID); plan != nil {
			planType = plan.PlanType
		}
	}

	// เริ่มจากค่ากลางของระบบ แล้วให้ตั้งค่าของโดเมน (ถ้ามี) ทับเป็นรายตัว
	settings := services.GetPlayerSettings()

	baseColor := settings.BaseColor
	bgColor := settings.BgColor
	autostart := settings.AutoPlay
	mute := settings.MuteSound
	continuePlay := settings.ContinuePlay
	continuePlayArk := settings.ContinuePlayArk
	loop := settings.RepeatVideo
	playbackRate := settings.PlaybackRate
	pip := settings.PIP
	displayTitle := settings.DisplayTitle
	seekForward := settings.FastForward
	seekBackward := settings.Rewind
	seekStep := settings.SeekStep

	if domain != nil && domain.Player != nil {
		p := domain.Player
		if p.BaseColor != "" {
			baseColor = p.BaseColor
		}
		// ทับเฉพาะตัวที่โดเมนตั้งมาจริง (ไม่ใช่ nil) — ตัวที่ยังไม่เคยตั้งคงค่ากลางไว้
		overrideBool(&autostart, p.AutoPlay)
		overrideBool(&mute, p.MuteSound)
		overrideBool(&continuePlay, p.ContinuePlay)
		overrideBool(&continuePlayArk, p.ContinuePlayArk)
		overrideBool(&loop, p.RepeatVideo)
		overrideBool(&playbackRate, p.PlaybackRate)
		overrideBool(&pip, p.PIP)
		overrideBool(&displayTitle, p.DisplayTitle)
		overrideBool(&seekForward, p.FastForward)
		overrideBool(&seekBackward, p.Rewind)
		// 0 = ยังไม่เคยตั้งค่า (ไม่ใช่ "กรอ 0 วินาที") — ใช้ค่ากลางแทน
		if p.SeekStep > 0 {
			seekStep = p.SeekStep
		}
	}

	pipIcon := "enabled"
	if !pip {
		pipIcon = "disabled"
	}

	title := ""
	if displayTitle {
		title = file.Name
	}

	adSlug := services.ResolveAdSlug(planType, domain, file.SpaceID)

	advertHost := strings.Split(requestHost(r), ":")[0]
	if staticHost != "" {
		advertHost = staticHost
	}

	embedConfig := services.EmbedPlayerConfig{
		Title:   title,
		Adverts: adSlug,
		VdoID:   slug,
		Node: services.EmbedNode{
			Static:   advertHost,
			Playlist: playlistHost,
		},
		Autostart:    autostart,
		Mute:         mute,
		PipIcon:      pipIcon,
		BaseColor:    baseColor,
		BgColor:      bgColor,
		Cast:         settings.Cast,
		Loop:         loop,
		PlaybackRate: playbackRate,
		Seek: services.EmbedSeek{
			Seconds:   seekStep,
			Indicator: settings.SeekIndicator,
			Forward:   seekForward,
			Backward:  seekBackward,
		},
		ContinuePlayBack: services.EmbedContinuePlayback{
			Enable:     continuePlay,
			Ark:        continuePlayArk,
			AutoResume: false,
			Countdown:  20,
		},
		Gtag: settings.Gtag,
	}

	return &EmbedResolveResult{
		File:        file,
		Slug:        slug,
		EmbedConfig: embedConfig,
		Content: EmbedContent{
			PosterURL:    posterURL,
			PlaylistM3U8: playlistM3U8,
			SpriteVttURL: spriteVttURL,
		},
	}, nil
}
