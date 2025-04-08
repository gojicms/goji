package core

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
)

func resourceHandler(flow *httpflow.HttpFlow) {
	path := "web/" + flow.Request.URL.Path

	fileInfo, err := os.Stat(path)
	if err == nil {
		// Add Last-Modified header
		lastModified := fileInfo.ModTime().Format(http.TimeFormat)
		flow.SetHeader("Last-Modified", lastModified)

		// Check If-Modified-Since header
		ifModifiedSince := flow.Request.Header.Get("If-Modified-Since")
		if ifModifiedSince != "" {
			ifModTime, err := time.Parse(http.TimeFormat, ifModifiedSince)
			if err == nil && fileInfo.ModTime().Before(ifModTime.Add(1*time.Second)) {
				flow.WriteHeaders(304)
				return
			}
		}
	}

	data := utils.Object{}

	res, serveErr := server.RenderFile(path, server.RenderOptions{
		Data: data,
	})
	if serveErr != nil {
		res := server.RenderErrorPage(serveErr.HttpCode, serveErr.Message, server.RenderOptions{
			Data: data,
		})
		flow.WriteHeaders(res.HttpCode)
		_, _ = flow.Write(res.Body)
		return
	}

	flow.SetHeader("Content-Type", res.ContentType)
	flow.SetHeader("Cache-Control", "public, max-age=0, must-revalidate")
	flow.WriteHeaders(200)
	_, _ = flow.Write(res.Body)
}

var rootDocHandler = func(flow *httpflow.HttpFlow) {
	// Trim the leading slash from the path
	path := strings.TrimPrefix(flow.Request.URL.Path, "/")

	// If path is empty, default to index.html
	if path == "" || path == "/" {
		path = "index.html"
	}

	// Initialize variables to store file path and any captured parameters
	var filePath string
	var found bool

	// Check options in order of priority
	// 1. Check for /a/b/c/index.html
	indexPath := path
	if !strings.HasSuffix(indexPath, "/") {
		indexPath += "/"
	}
	indexPath += "index.html"

	if fileExists("web/" + indexPath) {
		filePath = indexPath
		found = true
	}

	// 2. Check for /a/b/c.html
	if !found {
		htmlPath := path
		if !strings.HasSuffix(htmlPath, ".html") {
			htmlPath += ".html"
		}

		if fileExists("web/" + htmlPath) {
			filePath = htmlPath
			found = true
		}
	}

	// 3. Check for dynamic routes like /a/b/{{id}}.html
	if !found {
		// Split the path into segments
		segments := strings.Split(path, "/")
		if len(segments) > 0 {
			// Try replacing the last segment with {{id}}
			dynamicSegments := make([]string, len(segments))
			copy(dynamicSegments, segments)

			// Store the potential ID value
			lastSegmentIndex := len(dynamicSegments) - 1
			potentialID := dynamicSegments[lastSegmentIndex]

			// Replace last segment with {{id}}
			dynamicSegments[lastSegmentIndex] = "{{id}}"
			dynamicPath := strings.Join(dynamicSegments, "/")

			if !strings.HasSuffix(dynamicPath, ".html") {
				dynamicPath += ".html"
			}

			if fileExists("web/" + dynamicPath) {
				filePath = dynamicPath
				found = true
				flow.Set("path_id", potentialID)
				flow.Append("templateData", "path_id", potentialID)
			}
		}
	}

	// If no matching file was found, return 404
	if !found {
		res := server.RenderErrorPage(404, "Page not found", server.RenderOptions{})
		flow.WriteHeaders(res.HttpCode)
		_, _ = flow.Write(res.Body)
		return
	}

	// Render the file
	res, err := server.RenderFile("web/"+filePath, server.RenderOptions{
		TemplateRoot: "web/!partials",
		Data:         flow.Get("templateData").(utils.Object),
	})

	if err != nil {
		res := server.RenderErrorPage(err.HttpCode, err.Message, server.RenderOptions{})
		flow.WriteHeaders(res.HttpCode)
		_, _ = flow.Write(res.Body)
		return
	}

	flow.SetHeader("Content-Type", res.ContentType)
	log.Info("Web/Core", "Rendering file %s as %s", filePath, res.ContentType)
	flow.WriteHeaders(200)
	_, _ = flow.Write(res.Body)
}

//////////////////////////////////
// Resource Definitions         //
//////////////////////////////////

var publicResResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodGet, "^/public/.+"),
	Handler:       resourceHandler,
}

var httpResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodGet, ".+"),
	Handler:       rootDocHandler,
}

//////////////////////////////////
// Private Methods              //
//////////////////////////////////

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
