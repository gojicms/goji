package users

import (
	"encoding/base64"
	"errors"

	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/services/auth/groups"
	"github.com/gojicms/goji/core/utils/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Uuid        string `gorm:"unique"`
	Username    string `gorm:"unique"`
	Password    string
	Salt        string
	Email       string
	DisplayName string
	GroupName   string
	Group       *groups.Group `gorm:"foreignKey:GroupName;references:Name"`
}

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

func encodePassword(password string) (string, string, error) {
	saltUuid := uuid.New().String()
	salt := base64.StdEncoding.EncodeToString([]byte(saltUuid))[:16]
	passwordSaltPeppered := password + salt + config.ActiveConfig.Application.Pepper
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(passwordSaltPeppered), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return string(passwordHashed), string(salt), err
}

func Create(user *User) (*User, error) {
	// TODO: Add a config for user credential restrictions such as username/password length
	// Ensure a password was provided
	if user.Password == "" {
		return nil, errors.New("password is empty")
	}
	// Ensure a username is provided
	if user.Username == "" {
		return nil, errors.New("username is empty")
	}
	// Generate a salt and store the password
	passwordHashed, salt, err := encodePassword(user.Password)

	if err != nil {
		return nil, err
	}

	user.Salt = salt
	user.Password = string(passwordHashed)
	user.Uuid = uuid.New().String()

	db := database.GetDB()
	err = db.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetAll() (*[]User, error) {
	db := database.GetDB()
	var users []User
	err := db.Model(&User{}).Preload("Group").Find(&users).Error
	// Hide password
	for _, user := range users {
		user.Password = ""
	}
	if err != nil {
		return nil, err
	}
	return &users, nil
}

func GetById(id uint) (*User, error) {
	db := database.GetDB()
	var user User
	res := db.Model(&User{}).Preload("Group").First(&user, id)
	if res.Error != nil {
		log.Error("Users", "Could not find user with ID of %d", id)
		return nil, res.Error
	}
	user.Password = ""
	return &user, nil
}

func Count() (int64, error) {
	db := database.GetDB()
	var count int64
	res := db.Model(&User{}).Count(&count)
	if res.Error != nil {
		log.Error("Users", "Failed to count documents: %s", res.Error.Error())
		return 0, res.Error
	}
	return count, nil
}

func Update(user *User) error {
	db := database.GetDB()
	if user.Password != "" {
		passwordHashed, salt, err := encodePassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = passwordHashed
		user.Salt = salt
	}
	err := db.Model(&User{}).Where("id = ?", user.ID).Updates(user).Error
	if err != nil {
		log.Error("Users", "Failed to update user with ID of %d", user.ID)
		return err
	}
	return nil
}

func Delete(user *User) error {
	db := database.GetDB()
	err := db.Model(&User{}).Delete(&user, user.ID).Error
	if err != nil {
		log.Error("Users", "Failed to delete user with ID of %d", user.ID)
		return err
	}
	return nil
}

// ValidateLogin checks the username/password and returns a user if one exists.
// for security purposes the password should be stripped if passed to the client.
func ValidateLogin(username string, password string) (*User, error) {
	db := database.GetDB()
	user := &User{}
	db.Model(user).Preload("Group").Where("username = ?", username).First(user)

	passwordSaltPeppered := password + user.Salt + config.ActiveConfig.Application.Pepper
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordSaltPeppered))

	if err != nil {
		return nil, err
	}

	// Clear the password
	user.Password = ""
	return user, nil
}
