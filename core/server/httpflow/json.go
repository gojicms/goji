package httpflow

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gojicms/goji/core/utils"
)

func WriteErrorJson(flow *HttpFlow, status int, message string, data ...any) {
	flow.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	flow.Writer.WriteHeader(status)
	WriteJson(flow, utils.Object{
		"error": fmt.Sprintf(message, data...),
	})
}

func (f *HttpFlow) WriteErrorJson(status int, message string, data ...any) {
	WriteErrorJson(f, status, message, data...)
}

func WriteUnauthorizedJson(flow *HttpFlow) {
	flow.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")
	flow.Writer.WriteHeader(http.StatusUnauthorized)
	WriteJson(flow, utils.Object{
		"error": "unauthorized",
	})
}

func WriteJsonList[T any](flow *HttpFlow, limit int, offset int, total int, collectionName string, data *[]T) {
	var nextUrl = flow.Request.URL
	var prevUrl = flow.Request.URL

	nextOffset := offset + limit
	prevOffset := offset - limit

	if prevOffset < 0 {
		prevOffset = 0
	}

	// Modify the query for nextUrl
	nextQuery := nextUrl.Query()
	nextQuery.Set("offset", strconv.Itoa(nextOffset))
	nextUrl.RawQuery = nextQuery.Encode()

	// Modify the query for prevUrl
	prevQuery := prevUrl.Query()
	prevQuery.Set("offset", strconv.Itoa(prevOffset))
	prevUrl.RawQuery = prevQuery.Encode()

	links := utils.Object{}

	if total == 0 || limit+offset < total {
		links["next"] = nextUrl.String()
	}

	if offset > 0 {
		links["prev"] = prevUrl.String()
	}

	WriteJson(flow, utils.Object{
		"_itemCount":   len(*data),
		"_totalCount":  total,
		collectionName: data,
		"links":        links,
	})
}

func (f *HttpFlow) WriteJsonList(limit int, offset int, total int, collectionName string, data *[]any) {
	WriteJsonList[any](f, limit, offset, total, collectionName, data)
}

func WriteJson(flow *HttpFlow, data interface{}) {
	w := flow.Writer
	jsonResponse, err := json.Marshal(data)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(jsonResponse)
}

func (f *HttpFlow) WriteJson(data interface{}) {
	WriteJson(f, data)
}
