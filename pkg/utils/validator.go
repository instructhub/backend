package utils

import "github.com/go-playground/validator/v10"

func Validator() *validator.Validate {
	validate := validator.New()
	return validate
}
