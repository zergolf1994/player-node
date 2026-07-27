package handlers

import (
	"encoding/json"
	"net/http"

	"player-node/internal/services"
)

// HealthResponse — เพิ่ม host/slug เพื่อยืนยันว่าโดเมนนี้ผูกกับ record ไหน
type HealthResponse struct {
	Status  string `json:"status"`
	Service string `json:"service"`
	Version string `json:"version"`
	Host    string `json:"host"`
	Slug    string `json:"slug,omitempty"`
}

// Health ตอบสถานะบริการ พร้อม slug ของโดเมนที่เรียกเข้ามา
//
// ใช้ requestHost ตัวเดียวกับหน้า embed (อ่าน X-Forwarded-Host ก่อน) ผลลัพธ์
// จึงตรงกับโดเมนที่ระบบใช้ตัดสินใจจริง ไม่ใช่ค่าที่ proxy เห็น
//
//	localhost            → ไม่มี domain (slug ว่าง) ใช้ค่ากลางของระบบ
//	โดเมนที่ยังไม่ลงทะเบียน → domain: null
//	โดเมนที่ลงทะเบียนแล้ว   → มี slug/status/enable ให้ตรวจ
func Health(version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host := requestHost(r)

		res := HealthResponse{
			Status:  "ok",
			Service: "player-node",
			Version: version,
			Host:    host,
		}

		if domain, isDomainRequest := services.FindDomain(host); isDomainRequest && domain != nil {
			res.Slug = domain.Slug
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(res)
	}
}
