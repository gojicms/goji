package admin

import (
	_ "embed"
	"fmt"
	"net/http"

	"github.com/gojicms/goji/contrib/documents/documents"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/types"
	. "github.com/gojicms/goji/core/utils"
)

//go:embed listing.gohtml
var listTemplate []byte

//go:embed editor.gohtml
var editorTemplate []byte

func adminListing(flow *httpflow.HttpFlow) ([]byte, error) {
	flow.Append("templateData", "title", "Goji - Documents")

	offset := OrDefault(flow.Request.URL.Query().Get("offset"), "0")
	count := OrDefault(flow.Request.URL.Query().Get("count"), "10")

	offsetInt := Stoid(offset, 0)
	countInt := Stoid(count, 10)

	items, _ := documents.Get(countInt, offsetInt, "updated_at desc")
	totalItems, _ := documents.Count()

	return server.RenderTemplate(listTemplate, Object{
		"items":      items,
		"totalItems": totalItems,
		"offset":     offsetInt,
		"count":      countInt,
	}, server.DefaultRenderOptions)
}

func newDocEditor(flow *httpflow.HttpFlow) ([]byte, error) {
	flow.Append("templateData", "title", "Goji - Create Document")

	document := documents.Document{}

	result := Object{
		"status":  nil,
		"message": nil,
	}

	if flow.Request.Method == "POST" {
		user := flow.Get("user").(*types.User)

		result["status"] = "success"
		result["message"] = "Document created."

		title := flow.PostFormValue("title")
		content := flow.PostFormValue("content")

		if title == "" {
			result["status"] = "error"
			result["message"] = "Title cannot be empty."
			goto render
		}

		document.Title = title
		document.Content = content
		document.CreatedBy = user

		doc, err := documents.Create(document)

		if err != nil {
			result["status"] = "error"
			result["message"] = "Failed to save document: " + err.Error()
			goto render
		}

		if doc != nil {
			flow.SetHeader("Location", fmt.Sprintf("/admin/docs/%d", doc.ID))
			flow.WriteHeaders(http.StatusFound)
			return []byte{}, nil
		}
	}

render:
	return server.RenderTemplate(editorTemplate, Object{
		"document": document,
		"result":   result,
	}, server.DefaultRenderOptions)

}

func editDocEditor(flow *httpflow.HttpFlow) ([]byte, error) {
	flow.Append("templateData", "title", "Goji - Edit Document")

	id := flow.GetKvp("admin_meta", "id")
	document, err := documents.GetById(id)
	if err != nil {
		return nil, err
	}

	result := Object{
		"status":  nil,
		"message": nil,
	}

	if flow.Request.Method == "POST" {
		action := flow.PostFormValue("action")

		switch action {
		case "save":
			result["status"] = "success"
			result["message"] = "Document saved."

			title := flow.PostFormValue("title")
			content := flow.PostFormValue("content")

			if title == "" {
				result["status"] = "error"
				result["message"] = "Title cannot be empty."
				goto render
			}

			document.Title = title
			document.Content = content

			_, err = documents.Update(*document)
			if err != nil {
				result["status"] = "error"
				result["message"] = "Failed to save document: " + err.Error()
			}
			break
		case "delete":
			_, err := documents.DeleteById(id)
			if err != nil {
				result["status"] = "error"
				result["message"] = "Failed to delete document: " + err.Error()
				goto render
			}
			flow.SetHeader("Location", "/admin/docs")
			flow.WriteHeaders(http.StatusFound)
			return []byte{}, nil
		}
	}
render:
	return server.RenderTemplate(editorTemplate, Object{
		"document": document,
		"result":   result,
	}, server.DefaultRenderOptions)
}

func Init() {
	extend.AddSideMenuItem("Documents", "docs", 10, "", "document:view")
	extend.AddSideMenuItem("Create", "docs/new", 10, "Documents", "document:add")

	extend.AddAdminPage(extend.AdminPage{
		Route:  "docs",
		Render: adminListing,
	})

	extend.AddAdminPage(extend.AdminPage{
		Route:  "docs/new",
		Render: newDocEditor,
	})

	extend.AddAdminPage(extend.AdminPage{
		Route:  "docs/{id}",
		Render: editDocEditor,
	})
}
