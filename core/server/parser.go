package server

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gojicms/goji/core/config"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/utils"
	"github.com/gojicms/goji/core/utils/log"
	"github.com/google/uuid"
)

// TemplateConfig holds configuration for the template renderer
type TemplateConfig struct {
	Debug bool
	// PartialLoader is a function that loads partial templates by name
	// If nil, partial template inclusion won't work
	PartialLoader func(templatePath string) ([]byte, error)
}

// DefaultConfig is the default configuration
var DefaultConfig = TemplateConfig{
	Debug:         false,
	PartialLoader: nil,
}

type Script struct {
	Source string
	Type   string
}

func init() {
	extend.RegisterFunctions(template.FuncMap{
		// Objects
		"toJson": func(v interface{}) string {
			a, err := json.Marshal(utils.Object{"A": "B"})
			if err != nil {
				log.Warn("Parser", "Error converting object to json")
			}
			str := string(a)
			return str
		},
		"urlencode": func(v interface{}) string {
			a, err := url.QueryUnescape(v.(string))
			if err != nil {
				log.Warn("Parser", "Error encoding as URL")
			}
			return url.QueryEscape(a)
		},
		// String manipulation
		"concat": func(s ...string) string {
			var result string
			for _, str := range s {
				result += str
			}
			return result
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
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
		"script": func(s string, options string) template.HTML {
			optArray := strings.Split(options, " ")
			var extraParams []string

			for _, opt := range optArray {
				switch opt {
				case "module":
					extraParams = append(extraParams, `type="module"`)
				}
			}

			extraParamsString := strings.Join(extraParams, " ")
			return template.HTML(`<script lang="text/javascript" src="` + s + `" ` + extraParamsString + `></script>`)
		},
		// Date/Time
		"toFormattedDate": func(t time.Time, format string) string { return t.Format(format) },
		"toDateTime":      func(t time.Time) string { return t.Format(time.RFC1123) },
		"toDate":          func(t time.Time) string { return t.Format(time.DateOnly) },
		"toTime":          func(t time.Time) string { return t.Format(time.TimeOnly) },
		"toFuzzyTime":     func(t time.Time) string { return utils.FuzzyTime(t) },
		// Util
		"CSRF": func() string {
			return strings.ReplaceAll(uuid.New().String(), "-", "")
		},
	})
}

// RenderTemplate processes an HTML template string and replaces template expressions
// using Go's html/template package
func RenderTemplate(templateContent []byte, data interface{}, options RenderOptions) ([]byte, error) {
	templateConfig := TemplateConfig{
		Debug: config.ActiveConfig.Application.Debug,
		PartialLoader: func(templatePath string) ([]byte, error) {
			// Ensure the partial path is resolved relative to the public directory
			partialPath := filepath.Join(options.TemplateRoot, templatePath)
			partialContent, err := os.ReadFile(partialPath)
			if err != nil {
				log.Error("Parser", "Access to %s denied; %s", partialPath, err.Error())
				return nil, fmt.Errorf("failed to load partial %s: %w", templatePath, err)
			}
			return partialContent, nil
		},
	}

	// Create a new template with a unique name
	tmpl := template.New("inline")
	funcMap := extend.GlobalFunctions()

	// Add custom config methods
	for k, v := range options.Functions {
		funcMap[k] = v
	}

	// Add style injection support
	var styles []string
	funcMap["style"] = func(path string) (template.HTML, error) {
		styles = append(styles, path)
		return "", nil
	}
	funcMap["styles"] = func() (template.HTML, error) {
		var buf bytes.Buffer

		for _, style := range styles {
			buf.Write([]byte(`<link rel="stylesheet" href="` + style + `"/>`))
		}
		return template.HTML(buf.String()), nil
	}

	// Add script injection
	var scripts []Script
	funcMap["script"] = func(src string, scriptType string) (template.HTML, error) {
		scripts = append(scripts, Script{
			Source: src,
			Type:   scriptType,
		})
		return "", nil
	}
	funcMap["scripts"] = func() (template.HTML, error) {
		var buf bytes.Buffer
		for _, script := range scripts {
			buf.Write([]byte(`<script src="` + script.Source + `" type="` + script.Type + `"></script>`))
		}
		return template.HTML(buf.String()), nil
	}

	// Add the partial function if a PartialLoader is provided
	if templateConfig.PartialLoader != nil {
		funcMap["partial"] = func(partialPath string) (template.HTML, error) {
			content, err := templateConfig.PartialLoader(partialPath)
			if err != nil {
				if templateConfig.Debug {
					return template.HTML(fmt.Sprintf("<!-- Error loading partial %s: %v -->", partialPath, err)), nil
				}
				return "", err
			}

			// Create a sub-template for this partial
			subTmpl, err := template.New(partialPath).Funcs(funcMap).Parse(string(content))
			if err != nil {
				if templateConfig.Debug {
					return template.HTML(fmt.Sprintf("<!-- Error parsing partial %s: %v -->", partialPath, err)), nil
				}
				return "", err
			}

			// Execute the partial with the current data context
			var buf bytes.Buffer
			if err := subTmpl.Execute(&buf, data); err != nil {
				if templateConfig.Debug {
					return template.HTML(fmt.Sprintf("<!-- Error executing partial %s: %v -->", partialPath, err)), nil
				}
				return "", err
			}

			return template.HTML(buf.String()), nil
		}
	}

	// Apply all functions to the template
	tmpl = tmpl.Funcs(funcMap)
	tmpl = tmpl.Option("missingkey=zero")

	// Parse the main template
	parsedTmpl, err := tmpl.Parse(string(templateContent))
	if err != nil {
		if templateConfig.Debug {
			return []byte(fmt.Sprintf("Template parsing error: %v\n\nTemplate:\n%s", err, templateContent)), nil
		}
		return nil, fmt.Errorf("template parsing error: %w", err)
	}

	// Execute the template
	var buf bytes.Buffer
	if err := parsedTmpl.Execute(&buf, data); err != nil {
		if templateConfig.Debug {
			return []byte(fmt.Sprintf("Template execution error: %v\n\nTemplate:\n%s", err, templateContent)), nil
		}
		return nil, fmt.Errorf("template execution error: %w", err)
	}

	return buf.Bytes(), nil
}
