package extend

import (
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils"
)

//////////////////////////////////
// Types                        //
//////////////////////////////////

type ResourceDef struct {
	HttpValidator
	// The handler for this resource
	Handler func(flow *httpflow.HttpFlow) `json:"-"`
	// The method for this resource
	Description string `json:"-"`
	// Allows overriding the default can handle behaviors
	WillHandle func(flow *httpflow.HttpFlow) bool `json:"-"`
}

//////////////////////////////////
// Type Methods                 //
//////////////////////////////////

// ToApiJson Converts the resource to a user-friendly API-consumable object
func (r *ResourceDef) ToApiJson() interface{} {
	return utils.Object{
		"method":      r.Method,
		"path":        r.Path,
		"description": r.Description,
	}
}

// CanHandle Checks if the resource can handle the given HttpFlow. Additionally, if the route matches
// then
// - flow: The HTTP Flow to validate against
// Returns true if the resource can handle the request, false otherwise
func (r *ResourceDef) CanHandle(flow *httpflow.HttpFlow) bool {
	canHandle := r.HttpValidator.CanHandle(flow)
	if !canHandle {
		return false
	}

	if r.WillHandle != nil {
		return r.WillHandle(flow)
	}
	return canHandle
}
