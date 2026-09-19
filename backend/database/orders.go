package database

import "time"

// Prices are integer EUR cents; amounts are whole units.
type OrderPartFields struct {
	Name           string `json:"name"`
	Amount         int64  `json:"amount"`
	UnitPriceCents int64  `json:"unitPriceCents"`
	Shop           string `json:"shop"`
	Link           string `json:"link"`
	CategoryID     string `json:"categoryId"`
}
type OrderCategory struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
type OrderPart struct {
	ID string `json:"id"`
	OrderPartFields
	CreatedBy     string     `json:"createdBy"`
	CreatedByName string     `json:"createdByName"`
	CreatedAt     time.Time  `json:"createdAt"`
	RequestID     string     `json:"requestId,omitempty"`
	OrderedAt     *time.Time `json:"orderedAt"`
	OrderedBy     string     `json:"orderedBy"`
}
type OrderRequest struct {
	ID string `json:"id"`
	OrderPartFields
	Status        string     `json:"status"`
	CreatedBy     string     `json:"createdBy"`
	CreatedByName string     `json:"createdByName"`
	CreatedAt     time.Time  `json:"createdAt"`
	ReviewedBy    string     `json:"reviewedBy"`
	ReviewedAt    *time.Time `json:"reviewedAt"`
	ReviewNote    string     `json:"reviewNote"`
}
type OrderContent struct {
	Categories []OrderCategory `json:"categories"`
	Parts      []OrderPart     `json:"parts"`
	Requests   []OrderRequest  `json:"requests"`
}

// List mutations lock this row and check Version, so approval, closure and edits
// are atomic and a request cannot be approved twice or added after closure.
type OrderList struct {
	ID        string       `gorm:"type:uuid;primaryKey" json:"id"`
	Name      string       `gorm:"not null" json:"name"`
	Status    string       `gorm:"not null;index" json:"status"`
	Currency  string       `gorm:"not null" json:"currency"`
	Version   int          `gorm:"not null" json:"version"`
	Content   OrderContent `gorm:"serializer:json;type:jsonb;not null" json:"content"`
	CreatedBy string       `json:"createdBy"`
	CreatedAt time.Time    `json:"createdAt"`
	UpdatedAt time.Time    `json:"updatedAt"`
	ClosedAt  *time.Time   `json:"closedAt"`
}
type StandardPart struct {
	ID             string    `gorm:"type:uuid;primaryKey" json:"id"`
	Name           string    `gorm:"not null" json:"name"`
	Amount         int64     `gorm:"not null" json:"amount"`
	UnitPriceCents int64     `gorm:"not null" json:"unitPriceCents"`
	Shop           string    `gorm:"not null" json:"shop"`
	Link           string    `json:"link"`
	Version        int       `gorm:"not null" json:"version"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
