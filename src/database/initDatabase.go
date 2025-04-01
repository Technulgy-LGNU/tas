package database

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"time"
)

type Member struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name     string
	Email    string
	Password string
	Birthday time.Time
	Gender   string
	Tokens   *[]BrowserToken
	Perms    *Permission
	Request  *[]Request

	TeamID *uint64 `gorm:"index"`
	Team   *Team
}

type BrowserToken struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	DeviceId string
	Key      string
	MemberID uint64 `gorm:"index"`
	Member   Member
}

// Rule of thump: (Will be probably clarified at the point of implementation)
// 0: None
// 1: See
// 2: Edit
// 3: Delete

type Permission struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Login      bool
	Admin      bool
	Members    int
	Teams      int
	Events     int
	Newsletter int
	Form       int
	Website    int
	Orders     int
	Sponsors   int

	MemberID uint64 `gorm:"index"`
	Member   Member
}

type ResetPassword struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	UserId uint64
	Email  string
	Code   string
}

type Team struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name    string
	League  string
	Members []Member
	History []ParticipationHistory

	EventID uint64 `gorm:"index"`
	Event   Event
}

type ParticipationHistory struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Year   int
	Event  string
	League string
	Place  string

	TeamID uint64 `gorm:"index"`
	Team   Team
}

type Event struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name            string
	StartDate       time.Time
	EndDate         time.Time
	RegisteredTeams []Team
}

type Sponsor struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name   string
	Joined time.Time
	Left   *time.Time
	Gifts  []Gift
}

type Gift struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Year   int
	Amount int

	SponsorID uint64 `gorm:"index"`
	Sponsor   Sponsor
}

type Form struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name    string
	Email   string
	Message string
}

type Newsletter struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name      string
	Followers *[]Follower
	Letters   *[]News
}

type Follower struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Email string

	NewsletterID uint64 `gorm:"index"`
	Newsletter   Newsletter
}

type News struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name     string
	SendDate *time.Time
	Content  *string

	NewsletterID uint64 `gorm:"index"`
	Newsletter   Newsletter
}

type Link struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	OriginalURL string
	ForwardURL  string
	Clicks      int64
}

type Order struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name   string
	Orders []OrderEntry
}

type OrderEntry struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name   string
	Amount int
	Price  float32
	Shop   string
	Link   string
	Notes  string

	OrderID uint64 `gorm:"index"`
	Order   Order
}

type Request struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Approved bool
	Name     string
	Amount   int
	Shop     string
	Link     string
	Notes    string

	MemberID uint64 `gorm:"index"`
	Member   Member
}

type RepeatingOrder struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name  string
	Price float32
	Shop  string
	Link  string
}

type TDPList struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Team string
	Year int
	URL  string
}

func InitDatabase(db *gorm.DB) error {
	var err error

	// Auto migrating
	err = db.AutoMigrate(&Member{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&BrowserToken{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Permission{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&ResetPassword{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Team{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&ParticipationHistory{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Event{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Sponsor{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Gift{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Form{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Newsletter{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&News{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Link{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Order{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&OrderEntry{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&Request{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&RepeatingOrder{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&TDPList{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	log.Println("INFO: Database initialized")
	return nil
}
