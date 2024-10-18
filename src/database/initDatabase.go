package database

import (
	"errors"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"log"
	"os"
	"strings"
	"tas/src/config"
	"time"
)

type User struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name     string
	Email    string
	Password string
	Keys     *[]UserKey
	Perms    Permission

	TeamID *uint64 `gorm:"index"`
	Team   *Team
}

type UserKey struct {
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

	EventsID uint64 `gorm:"index"`
	Events   Event
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
	Gifts  []Gifts
}

type Gifts struct {
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

	Name        string
	Followers   *[]Follower
	Newsletters *[]News
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

type Order struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name     string
	OrdersDB string
}

type Website struct {
	gorm.Model
	ID uint64 `gorm:"primaryKey"`

	Name        string
	Text        string
	ReleaseDate time.Time
}

func InitDatabase(cfg *config.CFG, db *gorm.DB) {
	var err error

	// Auto migrating
	err = db.AutoMigrate(&User{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Users: %v\n", err)
	}

	err = db.AutoMigrate(&UserKey{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating UserKeys: %v\n", err)
	}

	err = db.AutoMigrate(&Permission{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Permissions: %v\n", err)
	}

	err = db.AutoMigrate(&Team{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Teams: %v\n", err)
	}

	err = db.AutoMigrate(&ParticipationHistory{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating ParticipationHistory: %v\n", err)
	}

	err = db.AutoMigrate(&Event{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Events: %v\n", err)
	}

	err = db.AutoMigrate(&Sponsor{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Sponsors: %v\n", err)
	}

	err = db.AutoMigrate(&Form{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Forms: %v\n", err)
	}

	err = db.AutoMigrate(&Newsletter{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Newsletters: %v\n", err)
	}

	err = db.AutoMigrate(&Order{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Orders: %v\n", err)
	}

	err = db.AutoMigrate(&Website{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Website: %v\n", err)
	}

	// Initial Admin user
	if os.Args[1] == "prod" {

	} else if os.Args[1] == "dev" {
		type data struct {
			AdminUserCreated bool `yaml:"AdminUserCreated"`
		}

		if _, err := os.Stat("data/cache.yaml"); errors.Is(err, os.ErrNotExist) {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Println("Cache file not found, creating ...")

			file, err := os.Create("data/cache.yaml")
			if err != nil {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Fatalf("Error creating cache file: %v\n", err)
			}
			log.Println("Cache file created")
			defer file.Close()

			_, err = file.Write([]byte("AdminUserCreated: false"))
			if err != nil {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Printf("Error writing to cache file: %v\n", err)
			}
		}
		var d data

		dataFile, err := os.Open("data/cache.yaml")
		if err != nil {
			log.SetFlags(log.LstdFlags & log.Lshortfile)
			log.Fatalf("Error readeing config file: %d\n", err)
		}

		yamlParser := yaml.NewDecoder(dataFile)
		err = yamlParser.Decode(&d)
		if err != nil {
			log.SetFlags(log.LstdFlags & log.Lshortfile)
			log.Fatalf("Error readeing config file: %d\n", err)
		}

		if !d.AdminUserCreated {
			log.Println("Creating admin user ...")
			var tr = true
			U := &User{
				Name:     cfg.User.AdminUserName,
				Email:    "admin@example.com",
				Password: cfg.User.AdminPassword,
				Perms: Permission{
					Login: &tr,
					Admin: &tr,
				},
			}
			db.Create(&U)
			changeAdminStatus()
			log.Println("Admin user created")
		}
	} else {
		panic("Error: Wrong command line argument")
	}
}

func changeAdminStatus() {
	file, err := os.ReadFile("data/cache.yaml")
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error opening cache file: %v\n", err)
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
		log.Fatalln(err)
	}
}
