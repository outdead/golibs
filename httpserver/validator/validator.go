package validator

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

// Validator is a wrapper for external validation package. Allows extending validation rules.
type Validator struct {
	v *validator.Validate
}

// New returns a new instance of Validator with sane defaults.
func New() *Validator {
	return &Validator{
		v: validator.New(),
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
		valErr := NewValidationError(vErr.Tag(), msg)
		fields = append(fields, *valErr)
	}

	return fields
}
