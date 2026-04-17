package utils

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"project-api/internal/models"
	"strings"

	"golang.org/x/crypto/argon2"
)

func Password(user *models.Exec, req models.Exec) error {
	parts := strings.SplitN(user.Password, ".", 2)
	if len(parts) != 2 {
		return ErrorHandler(errors.New("Error Spliting Password Hash"), "Internal Server Error")
	}

	saltBase64 := parts[0]
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return ErrorHandler(err, "Internal Server Error")
	}
	hashBase64 := parts[1]
	hash, err := base64.StdEncoding.DecodeString(hashBase64)
	if err != nil {
		return ErrorHandler(err, "Internal Server Error")
	}

	hashedPassword := argon2.IDKey([]byte(req.Password), salt, 1, 64*1024, 4, 32)
	if subtle.ConstantTimeCompare(hashedPassword, hash) != 1 {
		return ErrorHandler(errors.New("Password Hash Length does not Match"), "Incorrect Credentials")
	}
	return nil
}
