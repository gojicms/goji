package plugin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gojicms/goji/contrib/documents/admin"
	"github.com/gojicms/goji/contrib/documents/documents"
	"github.com/gojicms/goji/core/database"
	"github.com/gojicms/goji/core/extend"
	"github.com/gojicms/goji/core/server"
	"github.com/gojicms/goji/core/server/httpflow"
	"github.com/gojicms/goji/core/utils"
)

//////////////////////////////////
// Plugin Service               //
//////////////////////////////////

var PluginService extend.ServiceDef = Service

//////////////////////////////////
// Resource Definitions         //
//////////////////////////////////

var addDocResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodPost, "/api/v1/docs"),
	Description:   "Adds a new document",
	Handler: func(flow *httpflow.HttpFlow) {
		r := flow.Request
		w := flow.Writer

		var doc documents.Document
		if err := utils.DecodeJSONBody(r, &doc); err != nil {
			server.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		addedDoc, err := documents.Create(doc)
		if err != nil {
			server.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(addedDoc); err != nil {
			http.Error(w, "Failed to encode response JSON: "+err.Error(), http.StatusInternalServerError)
		}
	},
}

var deleteDocResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodDelete, "/api/v1/docs"),
	Description:   "Deletes a document",
	Handler: func(flow *httpflow.HttpFlow) {
		r := flow.Request
		w := flow.Writer

		id := r.PathValue("id")

		if id == "" {
			server.WriteError(w, http.StatusUnprocessableEntity, "Id is required")
			return
		}

		count, err := documents.DeleteById(id)
		if err != nil {
			server.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		if count == 0 {
			server.WriteError(w, http.StatusNotFound, "Document with the ID %s not found", id)
			return
		}

		server.WriteJson(w, utils.Object{"count": count})
	},
}

var getDocResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodGet, "/api/v1/docs/.+"),
	Description:   "Returns a single document",
	Handler: func(flow *httpflow.HttpFlow) {
		r := flow.Request
		w := flow.Writer

		id := r.PathValue("id")
		if id == "" {
			server.WriteError(w, http.StatusUnprocessableEntity, "Id is required")
			return
		}

		doc, err := documents.GetById(id)
		if err != nil {
			server.WriteError(w, http.StatusNotFound, "document not found")
			return
		}

		server.WriteJson(w, doc)
	},
}

var getDocsResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodGet, "/api/v1/docs"),
	Description:   "Displays a list of documents",
	Handler: func(flow *httpflow.HttpFlow) {
		r := flow.Request
		w := flow.Writer

		limit := utils.OrDefault(r.URL.Query().Get("limit"), "10")
		offset := utils.OrDefault(r.URL.Query().Get("offset"), "0")

		limitInt, err := strconv.Atoi(limit)
		if err != nil {
			server.WriteError(w, http.StatusUnprocessableEntity, "Limit must be an integer")
			return
		}
		offsetInt, err := strconv.Atoi(offset)
		if err != nil {
			server.WriteError(w, http.StatusUnprocessableEntity, "Offset must be an integer")
			return
		}

		docs, err := documents.Get(limitInt, offsetInt, "updated_at desc")
		if err != nil {
			server.WriteError(w, http.StatusInternalServerError, err.Error())
		}

		count, err := documents.Count()

		httpflow.WriteJsonList(flow, limitInt, offsetInt, int(count), "docs", &docs)
	},
}

var updateDocResource = extend.ResourceDef{
	HttpValidator: extend.NewHttpValidator(http.MethodPost, "/api/v1/docs/.+"),
	Description:   "Updates a document",
	Handler: func(flow *httpflow.HttpFlow) {
		r := flow.Request
		w := flow.Writer

		id := r.PathValue("id")

		doc, err := documents.GetById(id)

		if err := utils.DecodeJSONBody(r, &doc); err != nil {
			server.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		addedDoc, err := documents.Update(*doc)
		if err != nil {
			server.WriteError(w, http.StatusInternalServerError, err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(addedDoc); err != nil {
			http.Error(w, "Failed to encode response JSON: "+err.Error(), http.StatusInternalServerError)
		}
	},
}

//////////////////////////////////
// Service Definition           //
//////////////////////////////////

var Service = extend.ServiceDef{
	Name:         "documents",
	FriendlyName: "Documents",
	Resources: []extend.ResourceDef{
		getDocsResource,
		getDocResource,
		addDocResource,
		deleteDocResource,
		updateDocResource,
	},
	OnInit: func() error {
		admin.Init()

		database.AutoMigrate(&documents.Document{})

		extend.RegisterFunction("docs", func(limit int, offset int, sort string) []documents.Document {
			docs, _ := documents.Get(limit, offset, sort)
			return docs
		})
		extend.RegisterFunction("doc", func(id string) documents.Document {
			doc, _ := documents.GetById(id)
			return *doc
		})
		return nil
	},
}
