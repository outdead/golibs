package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator is a wrapper for external validation package. Allows extending validation rules.
type Validator struct {
	v *validator.Validate
}

// New returns a new instance of Validator with sane defaults.
func New() *Validator {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return &Validator{
		v: v,
	}
}

// Validate implement Validator interface.
func (v *Validator) Validate(i interface{}) error {
	err := v.v.Struct(i)
	if err == nil {
		return nil
	}

	var vErrs validator.ValidationErrors
	if !errors.As(err, &vErrs) {
		return err
	}

	fields := make(ValidationErrors, 0, len(vErrs))

	for _, vErr := range vErrs {
		msg := fmt.Sprintf("invalid on '%s' rule", vErr.Tag())
		valErr := NewValidationError(vErr.Field(), msg)
		fields = append(fields, *valErr)
	}

	return fields
}
