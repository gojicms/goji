package database

import (
	"reflect"

	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/utils/log"
	"gorm.io/gorm"
)

var (
	db *gorm.DB
)

func GetDB() *gorm.DB {
	var err error
	if db == nil {
		db, err = gorm.Open(
			config.ActiveConfig.Application.Database.Connector(),
			&config.ActiveConfig.Application.Database.Config)
		if err != nil {
			log.Fatal(log.RCDatabase, "Database", "Failed to connect to database: %s", err.Error())
		}
	}
	return db
}

func IsReadOnly() (bool, error) {
	// Try to create the table
	err := db.Exec("CREATE TABLE IF NOT EXISTS __rw_test (id INT)").Error
	if err != nil {
		// If there's an error other than the table already existing, return false and the error
		return false, err
	}

	// If the table creation succeeded or it already existed, try to drop it
	err = db.Exec("DROP TABLE IF EXISTS __rw_test").Error
	if err != nil {
		// If there's an error dropping the table, return false and the error
		return false, err
	}

	// If both operations succeeded, the database is writable
	return true, nil
}

func AutoMigrate(target interface{}) {
	log.Info("Database", "Migrating database for %s", reflect.TypeOf(target))
	err := GetDB().AutoMigrate(target)
	if err != nil {
		log.Fatal(log.RCDatabase, "Database", "Failed to migrate database: %s", err.Error())
	}
}
