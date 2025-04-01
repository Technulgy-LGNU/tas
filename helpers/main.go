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
	var userKey = dbHelpers.BrowserToken{
		DeviceId: "test",
		Key:      "test",
	}
	var user = dbHelpers.Member{
		Name:     "Elias Braun",
		Email:    "elias.braun@gmail.com",
		Password: "test",
		Birthday: time.Now(),
		Gender:   "male",
		Perms:    &perms,
		Tokens:   &[]dbHelpers.BrowserToken{userKey},
	}

	result := db.Create(&user)
	if result.Error != nil {
		panic(result.Error)
	}
}
