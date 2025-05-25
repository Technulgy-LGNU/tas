package util

import (
	"gorm.io/gorm"
	"log"
	"tas/src/database"
	"time"
)

// DeleteOldSessions Deletes session ids, that are older than thirty days, remember, these devices will be logged out
// runs twice a day
// Currently not used due to bugs
/*
func DeleteOldSessions(db *gorm.DB) {
	ticker := time.NewTicker(12 * time.Hour)
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	go func() {
		for {
			<-ticker.C
			var userKeys []database.BrowserToken
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
} */

// DeleteSoftDeletedUserKeys Deletes every user key, that was soft deleted
// (maybe including the other dbs as well in the future, if necessary)
// runs once a day
func DeleteSoftDeletedUserKeys(db *gorm.DB) {
	ticker := time.NewTicker(24 * time.Hour)
	sixMonthsAgo := time.Now().AddDate(0, -6, 0)
	go func() {
		<-ticker.C
		var userKeys []database.BrowserToken
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

// DeleteOldPasswordResetCodes Deletes password reset codes that are older than 1 hour
// runs every hour
func DeleteOldPasswordResetCodes(db *gorm.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	oneHourAgo := time.Now().Add(-1 * time.Hour)
	go func() {
		<-ticker.C
		var codes []database.ResetPassword
		err := db.Find(&codes).Where("created_at < ?", oneHourAgo).Error
		if err != nil {
			log.SetFlags(log.LstdFlags | log.Lshortfile)
			log.Printf("Error deleting old password reset codes: %v\n", err)
		}
		for _, code := range codes {
			db.Unscoped().Delete(&code)
		}
	}()
}
