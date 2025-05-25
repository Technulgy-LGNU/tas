package main

import (
	"golang.org/x/crypto/bcrypt"
	"helpers/config"
	"helpers/dbHelpers"
	"log"
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
		Inventory:  3,
	}
	var user = dbHelpers.Member{
		Name:     "Elias Braun",
		Email:    "braunelias@tghd.email",
		Password: HashString("test"),
		Gender:   "male",
		Perms:    &perms,
		TeamID:   1,
	}

	result := db.Create(&user)
	if result.Error != nil {
		panic(result.Error)
	}
}

func HashString(key string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing String: %v\n", err)
		return ""
	}
	return string(bytes)
}
