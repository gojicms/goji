package files

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
	"github.com/google/uuid"
	"golang.org/x/exp/maps"
)

//////////////////////////////////
// Types                        //
//////////////////////////////////

type fileService struct{}

// scriptInfo holds data for script rendering
type scriptInfo struct {
	Path       string
	Attributes string
}

//////////////////////////////////
// Service Interface Impl       //
//////////////////////////////////

func (p *fileService) Name() string {
	return "go-files"
}

func (p *fileService) Description() string {
	return "Default file handling service"
}

func (p *fileService) Priority() int {
	return 10
}

//////////////////////////////////
// Template Functions           //
//////////////////////////////////

// templateFuncs returns the map of template functions
func (p *fileService) templateFuncs(flow *httpflow.HttpFlow) template.FuncMap {
	// Define funcs map first so partial can reference it recursively
	funcs := template.FuncMap{
		// Objects
		"toJson": func(v interface{}) string {
			a, err := json.Marshal(v)
			if err != nil {
				log.Warn("Template", "Error converting object to json")
				return ""
			}
			return string(a)
		},
		"urlencode": func(v string) string {
			return url.QueryEscape(v)
		},

		// String manipulation
		"concat": func(s ...string) string {
			return strings.Join(s, "")
		},
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"md5": func(s string) string {
			h := md5.New()
			h.Write([]byte(s))
			return hex.EncodeToString(h.Sum(nil))
		},

		// HTML Manipulation
		"attr": func(s string) template.HTMLAttr {
			return template.HTMLAttr(s)
		},
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
		"style": func(path string) template.HTML {
			var currentStyles []string
			if existing := flow.Get("__files__styles__"); existing != nil {
				// Try to assert; if it fails, start fresh (overwrites incorrect type)
				if slice, ok := existing.([]string); ok {
					currentStyles = slice
				} else {
					log.Warn("Files", "[Style] Existing __files__styles__ was not a []string, overwriting.")
					// Initialize slice to allow append below
					currentStyles = []string{}
				}
			}
			// Get, Append, Set back
			flow.Set("__files__styles__", append(currentStyles, path))
			return ""
		},
		"styles": func() template.HTML {
			var buf bytes.Buffer
			rawStyles := flow.Get("__files__styles__")
			log.Debug("Files", "[Styles] Raw data retrieved from flow for __files__styles__: Type=%T, Value=%+v", rawStyles, rawStyles)

			if rawStyles != nil {
				styleSlice, ok := rawStyles.([]string)
				log.Debug("Files", "[Styles] Type assertion to []string successful: %t", ok)

				if ok {
					for _, style := range styleSlice {
						buf.WriteString(`<link rel="stylesheet" href="` + style + `"/>`)
					}
				} else {
					log.Warn("Files", "[Styles] Failed to assert retrieved style data as []string")
				}
			}
			return template.HTML(buf.String())
		},
		"scripts": func() template.HTML {
			var buf bytes.Buffer
			rawScripts := flow.Get("__files__scripts__")
			log.Debug("Files", "[Scripts] Raw data retrieved from flow for __files__scripts__: Type=%T, Value=%+v", rawScripts, rawScripts)

			if rawScripts != nil {
				scriptSlice, ok := rawScripts.([]scriptInfo)
				log.Debug("Files", "[Scripts] Type assertion to []scriptInfo successful: %t", ok)

				if ok {
					for _, script := range scriptSlice {
						// Use script.Path and script.Attributes
						buf.WriteString(fmt.Sprintf(`<script src="%s"%s></script>`, script.Path, script.Attributes))
					}
				} else {
					log.Warn("Files", "[Scripts] Failed to assert retrieved script data as []scriptInfo")
				}
			}
			return template.HTML(buf.String())
		},

		// Updated script function with optional CSV options
		"script": func(path string, options ...string) template.HTML {
			attributeList := []string{}
			if len(options) > 0 {
				optsStr := options[0]
				optsList := strings.Split(optsStr, ",")
				for _, opt := range optsList {
					trimmedOpt := strings.TrimSpace(opt)
					switch trimmedOpt {
					case "module":
						attributeList = append(attributeList, `type="module"`)
						// Add other attribute cases here
					}
				}
			}
			attrStr := strings.Join(attributeList, " ")
			if attrStr != "" {
				attrStr = " " + attrStr // Add leading space
			}

			newScript := scriptInfo{Path: path, Attributes: attrStr}

			var currentScripts []scriptInfo
			if existing := flow.Get("__files__scripts__"); existing != nil {
				if slice, ok := existing.([]scriptInfo); ok {
					currentScripts = slice
				} else {
					log.Warn("Files", "[Script] Existing __files__scripts__ was not a []scriptInfo, overwriting.")
					currentScripts = []scriptInfo{}
				}
			}
			// Get, Append, Set back
			flow.Set("__files__scripts__", append(currentScripts, newScript))
			return "" // Return empty string as we collect
		},

		// Date/Time
		"toFormattedDate": func(t time.Time, format string) string {
			return t.Format(format)
		},
		"toDateTime": func(t time.Time) string {
			return t.Format(time.RFC1123)
		},
		"toDate": func(t time.Time) string {
			return t.Format(time.DateOnly)
		},
		"toTime": func(t time.Time) string {
			return t.Format(time.TimeOnly)
		},
		"toFuzzyTime": utils.FuzzyTime,

		// Security
		"CSRF": func() string {
			return strings.ReplaceAll(uuid.New().String(), "-", "")
		},
	}

	// Add partial function, needs access to 'funcs' map
	funcs["partial"] = func(name string, data ...interface{}) (template.HTML, error) {
		log.Debug("Files", "Rendering partial: %s", name)
		var currentData interface{}
		if len(data) > 0 {
			currentData = data[0]
		} else {
			// If no data passed, use the data from the flow if available
			if flowData := flow.Get("templateData"); flowData != nil {
				currentData = flowData
			}
		}

		var partialContent []byte
		var err error
		partialFound := false

		log.Debug("Files", "[Partial] Searching for partial: %s/%s", flow.Get("__template__rootPath"), name)

		// 1. Check relative path (!partials/name.html)
		if rootPathVal := flow.Get("__template__rootPath"); rootPathVal != nil {
			if rootPath, ok := rootPathVal.(string); ok && rootPath != "" {
				relPartialPath := filepath.Join(rootPath, "!partials", name)
				log.Debug("Files", "[Partial] Checking relative path: %s", relPartialPath)
				partialContent, err = os.ReadFile(relPartialPath)
				if err == nil {
					log.Debug("Files", "[Partial] Found at relative path: %s", relPartialPath)
					partialFound = true
				} else if !os.IsNotExist(err) {
					log.Error("Files", "Error reading relative partial %s: %v", relPartialPath, err)
				} else {
					log.Debug("Files", "[Partial] Not found at relative path: %s", relPartialPath)
				}
			}
		}

		// 2. Check root path (!partials/name.html) if not found relatively
		if !partialFound {
			// TODO: Make root partial path configurable
			rootPartialPath := filepath.Join(".", "!partials", name)
			log.Debug("Files", "[Partial] Checking root path: %s", rootPartialPath)
			partialContent, err = os.ReadFile(rootPartialPath)
			if err == nil {
				log.Debug("Files", "[Partial] Found at root path: %s", rootPartialPath)
				partialFound = true
			} else if !os.IsNotExist(err) {
				log.Error("Files", "Error reading root partial %s: %v", rootPartialPath, err)
			} else {
				log.Debug("Files", "[Partial] Not found at root path: %s", rootPartialPath)
			}
		}

		if !partialFound {
			log.Warn("Files", "Partial not found after checking relative and root: %s", name)
			return template.HTML(fmt.Sprintf("<!-- PARTIAL %s NOT FOUND -->", name)), nil
		}

		// Create a new template instance for the partial to isolate parsing and execution
		partialTmpl := template.New(name).Funcs(funcs) // Pass funcs for recursion
		parsedPartialTmpl, err := partialTmpl.Parse(string(partialContent))
		if err != nil {
			log.Error("Files", "Error parsing partial %s: %v", name, err)
			return template.HTML(fmt.Sprintf("<!-- ERROR PARSING PARTIAL %s -->", name)), nil
		}

		// Execute the partial template
		var buf bytes.Buffer
		log.Debug("Files", "[Partial] Executing partial '%s' with data: %+v", name, currentData)
		if err := parsedPartialTmpl.Execute(&buf, currentData); err != nil {
			log.Error("Files", "Error executing partial %s: %v", name, err)
			return template.HTML(fmt.Sprintf("<!-- ERROR EXECUTING PARTIAL %s -->", name)), nil
		}

		return template.HTML(buf.String()), nil
	}

	return funcs
}

//////////////////////////////////
// Template Execution Methods   //
//////////////////////////////////

func (p *fileService) ExecuteTemplate(content []byte, flow *httpflow.HttpFlow) ([]byte, error) {
	// Note: No path context here means relative partials might not work as expected.
	log.Debug("Files", "Executing template without path context")

	if !bytes.Contains(content, []byte("{{")) {
		return content, nil
	}

	tmpl := template.New("main").Funcs(p.templateFuncs(flow))

	// Parse the main content FIRST
	var err error
	tmpl, err = tmpl.Parse(string(content))
	if err != nil {
		log.Error("Files", "Error parsing main template content: %v", err)
		return nil, err
	}

	// Prepare template data
	templateData := utils.Object{
		"goji": config.ActiveConfig.Cms,
	}
	if flowData := flow.Get("templateData"); flowData != nil {
		if data, ok := flowData.(utils.Object); ok {
			maps.Copy(templateData, data)
		} else {
			log.Warn("Files", "templateData in flow was not a utils.Object")
		}
	}

	// Execute the main template
	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "main", templateData); err != nil {
		log.Error("Files", "Error executing template 'main': %v", err)
		return nil, err
	}

	return buf.Bytes(), nil
}

func (p *fileService) ExecuteTemplateFromPath(path string, flow *httpflow.HttpFlow) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		log.Error("Files", "Could not read file for template execution: %s - %v", path, err)
		return nil, err
	}

	// Set the root path *before* executing the template so the partial function can use it
	flow.Set("__template__rootPath", filepath.Dir(path))

	return p.ExecuteTemplate(content, flow)
}

func (p *fileService) RenderTemplateFromPath(path string, flow *httpflow.HttpFlow) error {
	content, err := p.ExecuteTemplateFromPath(path, flow)
	if err != nil {
		log.Error("Files", "Error executing template from path %s: %v", path, err)
		flow.Writer.WriteHeader(http.StatusInternalServerError) // Use WriteHeader
		flow.Writer.Write([]byte("Internal Server Error rendering template"))
		return err // Return the original error
	}

	flow.SetHeader("Content-Type", "text/html; charset=utf-8") // Set Content-Type
	// WriteHeader(http.StatusOK) is called implicitly on first Write if not set
	_, err = flow.Writer.Write(content)
	if err != nil {
		log.Error("Files", "Error writing rendered template to response for %s: %v", path, err)
	}
	return err
}

func (p *fileService) RenderTemplate(template []byte, flow *httpflow.HttpFlow) error {
	content, err := p.ExecuteTemplate(template, flow)
	if err != nil {
		log.Error("Files", "Error executing template: %v", err)
		return err
	}

	flow.SetHeader("Content-Type", "text/html; charset=utf-8") // Set Content-Type
	flow.Writer.Write(content)

	return nil
}

//////////////////////////////////
// File Rendering Method        //
//////////////////////////////////

// RenderFile serves a static file or renders a template.
func (p *fileService) RenderFile(inPath string, flow *httpflow.HttpFlow) error {
	path := filepath.Clean(inPath)
	path = strings.TrimLeft(path, "/") // Ensure path is relative

	if !isValidPath(path) {
		log.Warn("Files", "Attempt to access invalid path: %s", path)
		flow.Writer.WriteHeader(http.StatusBadRequest)
		flow.Writer.Write([]byte("Invalid path"))
		return fmt.Errorf("invalid path requested: %s", path)
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Warn("Files", "File not found: %s", path)
			flow.Writer.WriteHeader(http.StatusNotFound)
			flow.Writer.Write([]byte("Not Found"))
			return err
		} else {
			log.Error("Files", "Error stating file %s: %v", path, err)
			flow.Writer.WriteHeader(http.StatusInternalServerError)
			flow.Writer.Write([]byte("Internal Server Error"))
			return err
		}
	}

	contentType := getContentType(path)

	// Serve non-HTML files, large files, or files without template markers directly
	if !isHTMLFile(path) || fileInfo.Size() >= config.ActiveConfig.Application.TemplateFileSizeLimit {
		log.Debug("Files", "Serving static file: %s", path)
		flow.SetHeader("Content-Type", contentType)
		http.ServeFile(flow.Writer, flow.Request, path)
		return nil // http.ServeFile handles errors internally somewhat
	}

	// Check content for template markers before reading again (optimization?)
	content, err := os.ReadFile(path)
	if err != nil {
		log.Error("Files", "Error reading file %s for templating: %v", path, err)
		flow.Writer.WriteHeader(http.StatusInternalServerError)
		flow.Writer.Write([]byte("Internal Server Error"))
		return err
	}

	if !bytes.Contains(content, []byte("{{")) {
		log.Debug("Files", "Serving HTML file directly (no markers): %s", path)
		flow.SetHeader("Content-Type", contentType)
		flow.Writer.Write(content)
		return nil
	}

	// Otherwise, render as template
	log.Debug("Files", "Rendering template file: %s", path)
	return p.RenderTemplateFromPath(path, flow)
}

//////////////////////////////////
// Helper Functions             //
//////////////////////////////////

// isValidPath ensures malicious paths aren't served
func isValidPath(path string) bool {
	// Prevent directory traversal and access to hidden (! prefix) files/dirs
	return !strings.Contains(path, "..") && !strings.Contains(path, "!")
}

// getContentType determines the MIME type based on file extension or content sniffing.
func getContentType(filePath string) string {
	extToMime := map[string]string{
		".html":   "text/html; charset=utf-8",
		".htm":    "text/html; charset=utf-8",
		".gohtml": "text/html; charset=utf-8",
		".css":    "text/css; charset=utf-8",
		".js":     "application/javascript",
		".json":   "application/json",
		".png":    "image/png",
		".jpg":    "image/jpeg",
		".jpeg":   "image/jpeg",
		".gif":    "image/gif",
		".svg":    "image/svg+xml",
		".pdf":    "application/pdf",
		".txt":    "text/plain; charset=utf-8",
		".xml":    "application/xml",
		".mp4":    "video/mp4",
		".webm":   "video/webm",
		".mp3":    "audio/mpeg",
		".wav":    "audio/wav",
		".ico":    "image/x-icon",
		".ttf":    "font/ttf",
		".otf":    "font/otf",
		".woff":   "font/woff",
		".woff2":  "font/woff2",
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if mime, ok := extToMime[ext]; ok {
		return mime
	}

	file, err := os.Open(filePath)
	if err == nil {
		defer file.Close()
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err == nil {
			return http.DetectContentType(buffer[:n])
		}
	}

	return "application/octet-stream"
}

// isHTMLFile checks if a file has an HTML extension.
func isHTMLFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".html" || ext == ".htm" || ext == ".gohtml"
}

//////////////////////////////////
// Plugin Definition           //
//////////////////////////////////

var Plugin = extend.PluginDef{
	Name:         "files",
	FriendlyName: "File Handler",
	Description:  "Handles file serving, caching, and template rendering",
	Internal:     true,
	Resources:    []extend.ResourceDef{},
	OnInit: func() error {
		// Register our file service
		services.RegisterService(&fileService{})
		return nil
	},
}
