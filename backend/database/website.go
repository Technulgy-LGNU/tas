package database

import "time"

type LocalizedText struct {
	DE string `json:"de"`
	EN string `json:"en"`
}

func (t LocalizedText) Get(lang string) string {
	if lang == "de" {
		return t.DE
	}
	return t.EN
}

type WebsiteImage struct {
	ID  string        `json:"id"`
	Alt LocalizedText `json:"alt"`
}
type WebsiteVideo struct {
	Title LocalizedText `json:"title"`
	URL   string        `json:"url"`
}
type TeamAward struct {
	EventID string        `json:"eventId"`
	League  LocalizedText `json:"league"`
	Result  LocalizedText `json:"result"`
}

// Content is stored as structured JSONB; shared fields have the same shape across editors.
type WebsiteContent struct {
	Name         LocalizedText  `json:"name"` 
	Description  LocalizedText  `json:"description"`
	Images       []WebsiteImage `json:"images"`
	CoverImageID string         `json:"coverImageId"`
	About        LocalizedText  `json:"about"`
	Videos       []WebsiteVideo `json:"videos"`
	TeamStatus   string         `json:"teamStatus"`
	Awards       []TeamAward    `json:"awards"`
	Date         string         `json:"date"`
	URL          string         `json:"url"`
	Blocks       []ArticleBlock `json:"blocks"`
}

type WebsiteEntry struct {
	ID        string         `gorm:"type:uuid;primaryKey" json:"id"`
	Kind      string         `gorm:"not null;uniqueIndex:website_kind_slug" json:"kind"`
	Slug      string         `gorm:"not null;uniqueIndex:website_kind_slug" json:"slug"`
	Published bool           `gorm:"not null" json:"published"`
	PublishAt *time.Time     `json:"publishAt"`
	SortOrder int            `gorm:"not null" json:"sortOrder"`
	Version   int            `gorm:"not null" json:"version"`
	Content   WebsiteContent `gorm:"serializer:json;type:jsonb;not null" json:"content"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

// References make dependency checks reliable without scanning translated text/JSON.
type WebsiteReference struct {
	EntryID  string `gorm:"type:uuid;primaryKey"`
	TargetID string `gorm:"type:uuid;primaryKey"`
	Kind     string `gorm:"primaryKey"`
}

// Text blocks contain Markdown; other blocks carry structured data.
type ArticleBlock struct {
	ID     string         `json:"id"`
	Type   string         `json:"type"`
	Text   LocalizedText  `json:"text"`
	Level  int            `json:"level"`
	Images []WebsiteImage `json:"images"`
	URL    string         `json:"url"`
}

func (c WebsiteContent) AllImages() []WebsiteImage {
	result := append([]WebsiteImage{}, c.Images...)
	for _, block := range c.Blocks {
		result = append(result, block.Images...)
	}
	return result
}
