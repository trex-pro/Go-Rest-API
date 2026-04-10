package utils

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"
)

func isValidSortOrder(order string) bool {
	return order == "asc" || order == "desc"
}

func isValidSortField(field string) bool {
	vaildFields := map[string]bool{
		"first_name": true,
		"last_name":  true,
		"email":      true,
		"class":      true,
		"subject":    true,
	}
	return vaildFields[field]
}

func GenerateInsertQuery(tableName string, model any) string {
	modelType := reflect.TypeOf(model)
	var columns, placeholders string
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		fmt.Println("dbTag:", dbTag)
		dbTag = strings.TrimSuffix(dbTag, ",omitempty")
		if dbTag != "" && dbTag != "id" {
			if columns != "" {
				columns += ", "
				placeholders += ", "
			}
			columns += dbTag
			placeholders += "?"
		}
	}
	return fmt.Sprintf("INSERT into %s (%s) VALUES (%s)", tableName, columns, placeholders)
}

func GetStructValues(model any) []any {
	modelValue := reflect.ValueOf(model)
	modelType := modelValue.Type()
	values := []any{}
	for i := 0; i < modelType.NumField(); i++ {
		dbTag := modelType.Field(i).Tag.Get("db")
		if dbTag != "" && dbTag != "id,omitempty" {
			values = append(values, modelValue.Field(i).Interface())
		}
	}
	log.Println("Values:", values)
	return values
}

func GetStudentFilter(r *http.Request, queryBuilder *strings.Builder) []any {
	var args []any
	params := []string{"first_name", "last_name", "email", "class", "subject"}
	for _, dbField := range params {
		value := r.URL.Query().Get(dbField)
		if value != "" {
			queryBuilder.WriteString(" AND " + dbField + " = ?")
			args = append(args, value)
		}
	}
	return args
}

func GetStudentSort(r *http.Request, queryBuilder *strings.Builder) {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		queryBuilder.WriteString(" ORDER BY")
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if isValidSortField(field) && isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(" " + field + " " + order)
		}
	}
}

func GetTeacherFilter(r *http.Request, queryBuilder *strings.Builder) []any {
	var args []any
	params := []string{"first_name", "last_name", "email", "class", "subject"}
	for _, dbField := range params {
		value := r.URL.Query().Get(dbField)
		if value != "" {
			queryBuilder.WriteString(" AND " + dbField + " = ?")
			args = append(args, value)
		}
	}
	return args
}

func GetTeacherSort(r *http.Request, queryBuilder *strings.Builder) {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		queryBuilder.WriteString(" ORDER BY")
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if isValidSortField(field) && isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(" " + field + " " + order)
		}
	}
}

func GetExecFilter(r *http.Request, queryBuilder *strings.Builder) []any {
	var args []any
	params := []string{"id", "first_name", "last_name", "email", "username", "user_created_at", "inactive_status", "role"}
	for _, dbField := range params {
		value := r.URL.Query().Get(dbField)
		if value != "" {
			queryBuilder.WriteString(" AND " + dbField + " = ?")
			args = append(args, value)
		}
	}
	return args
}

func GetExecSort(r *http.Request, queryBuilder *strings.Builder) {
	sortParams := r.URL.Query()["sortby"]
	if len(sortParams) > 0 {
		queryBuilder.WriteString(" ORDER BY")
		for i, param := range sortParams {
			parts := strings.Split(param, ":")
			if len(parts) != 2 {
				continue
			}
			field, order := parts[0], parts[1]
			if isValidSortField(field) && isValidSortOrder(order) {
				continue
			}
			if i > 0 {
				queryBuilder.WriteString(", ")
			}
			queryBuilder.WriteString(" " + field + " " + order)
		}
	}
}
