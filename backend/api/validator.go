package api

import (
	"github.com/go-playground/validator/v10"
	"simple_bank/pkg"
)

var validCurrency validator.Func = func(fl validator.FieldLevel) bool {
	if currency, ok := fl.Field().Interface().(string); ok {
		return pkg.IsSupportedCurrency(currency)
	}
	return false
}
