/*
httpflow is a series of utilities that simplify a start-to-finish flow of an HTTP request, allowing data to be
passed through the flow and utilities for properly responding to certain events that occur.
*/

package httpflow

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gojicms/goji/core/utils"
)

type HttpFlow struct {
	Writer     http.ResponseWriter
	Request    *http.Request
	terminated bool
	data       utils.Object
	mu         sync.RWMutex
}

// Data setters/getters

func (f *HttpFlow) Set(key string, value any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.data == nil {
		f.data = utils.Object{}
	}
	f.data[key] = value
}

func (f *HttpFlow) Get(key string) any {
	f.mu.RLock()
	defer f.mu.RUnlock()
	val, ok := f.data[key]
	if !ok {
		return nil
	}
	return val
}

func (f *HttpFlow) GetKvp(key string, subkey string) string {
	f.mu.RLock()
	defer f.mu.RUnlock()
	data := f.data[key].(utils.Object)
	return data[subkey].(string)
}

func (f *HttpFlow) Has(key string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if f.data == nil {
		return false
	}

	// This is the correct way to check if a key exists in a map
	value, exists := f.data[key]
	return exists && value != nil
}

func (f *HttpFlow) Del(key string) {
	delete(f.data, key)
}

// Append adds or updates a nested value in a utils.Object at the specified key/subkey
// Append adds or updates a nested value in a utils.Object at the specified key/subkey
func (f *HttpFlow) Append(key string, subkey string, value any) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Initialize f.data if it's nil
	if f.data == nil {
		f.data = utils.Object{}
	}

	var data utils.Object

	// Check if the key exists and is a utils.Object
	if existingValue, exists := f.data[key]; exists {
		// Try to convert to utils.Object
		if existingObj, ok := existingValue.(utils.Object); ok {
			// Use the existing object
			data = existingObj
		} else {
			// Key exists but isn't a utils.Object, replace with new one
			data = utils.Object{}
		}
	} else {
		// Key doesn't exist, create new utils.Object
		data = utils.Object{}
	}

	// Set the subkey in the object
	data[subkey] = value

	// Update the main data store
	f.data[key] = data
}

// HTTP Methods

func (f *HttpFlow) SetCookie(h *http.Cookie) {
	http.SetCookie(f.Writer, h)
}

func (f *HttpFlow) Redirect(s string, found int) {
	http.Redirect(f.Writer, f.Request, s, found)

}

func (f *HttpFlow) WriteHeaders(code int) {
	f.Writer.WriteHeader(code)
}

func (f *HttpFlow) Write(body []byte) (int, error) {
	return f.Writer.Write(body)
}

func (f *HttpFlow) SetHeader(s string, value string) {
	f.Writer.Header().Set(s, value)
}

func (f *HttpFlow) PostFormValue(s string) string {
	return f.Request.PostFormValue(s)
}

func (f *HttpFlow) DecodeJSONBody(m *map[string]interface{}) error {
	return utils.DecodeJSONBody(f.Request, m)
}

// Terminate sets a flag to indicate further processing should be stopped. This is intended for fatal
// but manageable conditions within middleware.
func (f *HttpFlow) Terminate() {
	f.terminated = true
}

func (f *HttpFlow) HasTerminated() bool {
	return f.terminated
}

// GetContentType returns the MIME type based on file extension; if the type is not determined, we will use
// http.DetectContentType to attempt to identify the content
func GetContentType(filePath string) string {
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
		defer func(file *os.File) {
			_ = file.Close()
		}(file)

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
