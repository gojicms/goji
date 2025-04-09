package server

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
)

type HttpServeError struct {
	HttpCode int
	Message  string
}

type HttpServeResponse struct {
	Body        []byte
	ContentType string
	HttpCode    int
}

type RenderOptions struct {
	SkipValidation bool         // SkipValidation If the user provides any portion of the path, this should be used with caution; allows traversal through relative and up-dirs.
	Data           utils.Object // Data Functions and values to be included in the template render
	TemplateRoot   string       // TemplateRoot the root folder for partials
	ErrorRoot      string
	Functions      template.FuncMap
}

var DefaultRenderOptions = RenderOptions{
	SkipValidation: false,
	Data:           utils.Object{},
	TemplateRoot:   "",
	ErrorRoot:      "",
	Functions:      template.FuncMap{},
}

// IsValidPath Ensure malicious paths aren't served, this includes
// ..., paths that start with /, path that include ! (used for private files)
func IsValidPath(path string) bool {
	return !(strings.HasPrefix(path, "..") || strings.Contains(path, "!"))
}

// RenderFile Renders a given HTML file using the template engine
func RenderFile(inPath string, options RenderOptions) (*HttpServeResponse, *HttpServeError) {
	path := filepath.Clean(inPath)
	path = strings.TrimLeft(path, "/")
	contentType := getContentType(path)

	if !options.SkipValidation && !IsValidPath(path) {
		log.Error("HTTP", "Preventing malicious access to %s", path)
		return nil, &HttpServeError{http.StatusBadRequest, "Invalid resource path."}
	}

	log.Info("HTTP", "Fetching file at %s", path)

	// Serve non-HTML files directly
	if !isHTMLFile(path) {
		file, err := os.ReadFile(path)
		if err != nil {
			log.Warn("HTTP", "File %s does not exist.", path)
			return &HttpServeResponse{[]byte("404"), contentType, 404}, nil
		}
		return &HttpServeResponse{file, contentType, 200}, nil
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		log.Error("HTTP", "Access to %s denied; file not found", path)
		return nil, &HttpServeError{http.StatusNotFound, "The requested resource could not be found."}
	}

	// Files that exceed template file size limit are rendered as-is
	if fileInfo.Size() >= config.ActiveConfig.Application.TemplateFileSizeLimit {
		file, _ := os.ReadFile(path)
		return &HttpServeResponse{file, contentType, 200}, nil
	}

	// Read the file content
	content, err := os.ReadFile(path)
	if err != nil {
		log.Error("HTTP", "Access to %s denied; file not found", path)
		return nil, &HttpServeError{http.StatusInternalServerError, "Error accessing the requested file."}
	}

	// Check if the content contains template markers
	// Only process if needed (contains {{ or <partial)
	if !bytes.Contains(content, []byte("{{")) && !bytes.Contains(content, []byte("<partial")) {
		// No template markers found, serve directly
		return &HttpServeResponse{content, contentType, 200}, nil
	}

	log.Info("HTTP", "Rendering %s", path)
	return RenderString(content, contentType, options)
}

func RenderString(content []byte, contentType string, options RenderOptions) (*HttpServeResponse, *HttpServeError) {
	renderedTemplate, err := RenderTemplate(content, options.Data, options)
	if err != nil {
		return nil, &HttpServeError{http.StatusInternalServerError, "Error rendering template."}
	}

	return &HttpServeResponse{renderedTemplate, contentType, 200}, nil
}

// RenderErrorPage serveErrorPage serves an error page with the appropriate status code
// DEPRECATED: This will be removed in favor of a new approach soon.
func RenderErrorPage(statusCode int, message string, options RenderOptions) *HttpServeResponse {
	errorRoot := "public/"
	if options.ErrorRoot != "" {
		errorRoot = options.ErrorRoot
	}

	// Try to load a custom error page for this status code
	var templateContent []byte
	errorPagePath := fmt.Sprintf(errorRoot+"/%d.html", statusCode)
	content, err := os.ReadFile(errorPagePath)

	if err == nil {
		// Use the custom error page if it exists
		templateContent = content
	} else {
		// Fall back to a generic error template
		content, _ := os.ReadFile("server/error_template.html")
		templateContent = content
	}

	// Prepare template data
	templateData := utils.Object{
		"status":  statusCode,
		"message": message,
		"goji":    config.ActiveConfig.Cms,
	}

	// Merge incoming template data in
	for key, value := range options.Data {
		templateData[key] = value
	}

	renderedTemplate, err := RenderTemplate(templateContent, templateData, options)
	if err != nil {
		return nil
		// TODO: This is probably wrong, fix
	}

	return &HttpServeResponse{renderedTemplate, "text/html; charset=utf-8", statusCode}
}

// getContentType returns the MIME type based on file extension
func getContentType(filePath string) string {
	// Common file extensions and their MIME types
	extToMime := map[string]string{
		".html":  "text/html; charset=utf-8",
		".htm":   "text/html; charset=utf-8",
		".css":   "text/css; charset=utf-8",
		".js":    "application/javascript",
		".json":  "application/json",
		".png":   "image/png",
		".jpg":   "image/jpeg",
		".jpeg":  "image/jpeg",
		".gif":   "image/gif",
		".svg":   "image/svg+xml",
		".pdf":   "application/pdf",
		".txt":   "text/plain; charset=utf-8",
		".xml":   "application/xml",
		".mp4":   "video/mp4",
		".webm":  "video/webm",
		".mp3":   "audio/mpeg",
		".wav":   "audio/wav",
		".ico":   "image/x-icon",
		".ttf":   "font/ttf",
		".otf":   "font/otf",
		".woff":  "font/woff",
		".woff2": "font/woff2",
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if mime, ok := extToMime[ext]; ok {
		return mime
	}

	// If extension not found in map, read a bit of the file to detect content type
	file, err := os.Open(filePath)
	if err == nil {
		defer file.Close()

		// Read the first 512 bytes to detect content type
		buffer := make([]byte, 512)
		n, err := file.Read(buffer)
		if err == nil {
			return http.DetectContentType(buffer[:n])
		}
	}

	// Default to octet-stream if detection fails
	return "application/octet-stream"
}

// isHTMLFile checks if a file has an HTML extension
func isHTMLFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return ext == ".html" || ext == ".htm" || ext == ".gohtml"
}
