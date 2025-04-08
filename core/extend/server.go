package extend

import (
	"regexp"
	"sort"
	"strings"

	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils/log"
)

//////////////////////////////////
// Types                        //
//////////////////////////////////

var globalMiddleware []Middleware
var globalHandlers []Handler

// HttpValidator TODO: This name is odd - rename
type HttpValidator struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	pathRegexp *regexp.Regexp
}

type Handler struct {
	HttpValidator
	ServeHTTP func(*httpflow.HttpFlow)
}

type Middleware struct {
	HttpValidator
	Action   func(*httpflow.HttpFlow)
	Priority int
}

//////////////////////////////////
// Type Methods                 //
//////////////////////////////////

// EnsureRegex ensures that the regular expression - if applicable - is compiled
func (h *HttpValidator) EnsureRegex() {
	if h.pathRegexp == nil && h.Path != "" && h.Path != "*" {
		h.pathRegexp = regexp.MustCompile(h.Path)
	}
}

// CanHandle validates if a given flow matches the given HttpValidator
func (h *HttpValidator) CanHandle(flow *httpflow.HttpFlow) bool {
	h.EnsureRegex()

	method := flow.Request.Method
	path := flow.Request.URL.Path

	matchesPath := h.pathRegexp == nil || h.pathRegexp.MatchString(path)
	matchesMethod := h.Method == "*" || h.Method == "" || h.Method == method

	return matchesPath && matchesMethod
}

//////////////////////////////////
// Public Methods               //
//////////////////////////////////

func NewHttpValidator(method string, path string) HttpValidator {
	v := HttpValidator{
		Method: method,
		Path:   path,
	}
	v.EnsureRegex()
	return v
}

func NewHandler(method string, path string, handler func(flow *httpflow.HttpFlow)) Handler {
	return Handler{
		HttpValidator: NewHttpValidator(method, path),
		ServeHTTP:     handler,
	}
}

func NewMiddleware(method string, path string, priority int, handler func(flow *httpflow.HttpFlow)) Middleware {
	return Middleware{
		HttpValidator: NewHttpValidator(method, path),
		Action:        handler,
		Priority:      priority,
	}
}

func AddMiddleware(middleware Middleware) {
	globalMiddleware = append(globalMiddleware, middleware)
	sort.Slice(globalMiddleware, func(i, j int) bool {
		return globalMiddleware[i].Priority < globalMiddleware[j].Priority
	})
	log.Debug("AddMiddleware", "Middleware %s, all %s", middleware, globalMiddleware)
}

func AddHandler(pattern string, handler func(flow *httpflow.HttpFlow)) {
	method, path := patternToPathAndMethod(pattern)
	globalHandlers = append(globalHandlers, Handler{
		HttpValidator: NewHttpValidator(method, path),
		ServeHTTP:     handler,
	})
}

func GetAllMiddleware() []Middleware {
	return globalMiddleware
}

func GetAllHandlers() []Handler {
	return globalHandlers
}

//////////////////////////////////
// Private Methods              //
//////////////////////////////////

func patternToPathAndMethod(pattern string) (string, string) {
	tokens := strings.Split(pattern, " ")

	method := "GET"
	path := ""

	if len(tokens) > 1 {
		method = tokens[0]
		path = tokens[1]
	} else {
		path = tokens[0]
	}

	// Path must start with ^ or end with $ - if
	startsWithCaret := strings.HasPrefix(path, "^")
	endsWithDollar := strings.HasSuffix(path, "$")

	if !startsWithCaret || !endsWithDollar {
		path = "^" + path + "$"
	}

	return method, path
}
