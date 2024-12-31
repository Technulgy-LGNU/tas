package database

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
	"tas/src/config"
	cLog "tas/src/log"
	"time"
)

type User struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name     string
	Email    string
	Password string
	Birthday time.Time
	Gender   string
	Tokens   *[]BrowserTokens
	Perms    Permission
	Request  *[]Request

	TeamID *uint64 `gorm:"index"`
	Team   *Team
}

type BrowserTokens struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	DeviceId string
	Key      string
	UserID   uint64 `gorm:"index"`
	User     User
}

// Rule of thump: (Will be probably clarified at the point of implementation)
// 1: See
// 2: Edit
// 3: Admin

type Permission struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Login      *bool
	Admin      *bool
	Members    *int
	Teams      *int
	Events     *int
	Newsletter *int
	Form       *int
	Website    *int
	Orders     *int

	UserID uint64 `gorm:"index"`
	User   *User
}

type ResetPassword struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Email string
	Code  string
}

type Team struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name    string
	League  string
	Members []User
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

	Name     string
	FromForm string
	Email    string
	Message  string
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
	Clicks      []Click
}

type Click struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	IP     string
	Clicks int

	LinkID uint64 `gorm:"index"`
	Link   Link
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

	UserID uint64 `gorm:"index"`
	User   User
}

type RepeatingOrder struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name  string
	Price float32
	Shop  string
	Link  string
}

type Post struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name        string
	Text        string
	ReleaseDate time.Time
}

type File struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name      string
	Location  string
	ShortLink string
}

func InitDatabase(Logger *cLog.Logger, cfg *config.CFG, db *gorm.DB) error {
	var err error

	// Auto migrating
	err = db.AutoMigrate(&User{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&BrowserTokens{})
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

	err = db.AutoMigrate(&Click{})
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

	err = db.AutoMigrate(&Post{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	err = db.AutoMigrate(&File{})
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	// Initial Admin user
	type data struct {
		AdminUserCreated bool `yaml:"AdminUserCreated"`
	}

	if _, err := os.Stat("data/cache.yaml"); errors.Is(err, os.ErrNotExist) {
		Logger.LogEvent("Cache file not found, creating new one, ...", "INFO")

		file, err := os.Create("data/cache.yaml")
		if err != nil {
			return errors.New(fmt.Sprintf("error creating file: %v\n", err))
		}
		log.Println("Cache file created")
		defer file.Close()

		_, err = file.Write([]byte("AdminUserCreated: false"))
		if err != nil {
			return errors.New(fmt.Sprintf("error writing to file: %v\n", err))
		}
	}
	var d data

	dataFile, err := os.Open("data/cache.yaml")
	if err != nil {
		return errors.New(fmt.Sprintf("error opening file: %v\n", err))
	}

	yamlParser := yaml.NewDecoder(dataFile)
	err = yamlParser.Decode(&d)
	if err != nil {
		return errors.New(fmt.Sprintf("error reading file: %v\n", err))
	}

	if !d.AdminUserCreated {
		log.Println("Creating admin user ...")
		var tr = true
		U := &User{
			Name:     "admin",
			Email:    "admin@example.com",
			Password: cfg.User.AdminPassword,
			Perms: Permission{
				Login: &tr,
				Admin: &tr,
			},
		}
		db.Create(&U)
		err = changeAdminStatus()
		if err != nil {
			return errors.New(fmt.Sprintf("error changing admin status: %v\n", err))
		}
		Logger.AddToLogBuffer("Admin user created", "INFO")
	}

	Logger.LogEvent("Database initialized", "INFO")
	return nil
}

func changeAdminStatus() error {
	file, err := os.ReadFile("data/cache.yaml")
	if err != nil {
		return errors.New(fmt.Sprintf("error migrating table: %v\n", err))
	}

	lines := strings.Split(string(file), "\n")
	for i, line := range lines {
		if strings.Contains(line, "AdminUserCreated: false") {
			lines[i] = "AdminUserCreated: true\n"
		}
	}
	output := strings.Join(lines, "\n")
	err = os.WriteFile("data/cache.yaml", []byte(output), 0644)
	if err != nil {
		return errors.New(fmt.Sprintf("error writing to file: %v\n", err))
	}

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	return nil
}
