package users

import (
	"encoding/base64"
	"errors"

	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/types"
	"github.com/gojicms/goji/core/utils/log"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//////////////////////////////////
// User Provider               //
//////////////////////////////////

type UserProvider struct{}

func (p *UserProvider) Name() string {
	return "users"
}

func (p *UserProvider) Description() string {
	return "Goji Users Management Service"
}

func (p *UserProvider) Priority() int {
	return 0
}

func (p *UserProvider) GetByID(id uint) (*types.User, error) {
	db := database.GetDB()
	var user types.User
	if err := db.Preload("Group").First(&user, id).Error; err != nil {
		log.Error("Users", "Could not find user with ID of %d", id)
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

func (p *UserProvider) GetByUsername(username string) (*types.User, error) {
	db := database.GetDB()
	var user types.User
	if err := db.Preload("Group").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	user.Password = ""
	return &user, nil
}

func (p *UserProvider) ValidateLogin(username, password string) (*types.User, error) {
	db := database.GetDB()
	var user types.User
	if err := db.Preload("Group").Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}

	passwordSaltPeppered := password + user.Salt + config.ActiveConfig.Application.Pepper
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(passwordSaltPeppered))

	if err != nil {
		return nil, errors.New("invalid password")
	}

	// Clear the password
	user.Password = ""
	return &user, nil
}

func (p *UserProvider) Create(user *types.User) (*types.User, error) {
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
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (p *UserProvider) Update(user *types.User) error {
	if user.Password != "" {
		passwordHashed, salt, err := encodePassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = passwordHashed
		user.Salt = salt
	}
	db := database.GetDB()
	return db.Save(user).Error
}

func (p *UserProvider) Delete(user *types.User) error {
	db := database.GetDB()
	return db.Delete(user).Error
}

func (p *UserProvider) Count() (int64, error) {
	db := database.GetDB()
	var count int64
	if err := db.Model(&types.User{}).Count(&count).Error; err != nil {
		log.Error("Users", "Failed to count documents: %s", err.Error())
		return 0, err
	}
	return count, nil
}

func (p *UserProvider) GetAll() (*[]types.User, error) {
	db := database.GetDB()
	var users []types.User
	if err := db.Preload("Group").Find(&users).Error; err != nil {
		return nil, err
	}
	// Hide password
	for _, user := range users {
		user.Password = ""
	}
	return &users, nil
}

func encodePassword(password string) (string, string, error) {
	saltUuid := uuid.New().String()
	salt := base64.StdEncoding.EncodeToString([]byte(saltUuid))[:16]
	passwordSaltPeppered := password + salt + config.ActiveConfig.Application.Pepper
	passwordHashed, err := bcrypt.GenerateFromPassword([]byte(passwordSaltPeppered), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}
	return string(passwordHashed), salt, nil
}

//////////////////////////////////
// Group Provider               //
//////////////////////////////////

type GroupProvider struct{}

func (p *GroupProvider) Name() string {
	return "groups"
}

func (p *GroupProvider) Description() string {
	return "Goji Groups Management Service"
}

func (p *GroupProvider) Priority() int {
	return 0
}

func (p *GroupProvider) GetByName(name string) (*types.Group, error) {
	db := database.GetDB()
	var group types.Group
	if err := db.Where("name = ?", name).First(&group).Error; err != nil {
		log.Error("Auth/Groups", "Failed to get group: %s", err.Error())
		return nil, err
	}
	return &group, nil
}

func (p *GroupProvider) GetAll() ([]*types.Group, error) {
	db := database.GetDB()
	var groups []*types.Group
	if err := db.Find(&groups).Error; err != nil {
		log.Error("Auth/Groups", "Failed to get groups: %s", err.Error())
		return nil, err
	}
	return groups, nil
}

func (p *GroupProvider) Create(group *types.Group) error {
	db := database.GetDB()
	return db.Create(group).Error
}

func (p *GroupProvider) Count() (int64, error) {
	db := database.GetDB()
	var count int64
	if err := db.Model(&types.Group{}).Count(&count).Error; err != nil {
		log.Error("Auth/Groups", "Failed to count groups: %s", err.Error())
		return 0, err
	}
	return count, nil
}

func (p *GroupProvider) Update(group *types.Group) error {
	db := database.GetDB()
	return db.Save(group).Error
}
