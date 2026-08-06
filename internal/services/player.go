package services

import "encoding/json"

// PlayerSettings holds global default player options used when no domain override exists.
type PlayerSettings struct {
	BaseColor       string `json:"baseColor"`
	BgColor         string `json:"bgColor"`
	AutoPlay        bool   `json:"autoPlay"`
	MuteSound       bool   `json:"muteSound"`
	ContinuePlay    bool   `json:"continuePlay"`
	ContinuePlayArk bool   `json:"continuePlayArk"`
	RepeatVideo     bool   `json:"repeatVideo"`
	PlaybackRate    bool   `json:"playbackRate"`
	PIP             bool   `json:"pip"`
	Cast            bool   `json:"cast"`
	DisplayTitle    bool   `json:"displayTitle"`
	FastForward     bool   `json:"fastForward"`
	Rewind          bool   `json:"rewind"`
	SeekIndicator   bool   `json:"seekIndicator"`
	SeekStep        int    `json:"seekStep"`
	Gtag            string `json:"gtag"`
}

// EmbedContinuePlayback matches test_jwplayer continuePlayBack object.
type EmbedContinuePlayback struct {
	Enable     bool `json:"enable"`
	Ark        bool `json:"ark"`
	AutoResume bool `json:"autoResume"`
	Countdown  int  `json:"countdown"`
}

// EmbedNode คือโดเมนปลายทางที่ player ใช้ประกอบ URL เอง
//
//	static   → advert feed + sprite/thumb
//	playlist → //{playlist}/{vdoId}/playlist.m3u8
type EmbedNode struct {
	Static   string `json:"static,omitempty"`
	Playlist string `json:"playlist,omitempty"`
}

// EmbedSeek คุมปุ่มกรอ — ตรงกับ PLAYER_CONFIG.seek ของ player-ui.js
type EmbedSeek struct {
	Seconds   int  `json:"seconds"`
	Indicator bool `json:"indicator"`
	Forward   bool `json:"forward"`
	Backward  bool `json:"backward"`
}

// EmbedPlayerConfig ถูกฉีดเป็น window.PLAYER_CONFIG ในหน้า embed
// (bundle อยู่ที่ asset-cdn.vdohide.com)
type EmbedPlayerConfig struct {
	Title            string                `json:"title,omitempty"`
	Adverts          string                `json:"adverts,omitempty"`
	VdoID            string                `json:"vdoId"`
	Node             EmbedNode             `json:"node"`
	Sprite           bool                  `json:"sprite,omitempty"`
	Image            any                   `json:"image,omitempty"`
	Autostart        bool                  `json:"autostart"`
	Mute             bool                  `json:"mute"`
	PipIcon          string                `json:"pipIcon"`
	BaseColor        string                `json:"baseColor"`
	BgColor          string                `json:"bgColor"`
	Cast             bool                  `json:"cast"`
	Loop             bool                  `json:"loop"`
	Seek             EmbedSeek             `json:"seek"`
	PlaybackRate     bool                  `json:"playbackRate"`
	ContinuePlayBack EmbedContinuePlayback `json:"continuePlayBack"`
	Gtag             string                `json:"gtag,omitempty"`
}

// DefaultSeekStep — ใช้เมื่อ domain ไม่ได้ตั้ง หรือตั้งเป็น 0
const DefaultSeekStep = 10

// GetPlayerSettings returns the hardcoded global default player settings.
func GetPlayerSettings() PlayerSettings {
	return PlayerSettings{
		BaseColor:       "#ff8800",
		BgColor:         "#000000",
		AutoPlay:        false,
		MuteSound:       false,
		ContinuePlay:    true,
		ContinuePlayArk: false,
		RepeatVideo:     false,
		PlaybackRate:    true,
		PIP:             true,
		Cast:            true,
		DisplayTitle:    false,
		FastForward:     true,
		Rewind:          true,
		SeekIndicator:   true,
		SeekStep:        DefaultSeekStep,
		Gtag:            "G-N8WE91Q200",
	}
}

// IsMaintenanceMode reads player_maintenance from setting.json.
func IsMaintenanceMode() bool {
	settings, err := ReadSettingFile()
	if err != nil {
		return false
	}
	raw, exists := settings["player_maintenance"]
	if !exists {
		return false
	}
	var val bool
	if err := json.Unmarshal(raw, &val); err != nil {
		return false
	}
	return val
}
