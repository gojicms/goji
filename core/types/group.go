package types

import (
	"github.com/gojicms/goji/core/utils"
	"gorm.io/gorm"
)

// Group defines the base group type
type Group struct {
	gorm.Model
	Name        string    `gorm:"unique"`
	Permissions utils.CSV `gorm:"type:VARCHAR(512)"`
	Internal    bool      `gorm:"default:false"`
}
