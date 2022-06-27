package api

import (
	"election/util"

	"github.com/go-playground/validator/v10"
)

var validDob validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if dob, ok := fieldLevel.Field().Interface().(string); ok {
		return util.IsDateOfBirth(dob)
	}
	return false
}
