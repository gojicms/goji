package documents

import (
	"regexp"

	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/services/auth/users"
	"github.com/gojicms/goji/core/utils/log"
	"gorm.io/gorm"
)

type Document struct {
	gorm.Model
	Title       string      `json:"title" gorm:"size:255"`
	Description string      `json:"description" gorm:"size:1000"`
	Content     string      `json:"content" gorm:"type:text"`
	CreatedBy   *users.User `json:"-" gorm:"foreignKey:CreatedById;constraint:OnUpdate:NO ACTION,OnDelete:SET NULL;"`
	CreatedById *uint       `json:"-" gorm:"index"`
}

// Summary returns a summary of the document content
// maxLength is the maximum length of the summary
// Returns the summary of the document content
func (document *Document) Summary(maxLength int) string {
	heads := regexp.MustCompile(`</h.*?>`)
	re := regexp.MustCompile(`<.*?>`)
	truncated := heads.ReplaceAllString(document.Content, " — ")
	truncated = re.ReplaceAllString(truncated, "")
	length := len(truncated)
	if length > maxLength {
		length = maxLength
	}
	return truncated[:length] + "..."
}

// Get gets all documents, limited by limit and starting at offset.
// sort is the field to sort by and which direction, and can be "created_at", "updated_at", or "title", eg. "createdAt DESC" or "title ASC"
// If sort is empty, it will default to "createdAt DESC"
func Get(limit int, offset int, sort string) ([]Document, error) {
	db := database.GetDB()
	var documents []Document

	if sort == "" {
		sort = "createdAt DESC"
	}

	res := db.Preload("CreatedBy").Limit(limit).Offset(offset).Order(sort).Find(&documents)

	if res.Error != nil {
		log.Error("Documents", "Failed to get documents: %s", res.Error.Error())
		return nil, res.Error
	}

	return documents, nil
}

// GetById gets a document by id
// id is the id of the document to get
// Returns the document and an error if there is one
func GetById(id string) (*Document, error) {
	db := database.GetDB()
	var document Document
	res := db.Preload("CreatedBy").Where("id = ?", id).Find(&document)
	if res.Error != nil {
		log.Error("Documents", "Failed to get document: %s", res.Error.Error())
		return nil, res.Error
	}
	return &document, nil
}

// Create creates a new document
// document is the document to create
// Returns the created document and an error if there is one
func Create(document Document) (*Document, error) {
	db := database.GetDB()
	res := db.Debug().Create(&document)
	if res.Error != nil {
		log.Error("Documents", "Failed to create document: %s", res.Error.Error())
		return nil, res.Error
	}
	return &document, nil
}

// Update updates a document
// document is the document to update
// Returns the updated document and an error if there is one
func Update(document Document) (*Document, error) {
	db := database.GetDB()
	res := db.Save(&document)
	if res.Error != nil {
		log.Error("Documents", "Failed to update document: %s", res.Error.Error())
	}
	return &document, nil
}

// DeleteById deletes a document by id
// id is the id of the document to delete
// Returns the number of rows affected and an error if there is one
func DeleteById(id string) (int64, error) {
	db := database.GetDB()
	res := db.Where("id = ?", id).Delete(&Document{})
	if res.Error != nil {
		log.Error("Documents", "Failed to delete document: %s", res.Error.Error())
		return 0, res.Error
	}
	return res.RowsAffected, nil
}

// Count counts the number of documents
// Returns the number of documents and an error if there is one
func Count() (int64, error) {
	db := database.GetDB()
	var count int64
	res := db.Model(&Document{}).Count(&count)
	if res.Error != nil {
		log.Error("Documents", "Failed to count documents: %s", res.Error.Error())
		return 0, res.Error
	}
	return count, nil
}

// GetByTitle gets a document by title
// title is the title of the document to get
// Returns the document and an error if there is one
func GetByTitle(title string) (*Document, error) {
	db := database.GetDB()
	var document Document
	res := db.Where("title = ?", title).Find(&document)
	if res.Error != nil {
		log.Error("Documents", "Failed to get document by title: %s", res.Error.Error())
		return nil, res.Error
	}
	return &document, nil
}
