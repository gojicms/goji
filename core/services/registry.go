package services

import (
	"github.com/gojicms/goji/core/utils/log"
)

type serviceRegistry struct {
	services map[string][]Service
}

var registry = &serviceRegistry{
	services: make(map[string][]Service),
}

// RegisterService registers a new service
func RegisterService(service Service) {
	// Get the type of service (e.g., "files", "template", etc.)
	serviceType := getServiceType(service)

	// Add to the list of services for this type
	registry.services[serviceType] = append(registry.services[serviceType], service)
	log.Info("Service", "Registered service for %s: %s", serviceType, service.Name())

	// Sort services by priority (highest first)
	sortServices(registry.services[serviceType])
}

// GetService returns the highest priority service of the given type
func GetService(serviceType string) Service {
	services := registry.services[serviceType]
	if len(services) == 0 {
		return nil
	}

	return services[0]
}

// GetServices returns all services of the given type, sorted by priority
func GetServices(serviceType string) []Service {
	return registry.services[serviceType]
}

// GetServiceOfType returns the highest priority service of the given type
// If no service is found, it logs a fatal error since we should always have a default service
func GetServiceOfType[T Service](serviceType string) T {
	service := GetService(serviceType)
	if service == nil {
		log.Fatal(log.RCUnknownError, "Service", "No service of type %s found - this should never happen", serviceType)
	}

	t, ok := service.(T)
	if !ok {
		log.Fatal(log.RCUnknownError, "Service", "Service is not of expected type %T", t)
	}

	return t
}

// Helper function to get the service type
func getServiceType(service Service) string {
	switch service.(type) {
	case FileService:
		return "files"
	case UserService:
		return "users"
	case GroupService:
		return "groups"
	default:
		return "unknown"
	}
}

// Helper function to sort services by priority
func sortServices(services []Service) {
	// Sort in descending order (highest priority first)
	for i := 0; i < len(services); i++ {
		for j := i + 1; j < len(services); j++ {
			if services[i].Priority() < services[j].Priority() {
				services[i], services[j] = services[j], services[i]
			}
		}
	}
}
