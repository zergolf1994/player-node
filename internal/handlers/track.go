package handlers

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"sync/atomic"
	"time"

	"player-node/internal/services"
)

// ─── Track heartbeat relay ───────────────────────────────────
//
// analytics.js ยิง POST /e/p แบบ same-origin (โดเมน embed นั้นๆ เลย —
// ไม่ต้องมีโดเมน track สาธารณะ ไม่มี CORS ไม่โดน ad-blocker จับโดเมนแปลก)
// player-node ส่งต่อให้ track-node ตาม setting "track_api" (sync จาก DB)
//
// สิ่งที่ต้องแนบไปเอง: ประเทศ + User-Agent ของคนดู — เพราะ request ขาออก
// จากเราไปหา track-node ผ่าน Cloudflare จะโดนประทับ CF-IPCountry เป็น
// ประเทศของ "เซิร์ฟเวอร์เรา" ไม่ใช่ของคนดู จึงส่งเป็น header แยก
// (X-Viewer-*) พร้อม X-Track-Key ให้ track-node เชื่อได้ว่ามาจากเราจริง

var trackClient = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConns:        64,
		MaxIdleConnsPerHost: 64, // ปลายทางเดียว — เก็บ connection ไว้ใช้ซ้ำ
		IdleConnTimeout:     60 * time.Second,
	},
}

// จำกัดจำนวน forward ที่ค้างอยู่ — track-node ช้า/ล่ม แล้ว goroutine
// ต้องไม่กองจนเรา (ตัวเสิร์ฟ player) ล้มตาม ทิ้ง heartbeat ดีกว่า
var trackInflight = make(chan struct{}, 256)

// TrackBeacon handles POST /e/p — ตอบ 204 ทันที forward เป็นงานเบื้องหลัง
func (h *Handler) TrackBeacon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 2048))

	// ตอบก่อนเสมอ — ปัญหาฝั่ง track ต้องไม่กลายเป็นความช้าของ player
	w.WriteHeader(http.StatusNoContent)

	if err != nil || len(body) == 0 {
		return
	}
	base := services.GetTrackAPI()
	if base == "" {
		return // ไม่ได้ตั้ง track_api = ปิด tracking เงียบๆ
	}

	country := r.Header.Get("CF-IPCountry")
	ua := r.Header.Get("User-Agent")
	// โดเมน player ที่ heartbeat นี้วิ่งเข้ามา (custom domain ของ embed) —
	// client ไม่ต้องส่งเอง เรารู้จาก Host ของ request อยู่แล้ว ปลอมยาก
	embedHost := requestHost(r)
	key := services.GetTrackAPIKey()

	select {
	case trackInflight <- struct{}{}:
		go func() {
			defer func() { <-trackInflight }()
			forwardTrack(base, key, body, country, ua, embedHost)
		}()
	default:
		// คิวเต็ม (track-node ช้า/ล่ม) — ทิ้ง heartbeat นี้ไป รอบหน้ามาใน 10 วิ
	}
}

func forwardTrack(base, key string, body []byte, country, ua, embedHost string) {
	req, err := http.NewRequest(http.MethodPost, base+"/e/p", bytes.NewReader(body))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "text/plain")
	if key != "" {
		req.Header.Set("X-Track-Key", key)
	}
	if country != "" {
		req.Header.Set("X-Viewer-Country", country)
	}
	if ua != "" {
		req.Header.Set("X-Viewer-UA", ua)
	}
	if embedHost != "" {
		req.Header.Set("X-Embed-Host", embedHost)
	}

	resp, err := trackClient.Do(req)
	if err != nil {
		// log แบบไม่รัว — heartbeat มาถี่มาก ถ้า track ล่มจะ log ท่วม
		logTrackErr(err)
		return
	}
	io.Copy(io.Discard, resp.Body)
	resp.Body.Close()
}

// unix วินาทีของ log ล่าสุด — atomic เพราะ forward รันหลาย goroutine พร้อมกัน
var lastTrackErrAt atomic.Int64

func logTrackErr(err error) {
	now := time.Now().Unix()
	last := lastTrackErrAt.Load()
	if now-last < 30 || !lastTrackErrAt.CompareAndSwap(last, now) {
		return
	}
	log.Printf("⚠️ track forward failed (throttled log): %v", err)
}
