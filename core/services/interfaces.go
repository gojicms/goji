package services

import (
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
	Count() (int64, error)
}
