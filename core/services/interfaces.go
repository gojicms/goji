package services

import (
	"io"
	"mime/multipart"

	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/types"
)

// Service is the base interface that all services must implement
type Service interface {
	// Name returns the unique identifier for this service
	Name() string

	// Description returns a human-readable description of what this service does
	Description() string

	// Priority returns the priority of this service (higher numbers take precedence)
	Priority() int
}

// FileService is the interface for file handling services
type FileService interface {
	Service

	// ExecuteTemplate executes the given template content within the context of an HttpFlow
	// This method only processes the template and returns the result, without any HTTP handling
	ExecuteTemplate(content []byte, flow *httpflow.HttpFlow) ([]byte, error)

	// ExecuteTemplateFromPath executes a template from the given path within the context of an HttpFlow
	// This method only processes the template and returns the result, without any HTTP handling
	ExecuteTemplateFromPath(path string, flow *httpflow.HttpFlow) ([]byte, error)

	// RenderTemplate renders a template from the given path and writes it directly to the HttpFlow
	// This method handles both template processing and HTTP response writing
	RenderTemplate(content []byte, flow *httpflow.HttpFlow) error

	// RenderTemplateFromPath renders a template from the given path and writes it directly to the HttpFlow
	RenderTemplateFromPath(path string, flow *httpflow.HttpFlow) error

	// RenderFile renders a file from the given path and writes it directly to the HttpFlow
	RenderFile(path string, flow *httpflow.HttpFlow) error
}

// UserService defines the interface for user operations
type UserService interface {
	Service
	GetByID(id uint) (*types.User, error)
	GetByUsername(username string) (*types.User, error)
	ValidateLogin(username, password string) (*types.User, error)
	Create(user *types.User) (*types.User, error)
	Update(user *types.User) error
	Delete(user *types.User) error
	Count() (int64, error)
	GetAll() (*[]types.User, error)
}

// GroupService defines the interface for group operations
type GroupService interface {
	Service
	GetByName(name string) (*types.Group, error)
	GetAll() ([]*types.Group, error)
	Create(group *types.Group) error
	Update(group *types.Group) error
	Count() (int64, error)
}

// MediaSource defines how a storage backend should work
type MediaSource interface {
	// Name returns the unique identifier for this source
	Name() string

	// FriendlyName returns a human-readable name
	FriendlyName() string

	// Store stores a file and returns its path/identifier
	Store(file *multipart.FileHeader, flow *httpflow.HttpFlow) (string, error)

	// Retrieve gets a file by its path/identifier
	Retrieve(path string) (io.Reader, error)

	// Delete removes a file
	Delete(path string) error

	// List files in a directory/path
	List(path string) ([]string, error)
}

// MediaService interface for the main service
type MediaService interface {
	Service

	// RegisterSource registers a new media source
	RegisterSource(source MediaSource)

	// GetSource returns a source by name
	GetSource(name string) (MediaSource, error)

	// Upload handles file uploads using the specified source
	Upload(sourceName string, file *multipart.FileHeader, flow *httpflow.HttpFlow) (string, error)

	// Get retrieves a file from the specified source
	Get(sourceName string, path string) (io.Reader, error)

	// Delete removes a file from the specified source
	Delete(sourceName string, path string) error

	// List files in a directory from the specified source
	List(sourceName string, path string) ([]string, error)
}
