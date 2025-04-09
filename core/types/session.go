package types

import (
	"time"

	"gorm.io/gorm"
)

// Session defines the base session type
type Session struct {
	gorm.Model
	SessionId string    `gorm:"index;size:36"` // UUID is 36 chars
	CSRF      string    `gorm:"size:255"`
	UserId    uint      `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
}
