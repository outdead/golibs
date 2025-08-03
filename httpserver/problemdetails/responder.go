package problemdetails

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Responder wraps on echo Context to be used on echo HTTP handlers to
// construct an HTTP response.
type Responder struct {
	logger Logger
}

// NewResponder creates and returns pointer to Responder.
func NewResponder(l Logger) *Responder {
	return &Responder{l}
}

// ServeResult sends a JSON response with the result data.
func (r *Responder) ServeResult(c echo.Context, i interface{}) error {
	return c.JSON(http.StatusOK, i)
}

// ServeError sends a JSON error response with status code.
func (r *Responder) ServeError(c echo.Context, err error, logPrefix ...string) error {
	if err == nil {
		return nil
	}

	body := NewError(err)

	if body.Status == StatusInternalServerError {
		if len(logPrefix) != 0 {
			err = fmt.Errorf("%s: %w", logPrefix[0], err)
		}

		r.logger.Error(err.Error())
	}

	return c.JSON(body.Status, body)
}
