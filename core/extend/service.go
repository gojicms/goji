package extend

var services []*ServiceDef

type ServiceDef struct {
	Name         string        `json:"name"`
	FriendlyName string        `json:"friendly_name"`
	Resources    []ResourceDef `json:"resources"`
	Internal     bool          `json:"internal"`
	OnInit       func() error  `json:"-"`
}

func (d ServiceDef) ToApiJson() interface{} {
	return map[string]interface{}{
		"name":          d.Name,
		"friendly_name": d.FriendlyName,
	}
}

func RegisterService(service *ServiceDef) {
	services = append(services, service)
}

func GetServices() []*ServiceDef {
	servicesCopy := make([]*ServiceDef, len(services))
	copy(servicesCopy, services)
	return servicesCopy
}

func GetService(name string) *ServiceDef {
	for _, service := range services {
		if service.Name == name {
			return service
		}
	}
	return nil
}
