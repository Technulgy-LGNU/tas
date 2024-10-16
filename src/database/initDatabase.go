package database

import (
	"gorm.io/gorm"
	"log"
	"time"
)

type Users struct {
	gorm.Model

	Name     string
	Email    string
	Password string
	Keys     *[]UserKeys
	Perms    Permissions
}

type UserKeys struct {
	gorm.Model
	DeviceId string
	Key      string
}

// Rule of thump: (Will be probably clarified at the point of implementation)
// 1: See
// 2: Edit
// 3: Admin

type Permissions struct {
	gorm.Model

	Login      bool
	Admin      bool
	Members    int
	Teams      int
	Events     int
	Newsletter int
	Form       int
	Website    int
	Orders     int
}

type Teams struct {
	gorm.Model

	Name    string
	Members []Users
	League  string
	History []ParticipationHistory
}

type ParticipationHistory struct {
	gorm.Model

	Year   int
	Event  string
	League string
	Place  string
}

type Events struct {
	gorm.Model

	Name            string
	StartDate       time.Time
	EndDate         time.Time
	RegisteredTeams []Teams
}

type Sponsors struct {
	gorm.Model

	Name   string
	Joined time.Time
	Left   *time.Time
	Gifts  []Gifts
}

type Gifts struct {
	gorm.Model

	Year   int
	Amount int
}

type Forms struct {
	gorm.Model

	Name     string
	FromForm string
	Email    string
	Message  string
}

type Newsletters struct {
	gorm.Model

	Name          string
	FollowersDB   string
	NewslettersDB string
}

type Orders struct {
	gorm.Model

	Name     string
	OrdersDB string
}

type Website struct {
	gorm.Model

	Name        string
	Text        string
	ReleaseDate time.Time
}

func InitDatabase(db *gorm.DB) {
	var err error

	// Auto migrating
	err = db.AutoMigrate(&Users{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Users: %v\n", err)
	}

	err = db.AutoMigrate(&UserKeys{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating UserKeys: %v\n", err)
	}

	err = db.AutoMigrate(&Permissions{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Permissions: %v\n", err)
	}

	err = db.AutoMigrate(&Teams{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Teams: %v\n", err)
	}

	err = db.AutoMigrate(&ParticipationHistory{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating ParticipationHistory: %v\n", err)
	}

	err = db.AutoMigrate(&Events{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Events: %v\n", err)
	}

	err = db.AutoMigrate(&Sponsors{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Sponsors: %v\n", err)
	}

	err = db.AutoMigrate(&Forms{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Forms: %v\n", err)
	}

	err = db.AutoMigrate(&Newsletters{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Newsletters: %v\n", err)
	}

	err = db.AutoMigrate(&Orders{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Orders: %v\n", err)
	}

	err = db.AutoMigrate(&Website{})
	if err != nil {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Printf("Error migrating Website: %v\n", err)
	}

	// Schemas and nonGorm Stuff
	var sqlDB, _ = db.DB()
	defer sqlDB.Close()
	// Initial Admin user
}
