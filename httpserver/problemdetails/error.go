package problemdetails

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/outdead/golibs/httpserver/validator"
)

// Error types for RFC-7807 Problem Details format.
// See: https://datatracker.ietf.org/doc/html/rfc7807
const (
	TypeValidation          = "validation-error"
	TypeUnauthorized        = "unauthorized"
	TypeForbidden           = "forbidden"
	TypeNotFound            = "data-not-found"
	TypeInternalServerError = "internal-server-error"
	TypeHTTP                = "http-error"
)

// Titles for RFC-7807 Problem Details format.
const (
	TitleValidation          = "Your request parameters didn't validate."
	TitleUnauthorized        = "Your request has not been applied."
	TitleForbidden           = "Your request has been forbidden."
	TitleNotFound            = "Not Found"
	TitleInternalServerError = "Internal Server Error"
)

// Statuses for RFC-7807 Problem Details format.
const (
	StatusValidation          = http.StatusBadRequest          // 400
	StatusUnauthorized        = http.StatusUnauthorized        // 401
	StatusForbidden           = http.StatusForbidden           // 403
	StatusNotFound            = http.StatusNotFound            // 404
	StatusBusinessError       = http.StatusUnprocessableEntity // 422
	StatusInternalServerError = http.StatusInternalServerError // 500
)

// Error contains error in RFC-7807 Problem Details format.
type Error struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail,omitempty"`
	// InvalidParams contains the explanation of errors in RFC-7807 format.
	InvalidParams validator.ValidationErrors `json:"invalid-params,omitempty"` //nolint
}

// Error represents an error condition, with the nil value representing no error.
func (e *Error) Error() string {
	msg := fmt.Sprintf("%d %s: %s", e.Status, e.Type, e.Title)
	if len(e.InvalidParams) != 0 {
		msg += " - " + e.InvalidParams.Error()
	}

	return msg
}

func NewError(err error) *Error {
	body := NewInternalServerError()

	if ok := errors.As(err, &body); !ok {
		var (
			validationError  *validator.ValidationError
			validationErrors validator.ValidationErrors
			httpError        *echo.HTTPError
		)

		switch {
		case errors.As(err, &validationError):
			body = NewValidationError(*validationError)
			if validationError.Name == "" {
				body.Detail = validationError.Reason
			}
		case errors.As(err, &validationErrors):
			body = NewValidationError(validationErrors...)
		case errors.As(err, &httpError):
			body = NewHTTPError(httpError)
		case errors.Is(err, validator.ErrNotFound):
			body = NewNotFoundError(err)
		}
	}

	return body
}

func NewValidationError(err ...validator.ValidationError) *Error {
	return &Error{
		Type:          TypeValidation,
		Title:         TitleValidation,
		Status:        StatusValidation,
		Detail:        "",
		InvalidParams: err,
	}
}

func NewUnauthorizedError(err ...validator.ValidationError) *Error {
	return &Error{
		Type:          TypeUnauthorized,
		Title:         TitleUnauthorized,
		Status:        StatusUnauthorized,
		Detail:        "",
		InvalidParams: err,
	}
}

func NewForbiddenError(err ...validator.ValidationError) *Error {
	return &Error{
		Type:          TypeForbidden,
		Title:         TitleForbidden,
		Status:        StatusForbidden,
		Detail:        "",
		InvalidParams: err,
	}
}

func NewNotFoundError(err error) *Error {
	return &Error{
		Type:   TypeNotFound,
		Title:  TitleNotFound,
		Status: StatusNotFound,
		Detail: err.Error(),
	}
}

func NewInternalServerError() *Error {
	return &Error{
		Type:   TypeInternalServerError,
		Title:  TitleInternalServerError,
		Status: StatusInternalServerError,
		Detail: TitleInternalServerError,
	}
}

func NewHTTPError(err *echo.HTTPError) *Error {
	detail, _ := err.Message.(string)

	return &Error{
		Type:   TypeHTTP,
		Title:  http.StatusText(err.Code),
		Status: err.Code,
		Detail: detail,
	}
}
