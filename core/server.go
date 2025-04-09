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
	"github.com/gojicms/goji/core/plugins/admin"
	"github.com/gojicms/goji/core/plugins/core"
	"github.com/gojicms/goji/core/plugins/files"
	"github.com/gojicms/goji/core/plugins/health"
	"github.com/gojicms/goji/core/plugins/sessions"
	"github.com/gojicms/goji/core/plugins/site"
	"github.com/gojicms/goji/core/plugins/users"
	"github.com/gojicms/goji/core/server"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils/log"
)

//////////////////////////////////
// Types                        //
//////////////////////////////////

//////////////////////////////////
// Public Methods               //
//////////////////////////////////

// PrepareServer sets up the Goji server by providing necessary configuration,
// checking certain configuration for sanity, and adding the mandatory core plugins.
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

	// The health plugin checks certain things to alert the user of potential issues.
	extend.RegisterPlugin(&health.Plugin)

	// The templates plugin handles template rendering
	extend.RegisterPlugin(&files.Plugin)

	// The Admin plugin handles administration features; In the future - when plugins
	// are hot loaded - it will offer the ability to be disabled/enabled while running for added security
	extend.RegisterPlugin(&admin.Plugin)

	// The Auth plugin handles the ability to log in and out
	extend.RegisterPlugin(&users.Plugin)

	// Core lays the framework for basic APIs; Technically, despite its name, it isn't strictly
	// necessary and in fact it MAY work without it, but since core does handle the public web
	// side of things, this limits you to the core CMS functionality; in a headless setup,
	// this could be disabled - but do note that service discovery is a part of this.
	extend.RegisterPlugin(&core.Plugin)

	// Sessions manages authentication sessions
	extend.RegisterPlugin(&sessions.Plugin)

	// Site allows configuring and writing core site details
	extend.RegisterPlugin(&site.Plugin)

	// Load dynamic modules after core plugins
	if err := loadDynamicModules(); err != nil {
		log.Error("Core", "Failed to load dynamic modules: %v", err)
	}
}

func StartServer() {
	if !config.ActiveConfig.Cms.Configured {
		log.Error("Core", "Configuration not configured. Invoke PrepareServer before calling StartServer.")
		return
	}

	for _, plugin := range extend.GetPlugins() {
		log.Info("Core", "Starting plugin "+plugin.Name)
		if plugin.OnInit == nil {
			log.Fatal(log.RCServicesConfig, "Core", "Plugin "+plugin.Name+" has no OnInit function")
		}
		err := plugin.OnInit()
		if err != nil {
			log.Fatal(log.RCServicesConfig, "Core", "Failed to initialize plugin %s (%s) - please ensure a valid configuration is provided.", plugin.FriendlyName, plugin.Name)
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

// loadDynamicModules loads all .so files from the modules directory and registers their plugins.
// This allows for dynamic loading of additional functionality without requiring a server restart.
// Each module must export a PluginService symbol that is a pointer to a PluginDef.
func loadDynamicModules() error {
	modulesDir := "modules"
	if err := os.MkdirAll(modulesDir, 0755); err != nil {
		return err
	}

	files, err := filepath.Glob(filepath.Join(modulesDir, "*.so"))
	if err != nil {
		return err
	}

	var loadedPlugins []*extend.PluginDef
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

		plugin, ok := sym.(*extend.PluginDef)
		if !ok {
			log.Error("Core", "Module %s PluginService is not a PluginDef", file)
			continue
		}

		extend.RegisterPlugin(plugin)
		loadedPlugins = append(loadedPlugins, plugin)
		log.Info("Core", "Loaded module %s with plugin %s", file, plugin.Name)
	}

	if len(loadedPlugins) > 0 {
		log.Info("Core", "Successfully loaded %d modules:", len(loadedPlugins))
		for _, plugin := range loadedPlugins {
			log.Info("Core", "  - %s (%s)", plugin.FriendlyName, plugin.Name)
		}
	} else {
		log.Info("Core", "No modules found in modules directory")
	}

	return nil
}
