package server

import (
	"net/http"

	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils/log"
)

var ServerMux = &GojiServerMux{
	mux: http.NewServeMux(),
}

type GojiServerMux struct {
	mux *http.ServeMux
}

// Handle is a lightweight API for handling callbacks
func (server *GojiServerMux) Handle(pattern string, handler func(flow *httpflow.HttpFlow)) {
	extend.AddHandler(pattern, handler)
}

// HandleFunc mimics the traditional method in Go - primarily to allow compatibility with methods that expect
// the traditional syntax.
func (server *GojiServerMux) HandleFunc(pattern string, handler func(res http.ResponseWriter, req *http.Request)) {
	extend.AddHandler(pattern, func(flow *httpflow.HttpFlow) {
		handler(flow.Writer, flow.Request)
	})
}

func (server *GojiServerMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flow := &httpflow.HttpFlow{
		Writer:  w,
		Request: r,
	}

	log.Debug("Server", "Checking Path %s", flow.Request.URL.Path)

	for _, middleware := range extend.GetAllMiddleware() {
		if middleware.CanHandle(flow) {
			middleware.Action(flow)
			log.Debug("Server", "Running Middleware %s", middleware.Path)
			if flow.HasTerminated() {
				return
			}
		}
	}

	// We'll check services first; in the future, we'll cache requests to which resource/handler they may be tied to
	for _, service := range extend.GetPlugins() {
		log.Debug("Server", "Plugin %s", service.FriendlyName)
		for _, resource := range service.Resources {
			log.Debug("Server", "Resource %s", resource.Path)
			if resource.CanHandle(flow) {
				log.Debug("Server", "Running Plugin %s > Resource %s", service.FriendlyName, resource.Path)
				resource.Handler(flow)
				return
			}
		}
	}

	for _, handler := range extend.GetAllHandlers() {
		if handler.CanHandle(flow) {
			log.Debug("Server", "Running Handler %s", handler.Path)
			handler.ServeHTTP(flow)
			return
		}
	}

	log.Warn("Server", "No handler found")
	flow.WriteErrorJson(http.StatusNotFound, "Not found")
}
