package endpoint

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// decode decodes a value from JSON and returns it.
func decode[T any](r *http.Request) (T, error) {
	var v T
	decoder := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return v, fmt.Errorf("decode json: multiple values are not allowed")
		}
		return v, fmt.Errorf("decode json: %w", err)
	}
	return v, nil
}
