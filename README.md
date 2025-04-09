# Goji

---

Goji is a content management platform written in Go with the intent of
replicating the ease of installation and setup of many PHP-based CMS systems
but written in a modern, performant language.

## Services
Features in Goji are "services" - generally isolated chunks of 
functionality that utilize the extension system of Goji to add 
functionality to both the admin panel as well as the front end.

Services are responsible for adding themselves to administration
panels, for exposing methods to public templates, and databases.
That said, tools are provided by Goji to simplify the process
of creating tables, linking them to existing tables, so on.

There are services that are required for Goji that are loaded automatically:

* admin - The Admin service handles the admin panel including rendering services
          that expose administrative tools.
* auth - The auth service handles authentication and user session management.
* core - Core handles the root API route and web root.

Additional services are available in the `contrib` module:

* docs - Adds support for documents, which can be seen as identical to 
         WordPress posts.

## Getting Started

There are two main ways to get started with Goji:

### 1. Using the Application Template (Recommended)

The simplest way is to copy the `Application` directory from the Goji repository. This provides:

- A pre-configured project structure
- The admin interface already set up
- Basic configuration files
- Example templates

Simply:
1. Copy the `Application` directory
2. Run `go mod tidy` to install dependencies
3. Start customizing!

### 2. Custom Integration

For more control, you can create your own Go project and integrate Goji manually:

1. Create a new Go project:
   ```bash
   mkdir my-goji-site
   cd my-goji-site
   go mod init mysite
   ```

2. Add Goji as a dependency:
   ```bash
   go get github.com/gojicms/goji/core
   ```

3. Create a main.go file:
   ```go
   package main

   import (
        "github.com/gojicms/goji/core"
        "github.com/gojicms/goji/core/config"
        // Choose a database of your liking
        _ "github.com/mattn/go-sqlite3"
        // And the database to go with it
        "gorm.io/driver/postgres"
        // And we'll need the dialector - this will be simplified in future versions
        "gorm.io/gorm"
   )

   func main() {
        app := core.PrepareServer(config.ApplicationConfig{
            Host: "0.0.0.0",
            Pepper: "<a value of your choosing>",
            Database: config.DatabaseConfig{
                Connector: func() gorm.Dialector {
                    return postgres.Open(utils.GetEnv("DB_DSN", "<default if env not set>"))
                }
            }
        })
        core.StartServer()
   }
   ```

4. Create required directories:
   ```bash
   mkdir -p templates/admin templates/site public
   ```

5. Copy the files from this repositories application/admin into the admin folder


6. Run your application:
   ```bash
   go run main.go
   ```
   
   This will start the Goji server. By default it runs on port 8080, so you can access it at http://localhost:8080

This gives you a minimal Goji installation that you can build upon.

## Basics

### Web Directory Structure

The web directory in Goji follows these conventions:

- Folders starting with `!` are not served publicly
- The `!partials` folder stores partial templates that can be included using `{{partial "path"}}`
  - Example: `{{partial "header"}}` will load `!partials/header.gohtml`

This applies to both the main web directory and plugin admin interfaces. Plugins can expose variables that will be available to templates, with debugging tools planned to help view this information.

### Plugins

Goji supports dynamic loading of plugins via `.so` files. To create a plugin:

1. Create a new Go module for your plugin
2. Define your service and export it as `PluginService`:
   ```go
   var PluginService *extend.ServiceDef = &ServiceDef{
       Name: "my-plugin",
       FriendlyName: "My Plugin",
       OnInit: func() error {
           // Initialize your plugin
           return nil
       }
   }
   ```
3. Build your plugin as a shared object:
   ```bash
   go build -buildmode=plugin -o my-plugin.so
   ```
4. Place the .so file in the `modules` directory next to your Goji binary

Goji will automatically detect and load plugins from the modules directory on startup. The plugin must export a `PluginService` symbol that is a pointer to a `ServiceDef`. The service will be registered and initialized along with core services.

Each plugin can:
- Register routes and handlers
- Add template functions
- Extend the admin interface
- Access core services and database

Plugins are loaded after core services but before the server starts, ensuring they have access to all core functionality.


