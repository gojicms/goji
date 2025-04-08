package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func DecodeJSONBody[T any](r *http.Request, dst *T) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return fmt.Errorf("failed to read the request body: %v", err)
	}
	defer r.Body.Close() // Ensure the body is closed after reading

	err = json.Unmarshal(body, dst)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	return nil
}
