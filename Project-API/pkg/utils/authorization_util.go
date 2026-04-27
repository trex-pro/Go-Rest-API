package utils

import (
	"errors"
	"slices"
)

type ContextKey string

func AuthorizeExec(execRole string, allowedRoles ...string) (bool, error) {
	if !slices.Contains(allowedRoles, execRole) {
		return false, errors.New("Exec Not Authorized")
	}
	return true, nil
}
