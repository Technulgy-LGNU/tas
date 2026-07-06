package database

import "time"

const (
	MemberStatusPending  = "pending"
	MemberStatusApproved = "approved"
	MemberStatusRejected = "rejected"

	OrderRequestStatusPending  = "pending"
	OrderRequestStatusApproved = "approved"
	OrderRequestStatusRejected = "rejected"

	OrderListStatusDraft     = "draft"
	OrderListStatusPublished = "published"
	OrderListStatusArchived  = "archived"
)

type Base struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Permission struct {
	Base
	Key         string `gorm:"uniqueIndex;size:128;not null" json:"key"`
	Description string `gorm:"size:255" json:"description"`
}

type Role struct {
	Base
	Name        string       `gorm:"uniqueIndex;size:96;not null" json:"name"`
	Description string       `gorm:"size:255" json:"description"`
	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

type Member struct {
	Base
	FusionAuthUserID string     `gorm:"uniqueIndex;size:128;not null" json:"fusion_auth_user_id"`
	Email            string     `gorm:"size:255;index" json:"email"`
	Name             string     `gorm:"size:255" json:"name"`
	Status           string     `gorm:"size:32;index;not null;default:'pending'" json:"status"`
	LastLoginAt      *time.Time `json:"last_login_at"`
	Roles            []Role     `gorm:"many2many:member_roles;" json:"roles,omitempty"`
}

type InventoryCategory struct {
	Base
	Name      string              `gorm:"size:160;not null;index:idx_inventory_category_parent_name,unique" json:"name"`
	ParentID  *uint               `gorm:"index:idx_inventory_category_parent_name,unique" json:"parent_id"`
	Parent    *InventoryCategory  `json:"parent,omitempty"`
	Children  []InventoryCategory `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	SortOrder int                 `json:"sort_order"`
}

type InventoryItem struct {
	Base
	Name                   string             `gorm:"size:255;not null;index" json:"name"`
	Quantity               int                `gorm:"not null;default:0" json:"quantity"`
	VendorID               string             `gorm:"size:160;index" json:"vendor_id"`
	ProductURL             string             `gorm:"type:text" json:"product_url"`
	Website                string             `gorm:"type:text" json:"website"`
	Notes                  string             `gorm:"type:text" json:"notes"`
	CategoryID             *uint              `gorm:"index" json:"category_id"`
	Category               *InventoryCategory `json:"category,omitempty"`
	Confirmed              bool               `gorm:"not null;default:true;index" json:"confirmed"`
	MatchedInventoryItemID *uint              `gorm:"index" json:"matched_inventory_item_id"`
	MatchedInventoryItem   *InventoryItem     `json:"matched_inventory_item,omitempty"`
	SourceOrderListItemID  *uint              `gorm:"index" json:"source_order_list_item_id"`
}

type OrderRequest struct {
	Base
	Name            string         `gorm:"size:255;not null;index" json:"name"`
	Quantity        int            `gorm:"not null;default:1" json:"quantity"`
	UnitPriceCents  int64          `gorm:"not null;default:0" json:"unit_price_cents"`
	TotalPriceCents int64          `gorm:"not null;default:0" json:"total_price_cents"`
	URL             string         `gorm:"type:text" json:"url"`
	Notes           string         `gorm:"type:text" json:"notes"`
	ShopDomain      string         `gorm:"size:255;index" json:"shop_domain"`
	ShopName        string         `gorm:"size:255;index" json:"shop_name"`
	Status          string         `gorm:"size:32;index;not null;default:'pending'" json:"status"`
	RequesterID     uint           `gorm:"index;not null" json:"requester_id"`
	Requester       *Member        `json:"requester,omitempty"`
	ApprovedByID    *uint          `gorm:"index" json:"approved_by_id"`
	ApprovedBy      *Member        `json:"approved_by,omitempty"`
	RejectedByID    *uint          `gorm:"index" json:"rejected_by_id"`
	RejectedBy      *Member        `json:"rejected_by,omitempty"`
	OrderListItemID *uint          `gorm:"index" json:"order_list_item_id"`
	OrderListItem   *OrderListItem `json:"order_list_item,omitempty"`
}

func (r *OrderRequest) BeforeSave() error {
	if r.Quantity < 1 {
		r.Quantity = 1
	}
	if r.UnitPriceCents < 0 {
		r.UnitPriceCents = 0
	}
	r.TotalPriceCents = int64(r.Quantity) * r.UnitPriceCents
	return nil
}

type OrderList struct {
	Base
	Name        string          `gorm:"size:255;not null" json:"name"`
	Status      string          `gorm:"size:32;index;not null;default:'draft'" json:"status"`
	PublishedAt *time.Time      `json:"published_at"`
	CreatedByID uint            `gorm:"index;not null" json:"created_by_id"`
	CreatedBy   *Member         `json:"created_by,omitempty"`
	Items       []OrderListItem `json:"items,omitempty"`
}

type OrderListItem struct {
	Base
	OrderListID             uint           `gorm:"index;not null" json:"order_list_id"`
	OrderList               *OrderList     `json:"order_list,omitempty"`
	OrderRequestID          *uint          `gorm:"index" json:"order_request_id"`
	OrderRequest            *OrderRequest  `json:"order_request,omitempty"`
	Name                    string         `gorm:"size:255;not null;index" json:"name"`
	Quantity                int            `gorm:"not null;default:1" json:"quantity"`
	UnitPriceCents          int64          `gorm:"not null;default:0" json:"unit_price_cents"`
	TotalPriceCents         int64          `gorm:"not null;default:0" json:"total_price_cents"`
	URL                     string         `gorm:"type:text" json:"url"`
	Notes                   string         `gorm:"type:text" json:"notes"`
	ShopDomain              string         `gorm:"size:255;index" json:"shop_domain"`
	ShopName                string         `gorm:"size:255;index" json:"shop_name"`
	Ordered                 bool           `gorm:"not null;default:false" json:"ordered"`
	OrderedAt               *time.Time     `json:"ordered_at"`
	Received                bool           `gorm:"not null;default:false" json:"received"`
	ReceivedAt              *time.Time     `json:"received_at"`
	ReceivedByID            *uint          `gorm:"index" json:"received_by_id"`
	ReceivedBy              *Member        `json:"received_by,omitempty"`
	ReceivedInventoryItemID *uint          `gorm:"index" json:"received_inventory_item_id"`
	ReceivedInventoryItem   *InventoryItem `json:"received_inventory_item,omitempty"`
}

func (i *OrderListItem) BeforeSave() error {
	if i.Quantity < 1 {
		i.Quantity = 1
	}
	if i.UnitPriceCents < 0 {
		i.UnitPriceCents = 0
	}
	i.TotalPriceCents = int64(i.Quantity) * i.UnitPriceCents
	return nil
}

type UploadedImage struct {
	Base
	CloudflareImageID string  `gorm:"uniqueIndex;size:255;not null" json:"cloudflare_image_id"`
	Filename          string  `gorm:"size:255" json:"filename"`
	ContentType       string  `gorm:"size:120" json:"content_type"`
	DeliveryURL       string  `gorm:"type:text" json:"delivery_url"`
	VariantsJSON      string  `gorm:"type:text" json:"variants_json"`
	UploadedByID      *uint   `gorm:"index" json:"uploaded_by_id"`
	UploadedBy        *Member `json:"uploaded_by,omitempty"`
}

type Team struct {
	Base
	Name      string         `gorm:"size:180;not null;index" json:"name"`
	Slug      string         `gorm:"uniqueIndex;size:180;not null" json:"slug"`
	Summary   string         `gorm:"type:text" json:"summary"`
	ImageID   *uint          `gorm:"index" json:"image_id"`
	Image     *UploadedImage `json:"image,omitempty"`
	Published bool           `gorm:"not null;default:false;index" json:"published"`
	SortOrder int            `json:"sort_order"`
	Prizes    []Prize        `json:"prizes,omitempty"`
}

type Competition struct {
	Base
	Name      string  `gorm:"size:220;not null;index" json:"name"`
	Slug      string  `gorm:"uniqueIndex;size:220;not null" json:"slug"`
	Location  string  `gorm:"size:220" json:"location"`
	StartsOn  *string `gorm:"size:32" json:"starts_on"`
	EndsOn    *string `gorm:"size:32" json:"ends_on"`
	Published bool    `gorm:"not null;default:false;index" json:"published"`
	SortOrder int     `json:"sort_order"`
	Prizes    []Prize `json:"prizes,omitempty"`
}

type Prize struct {
	Base
	Title         string       `gorm:"size:220;not null" json:"title"`
	Placement     string       `gorm:"size:120" json:"placement"`
	Notes         string       `gorm:"type:text" json:"notes"`
	TeamID        uint         `gorm:"index;not null" json:"team_id"`
	Team          *Team        `json:"team,omitempty"`
	CompetitionID uint         `gorm:"index;not null" json:"competition_id"`
	Competition   *Competition `json:"competition,omitempty"`
	SortOrder     int          `json:"sort_order"`
}

type SponsorCategory struct {
	Base
	Name      string    `gorm:"uniqueIndex;size:180;not null" json:"name"`
	SortOrder int       `json:"sort_order"`
	Sponsors  []Sponsor `json:"sponsors,omitempty"`
}

type Sponsor struct {
	Base
	Name        string           `gorm:"size:220;not null;index" json:"name"`
	CategoryID  uint             `gorm:"index;not null" json:"category_id"`
	Category    *SponsorCategory `json:"category,omitempty"`
	LogoImageID *uint            `gorm:"index" json:"logo_image_id"`
	LogoImage   *UploadedImage   `json:"logo_image,omitempty"`
	WebsiteURL  string           `gorm:"type:text" json:"website_url"`
	Published   bool             `gorm:"not null;default:false;index" json:"published"`
	SortOrder   int              `json:"sort_order"`
}

type HomeArticle struct {
	Base
	Slot      int            `gorm:"uniqueIndex;not null" json:"slot"`
	Title     string         `gorm:"size:220;not null" json:"title"`
	Body      string         `gorm:"type:text" json:"body"`
	ImageID   *uint          `gorm:"index" json:"image_id"`
	Image     *UploadedImage `json:"image,omitempty"`
	Published bool           `gorm:"not null;default:false;index" json:"published"`
}
