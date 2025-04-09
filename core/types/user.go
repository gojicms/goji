package types

import (
	"gorm.io/gorm"
)

// User defines the base user type
type User struct {
	gorm.Model
	Uuid        string `gorm:"unique;size:36"` // UUID is 36 chars
	Username    string `gorm:"unique;size:255"`
	Password    string `gorm:"size:255"`
	Salt        string `gorm:"size:32"`
	Email       string `gorm:"size:255"`
	DisplayName string `gorm:"size:255"`
	GroupName   string `gorm:"size:255"`
	Group       *Group `gorm:"foreignKey:GroupName;references:Name"`
}

// HasPermission checks if the user has a specific permission
func (u User) HasPermission(s string) bool {
	if s == "" {
		return true
	}
	for _, v := range u.Group.Permissions {
		if v == s {
			return true
		}
	}
	return false
}
