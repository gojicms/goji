package extend

import (
	"html/template"
	"sync"
)

// GlobalFuncMap is a map of functions that are available to all templates
// It's safe for concurrent use from multiple goroutines
var (
	globalFuncMap = template.FuncMap{}
	funcMapMutex  = &sync.RWMutex{}
)

// RegisterFunction adds a function to the global function map
// This is safe for concurrent use from multiple goroutines
// - name: The name of the function as it will be available to templates
// - fn: The function that gets called. Can take any amount of arguments but should only return one value.
func RegisterFunction(name string, fn interface{}) {
	funcMapMutex.Lock()
	defer funcMapMutex.Unlock()
	globalFuncMap[name] = fn
}

// RegisterFunctions adds multiple functions to the global function map
// This is safe for concurrent use from multiple goroutines
// - funcs: Multiple FuncMap functions
func RegisterFunctions(funcs template.FuncMap) {
	funcMapMutex.Lock()
	defer funcMapMutex.Unlock()
	for name, fn := range funcs {
		globalFuncMap[name] = fn
	}
}

// GlobalFunctions returns a copy of the current global functions
func GlobalFunctions() template.FuncMap {
	funcMapMutex.RLock()
	funcMap := make(template.FuncMap, len(globalFuncMap))
	for name, fn := range globalFuncMap {
		funcMap[name] = fn
	}
	funcMapMutex.RUnlock()
	return funcMap
}
