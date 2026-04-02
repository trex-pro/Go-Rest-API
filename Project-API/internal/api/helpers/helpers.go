package helpers

import (
	"errors"
	"project-api/pkg/utils"
	"reflect"
	"strings"
)

func GetFieldNames(model any) []string {
	val := reflect.TypeOf(model)
	fields := []string{}

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		fieldToAdd := strings.TrimSuffix(field.Tag.Get("json"), ",omitempty")
		fields = append(fields, fieldToAdd)
	}
	return fields
}

func CheckingBlankFields(value any) error {
	val := reflect.ValueOf(value)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String && field.String() == "" {
			return utils.ErrorHandler(errors.New("All Fields are Required"), "All Fields are Required.")
		}
	}
	return nil
}
