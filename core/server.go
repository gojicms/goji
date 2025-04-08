package core

import (
	"net/http"
	"net/http/pprof"
	"os"
	"path/filepath"
	"plugin"

	"dario.cat/mergo"
	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/health"
	"github.com/gojicms/goji/core/server"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services/admin"
	"github.com/gojicms/goji/core/services/auth"
	"github.com/gojicms/goji/core/services/core"
	"github.com/gojicms/goji/core/services/sessions"
	"github.com/gojicms/goji/core/services/site"
	"github.com/gojicms/goji/core/utils/log"
)

//////////////////////////////////
// Types                        //
//////////////////////////////////

//////////////////////////////////
// Public Methods               //
//////////////////////////////////

// PrepareServer sets up the Goji server by providing necessary configuration,
// checking certain configuration for sanity, and adding the mandatory core services.
func PrepareServer(inConfig config.ApplicationConfig) {
	// Update the server config to use the provided app config; this is required
	err := mergo.Merge(&config.ActiveConfig.Application, inConfig, mergo.WithOverride, mergo.WithoutDereference)
	if err != nil {
		log.Error("Core", "Failed to merge config - please ensure a valid configuration is provided.")
		return
	}
	config.ActiveConfig.Cms.Configured = true
	if config.ActiveConfig.Application.LogLevel != 0 {
		log.Level = config.ActiveConfig.Application.LogLevel
	}

	// The health service checks certain things to alert the user of potential issues.
	extend.RegisterService(&health.Service)

	// The Admin service handles administration features; In the future - when services
	// are hot loaded - it will offer the ability to be disabled/enabled while running for added security
	extend.RegisterService(&admin.Service)

	// The Auth service handles the ability to log in and out
	extend.RegisterService(&auth.Service)

	// Core lays the framework for basic APIs; Technically, despite its name, it isn't strictly
	// necessary and in fact it MAY work without it, but since core does handle the public web
	// side of things, this limits you to the core CMS functionality; in a headless setup,
	// this could be disabled - but do note that service discovery is a part of this.
	extend.RegisterService(&core.Service)

	// Sessions manages authentication sessions
	extend.RegisterService(&sessions.Service)

	// Site allows configuring and writing core site details
	extend.RegisterService(&site.Service)

	// Load dynamic modules after core services
	if err := loadDynamicModules(); err != nil {
		log.Error("Core", "Failed to load dynamic modules: %v", err)
	}
}

func StartServer() {
	if !config.ActiveConfig.Cms.Configured {
		log.Error("Core", "Configuration not configured. Invoke PrepareServer before calling StartServer.")
		return
	}

	for _, service := range extend.GetServices() {
		log.Info("Core", "Starting service "+service.Name)
		if service.OnInit == nil {
			log.Fatal(log.RCServicesConfig, "Core", "Service "+service.Name+" has no OnInit function")
		}
		err := service.OnInit()
		if err != nil {
			log.Fatal(log.RCServicesConfig, "Core", "Failed to initialize service %s (%s) - please ensure a valid configuration is provided.", service.FriendlyName, service.Name)
		}
	}

	host := config.ActiveConfig.Application.Host
	port := config.ActiveConfig.Application.Port

	log.Log("Core", "Goji Server ")
	log.Success("Core", "Server listening on %s:%s", host, port)

	serverInst := &http.Server{
		Addr:    host + ":" + port,
		Handler: server.ServerMux,
	}

	server.ServerMux.Handle("/", func(flow *httpflow.HttpFlow) {
		//		server.WriteTemplate(r, w, templates.NotFoundTemplate, nil)
	})

	if config.ActiveConfig.Application.Debug {
		server.ServerMux.HandleFunc("GET /debug/pprof/", pprof.Index)
		server.ServerMux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
		server.ServerMux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
		server.ServerMux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
		server.ServerMux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)
		server.ServerMux.Handle("GET /debug/config", func(flow *httpflow.HttpFlow) {
			flow.WriteJson(config.ActiveConfig)
		})
	}

	if err := serverInst.ListenAndServe(); err != nil {
		log.Error("Core", "Error listening on %s:%s", host, port)
	}
}

//////////////////////////////////
// Private Methods              //
//////////////////////////////////

// loadDynamicModules loads all .so files from the modules directory and registers their services.
// This allows for dynamic loading of additional functionality without requiring a server restart.
// Each module must export a PluginService symbol that is a pointer to a ServiceDef.
func loadDynamicModules() error {
	modulesDir := "modules"
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(modulesDir, "*.so"))
	if err != nil {
		return err
	}

	var loadedServices []*extend.ServiceDef
	for _, file := range files {
		p, err := plugin.Open(file)
		if err != nil {
			log.Error("Core", "Failed to load module %s: %v", file, err)
			continue
		}

		sym, err := p.Lookup("PluginService")
		if err != nil {
			log.Error("Core", "Module %s does not export PluginService: %v", file, err)
			continue
		}

		service, ok := sym.(*extend.ServiceDef)
		if !ok {
			log.Error("Core", "Module %s PluginService is not a ServiceDef", file)
			continue
		}

		extend.RegisterService(service)
		loadedServices = append(loadedServices, service)
		log.Info("Core", "Loaded module %s with service %s", file, service.Name)
	}

	if len(loadedServices) > 0 {
		log.Info("Core", "Successfully loaded %d modules:", len(loadedServices))
		for _, service := range loadedServices {
			log.Info("Core", "  - %s (%s)", service.FriendlyName, service.Name)
		}
	} else {
		log.Info("Core", "No modules found in modules directory")
	}

	return nil
}
