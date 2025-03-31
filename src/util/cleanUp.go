package util

import (
	"gorm.io/gorm"
	"log"
	"tas/src/database"
	"time"
)

// DeleteOldSessions Deletes session ids, that are older than thirty days, remember, these devices will be logged out
// runs twice a day
func DeleteOldSessions(db *gorm.DB) {
	ticker := time.NewTicker(12 * time.Hour)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	go func() {
		for {
			<-ticker.C
			var userKeys []database.BrowserTokens
			err := db.Find(&userKeys).Where("created_at < ?", thirtyDaysAgo).Error
			if err != nil {
				log.SetFlags(log.LstdFlags | log.Lshortfile)
				log.Printf("Error deleting old user keys: %v\n", err)
			}
			for _, userKey := range userKeys {
				db.Delete(&userKey)
			}
		}
	}()
}

// DeleteSoftDeletedUserKeys Deletes every user key, that was soft deleted
// (maybe including the other dbs as well in the future, if necessary)
// runs once a day
func DeleteSoftDeletedUserKeys(db *gorm.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)
	go func() {
		<-ticker.C
		var userKeys []database.BrowserTokens
		err := db.Find(&userKeys).Where("deleted_at < ?", sixMonthsAgo).Error
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error deleting soft deleted user keys: %v\n", err)
		}
		for _, userKey := range userKeys {
			db.Unscoped().Delete(&userKey)
		}
	}()
}

// DeleteOldTDPs Deletes TDPs that were already soft deleted and are older than 30 days
// runs once a day
func DeleteOldTDPs(db *gorm.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	go func() {
		<-ticker.C
		var tdp []database.TDPList
		err := db.Find(&tdp).Where("deleted_at < ?", thirtyDaysAgo).Error
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error deleting old TDPs: %v\n", err)
		}
		for _, t := range tdp {
			db.Unscoped().Delete(&t)
		}
	}()
}

// DeleteOldForms Deletes forms that were already soft deleted and are older than 3 months
// runs once a day
func DeleteOldForms(db *gorm.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	go func() {
		<-ticker.C
		var forms []database.Form
		err := db.Find(&forms).Where("deleted_at < ?", threeMonthsAgo).Error
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error deleting old forms: %v\n", err)
		}
		for _, form := range forms {
			db.Unscoped().Delete(&form)
		}
	}()
}
