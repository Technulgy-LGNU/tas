package database

import "time"

// Image stores library metadata; the actual file lives in Cloudflare Images.
type Image struct {
	ID           string    `gorm:"type:uuid;primaryKey" json:"id"`
	CloudflareID string    `gorm:"not null;uniqueIndex" json:"cloudflareId"`
	Name         string    `gorm:"type:varchar(200);not null" json:"name"`
	AltText      string    `gorm:"type:varchar(1000);not null" json:"altText"`
	Filename     string    `gorm:"not null" json:"filename"`
	ContentType  string    `gorm:"not null" json:"contentType"`
	Size         int64     `gorm:"not null" json:"size"`
	Status       string    `gorm:"not null;default:pending" json:"status"`
	UploadedBy   string    `gorm:"not null" json:"uploadedBy"`
	CreatedAt    time.Time `gorm:"index" json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}
