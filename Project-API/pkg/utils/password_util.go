package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func VerifyPassword(password, encodedHash string) error {
	parts := strings.Split(encodedHash, ".")
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

	hashedPassword := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	if subtle.ConstantTimeCompare(hashedPassword, hash) != 1 {
		return ErrorHandler(errors.New("Password Hash Length does not Match"), "Incorrect Credentials")
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if password == "" {
		return "", ErrorHandler(errors.New("Blank Password"), "Please Enter Password.")
	}
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", ErrorHandler(errors.New("Failed to Generate Salt"), "Error Addding Data to DB.")
	}
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	saltBase64 := base64.StdEncoding.EncodeToString(salt)
	hashBase64 := base64.StdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf("%s.%s", saltBase64, hashBase64)
	return encodedHash, nil
}
