package admin

import (
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/services"
)

func renderError(flow *httpflow.HttpFlow, status int, message string) {
	flow.Append("templateData", "error", message)

	fileService := services.GetServiceOfType[services.FileService]("files")
	if err := fileService.RenderTemplateFromPath("admin/500.html", flow); err != nil {
		// If we can't render the error page, fall back to plain text
		flow.WriteHeaders(status)
		flow.SetHeader("Content-Type", "text/plain")
		_, _ = flow.Write([]byte(message))
	}
}
