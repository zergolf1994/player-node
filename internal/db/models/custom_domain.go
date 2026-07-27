package models

import (
	"time"

	"github.com/zergolf1994/goose"
)

// PlayerConfig holds video player configuration for a custom domain.
//
// ⚠ ทุก toggle เป็น *bool ไม่ใช่ bool — schema ฝั่งหลัก (custom-domain.model.ts)
// ตั้ง default: false ทุกตัว ถ้าใช้ bool ธรรมดา doc ที่ยังไม่มีฟิลด์นั้นจะอ่านได้
// false เหมือนกับ "ตั้งใจปิด" แยกกันไม่ออก → ฟีเจอร์หายเงียบๆ
// nil = ไม่ได้ตั้ง ให้ไปใช้ค่ากลางจาก services.GetPlayerSettings()
type PlayerConfig struct {
	Skin            string  `bson:"skin" json:"skin"`
	LogoImageURL    *string `bson:"logoImageUrl,omitempty" json:"logoImageUrl,omitempty"`
	LogoWebsiteURL  *string `bson:"logoWebsiteUrl,omitempty" json:"logoWebsiteUrl,omitempty"`
	LogoPosition    *string `bson:"logoPosition,omitempty" json:"logoPosition,omitempty"`
	PosterURL       *string `bson:"posterUrl,omitempty" json:"posterUrl,omitempty"`
	BaseColor       string  `bson:"baseColor" json:"baseColor"`
	DisplayTitle    *bool   `bson:"displayTitle,omitempty" json:"displayTitle,omitempty"`
	AutoPlay        *bool   `bson:"autoPlay,omitempty" json:"autoPlay,omitempty"`
	MuteSound       *bool   `bson:"muteSound,omitempty" json:"muteSound,omitempty"`
	RepeatVideo     *bool   `bson:"repeatVideo,omitempty" json:"repeatVideo,omitempty"`
	ContinuePlay    *bool   `bson:"continuePlay,omitempty" json:"continuePlay,omitempty"`
	ContinuePlayArk *bool   `bson:"continuePlayArk,omitempty" json:"continuePlayArk,omitempty"`
	Sharing         *bool   `bson:"sharing,omitempty" json:"sharing,omitempty"`
	Captions        *bool   `bson:"captions,omitempty" json:"captions,omitempty"`
	PlaybackRate    *bool   `bson:"playbackRate,omitempty" json:"playbackRate,omitempty"`
	Keyboard        *bool   `bson:"keyboard,omitempty" json:"keyboard,omitempty"`
	Download        *bool   `bson:"download,omitempty" json:"download,omitempty"`
	PIP             *bool   `bson:"pip,omitempty" json:"pip,omitempty"`
	ShowPreviewTime *bool   `bson:"showPreviewTime,omitempty" json:"showPreviewTime,omitempty"`
	FastForward     *bool   `bson:"fastForward,omitempty" json:"fastForward,omitempty"`
	Rewind          *bool   `bson:"rewind,omitempty" json:"rewind,omitempty"`
	SeekStep        int     `bson:"seekStep" json:"seekStep"`
}

// AdContent holds a single embedded advert entry (video, image, or script).
type AdContent struct {
	ID          string   `bson:"_id,omitempty" json:"_id,omitempty"`
	Enabled     bool     `bson:"enabled" json:"enabled"`
	Name        string   `bson:"name" json:"name"`
	MP4URL      *string  `bson:"mp4Url,omitempty" json:"mp4Url,omitempty"`
	SkipSeconds *int     `bson:"skipSeconds,omitempty" json:"skipSeconds,omitempty"`
	ImageURL    *string  `bson:"imageUrl,omitempty" json:"imageUrl,omitempty"`
	ShowOn      []string `bson:"showOn,omitempty" json:"showOn,omitempty"`
	WebsiteURL  *string  `bson:"websiteUrl,omitempty" json:"websiteUrl,omitempty"`
	Script      *string  `bson:"script,omitempty" json:"script,omitempty"`
}

// DomainAdvertCategory groups enabled flag with a list of adverts.
type DomainAdvertCategory struct {
	Enabled bool        `bson:"enabled" json:"enabled"`
	List    []AdContent `bson:"list" json:"list"`
}

// DomainAdverts holds embedded video, image, and script adverts on a domain.
type DomainAdverts struct {
	Video  DomainAdvertCategory `bson:"video" json:"video"`
	Image  DomainAdvertCategory `bson:"image" json:"image"`
	Script DomainAdvertCategory `bson:"script" json:"script"`
}

// DomainDNS holds DNS configuration for domain verification.
type DomainDNS struct {
	RecordType        string     `bson:"recordType" json:"recordType"`
	Value             string     `bson:"value" json:"value"`
	TTL               int        `bson:"ttl" json:"ttl"`
	VerificationToken string     `bson:"verificationToken" json:"verificationToken"`
	RetryCount        int        `bson:"retryCount" json:"retryCount"`
	LastVerified      *time.Time `bson:"lastVerified,omitempty" json:"lastVerified,omitempty"`
	Reason            *string    `bson:"reason,omitempty" json:"reason,omitempty"`
}

// CustomDomain represents a custom domain with player/ad config.
// Collection: "custom_domains" | _id: String (UUID)
type CustomDomain struct {
	ID        string         `bson:"_id" json:"id" goose:"required,default:uuid"`
	Enable    bool           `bson:"enable" json:"enable"`
	Name      string         `bson:"name" json:"name" goose:"required,unique"`
	Status    string         `bson:"status" json:"status" goose:"default:pending"` // pending, active, failed, expired
	CreatorID *string        `bson:"creatorId,omitempty" json:"creatorId,omitempty" goose:"ref:user,index"`
	SpaceID   *string        `bson:"spaceId,omitempty" json:"spaceId,omitempty" goose:"ref:workspaces,index"`
	Slug      string         `bson:"slug" json:"slug" goose:"unique,default:random(11),index"`
	DNS       *DomainDNS     `bson:"dns,omitempty" json:"dns,omitempty"`
	Player    *PlayerConfig  `bson:"player,omitempty" json:"player,omitempty"`
	Adverts   *DomainAdverts `bson:"adverts,omitempty" json:"adverts,omitempty"`
	CreatedAt time.Time      `bson:"createdAt" json:"createdAt" goose:"default:now"`
	UpdatedAt time.Time      `bson:"updatedAt" json:"updatedAt" goose:"default:now"`
}

// CustomDomainModel is the goose model for the "custom_domains" collection.
var CustomDomainModel = goose.NewModel[CustomDomain]("custom_domains")
