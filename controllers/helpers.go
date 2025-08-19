package controllers

import (
	"encoding/json"
	"net/http"
)

// parseJSON decodes the JSON body of a request into the
// provided destination interface
func parseJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}
