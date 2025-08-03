package problemdetails

import (
	"encoding/json"
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/outdead/golibs/httpserver/validator"
)

type Binder struct {
	logger Logger
}

// NewBinder creates and returns new Binder instance.
func NewBinder(l Logger) *Binder {
	return &Binder{
		logger: l,
	}
}

// Bind parses echo Context to data structure. Returns en error if data is invalid.
func (b *Binder) Bind(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		var t *json.UnmarshalTypeError
		if ok := errors.As(err, &t); ok {
			return validator.NewValidationError(t.Field, err.Error())
		}

		// TODO: There is no way to get field names of int, time.Time, uuid.CartUUID if got incorrect data.
		return validator.NewValidationError("", err.Error())
	}

	return nil
}

// BindAndValidate parses echo Context to data structure and validate received
// data. Returns en error if data is invalid.
func (b *Binder) BindAndValidate(c echo.Context, req interface{}) error {
	if err := b.Bind(c, req); err != nil {
		return err
	}

	return c.Validate(req)
}
