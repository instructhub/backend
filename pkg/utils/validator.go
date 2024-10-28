package utils

import "github.com/go-playground/validator/v10"

// For check if the user request is vaild
func Validator() *validator.Validate {
	validate := validator.New()
	return validate
}
