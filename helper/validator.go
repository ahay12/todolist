package helper

import "github.com/go-playground/validator/v10"

var validate *validator.Validate

func validateHandler() {
	validate = validator.New(validator.WithPrivateFieldValidation())

}
