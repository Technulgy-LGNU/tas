package main

import (
	"helpers/config"
	"helpers/dbHelpers"
	"time"
)

func main() {
	cfg := config.GetConfig()

	db, err := dbHelpers.GetDatabase(cfg)
	if err != nil {
		panic(err)
	}
	err = dbHelpers.InitDatabase(db)
	if err != nil {
		panic(err)
	}

	var perms = dbHelpers.Permission{
		Login:      true,
		Admin:      true,
		Members:    3,
		Teams:      3,
		Events:     3,
		Newsletter: 3,
		Form:       3,
		Website:    3,
		Orders:     3,
	}
	var user = dbHelpers.Member{
		Name:     "Elias Braun",
		Email:    "braunelias@tghd.email",
		Password: "test",
		Birthday: time.Now(),
		Gender:   "male",
		Perms:    &perms,
	}

	result := db.Create(&user)
	if result.Error != nil {
		panic(result.Error)
	}
}
