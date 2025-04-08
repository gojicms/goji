package groups

import (
	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
	"gorm.io/gorm"
)

type Group struct {
	gorm.Model
	Name        string    `gorm:"unique"`
	Permissions utils.CSV `gorm:"type:VARCHAR(512)"`
}

func Create(group *Group) error {
	db := database.GetDB()
	return db.Create(group).Error
}

func Count() (int64, error) {
	db := database.GetDB()
	var count int64
	res := db.Model(&Group{}).Count(&count)
	if res.Error != nil {
		log.Error("Auth/Groups", "Failed to count groups: %s", res.Error.Error())
		return 0, res.Error
	}
	return count, nil
}

func GetByName(s string) (*Group, error) {
	db := database.GetDB()
	var group Group
	res := db.Where("name = ?", s).First(&group)
	if res.Error != nil {
		log.Error("Auth/Groups", "Failed to get group: %s", res.Error.Error())
		return nil, res.Error
	}
	return &group, nil
}

func GetAll() ([]*Group, error) {
	db := database.GetDB()
	var groups []*Group
	res := db.Find(&groups)
	if res.Error != nil {
		log.Error("Auth/Groups", "Failed to get groups: %s", res.Error.Error())
		return nil, res.Error
	}
	return groups, nil
}
