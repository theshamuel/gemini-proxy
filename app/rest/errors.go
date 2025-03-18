package rest

import (
	"github.com/go-chi/render"
	"log"
	"net/http"
)

// nolint:revive
const (
	ErrServerInternal = 0 // server internal error
	ErrJSONDecode     = 1 // failed unmarshalling incoming request
)

// SendErrorJSON create response JSON in schema  {error: err, details: more details, code: 1} json body and responds with error code
func SendErrorJSON(w http.ResponseWriter, r *http.Request, httpStatusCode int, err error, errCode int, details string) {
	log.Printf("[WARN] %d, %v, %d, %s ", httpStatusCode, err, errCode, details)
	render.Status(r, httpStatusCode)
	render.JSON(w, r, map[string]interface{}{"error": err.Error(), "code": errCode, "details": details})
}
