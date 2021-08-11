package middleware

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	val *validator.Validate
}

var Val Validator

func init() {
	v := validator.New()
	v.RegisterValidation("validRoles", func(fl validator.FieldLevel) bool {
		inter := fl.Field()
		roles, ok := inter.Interface().([]string)
		if !ok {
			return false
		}
		return IsValidRole(roles)
	})

	Val.val = v
}

func (u Validator) Validate(i interface{}) error {
	return u.val.Struct(i)
}

func GetErrorString(err error) string {
	errorM := ""
	for _, err := range err.(validator.ValidationErrors) {
		e := err.Error()
		n := strings.Split(e, "failed")
		errorM += n[0]
	}
	return errorM
}

func IsValidRole(role []string) bool {
	for _, v := range role {
		switch strings.ToLower(v) {
		case
			"admin",
			"student",
			"teacher":
			return true
		}
	}
	return false
}
