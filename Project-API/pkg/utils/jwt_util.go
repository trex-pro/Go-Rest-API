package utils

import (
	"os"
	"project-api/internal/models"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func JWT(user *models.Exec) (string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	jwtExpiry := os.Getenv("JWT_EXPIRY")

	claims := jwt.MapClaims{
		"id":       user.ID,
		"username": user.Username,
		"role":     user.Role,
	}

	if jwtExpiry != "" {
		jwtDuration, err := time.ParseDuration(jwtExpiry)
		if err != nil {
			ErrorHandler(err, "JWT Expiration Error")
		}
		claims["expiry"] = jwt.NewNumericDate(time.Now().Add(jwtDuration))
	} else {
		claims["expiry"] = jwt.NewNumericDate(time.Now().Add(15 * time.Minute))
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := jwtToken.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", ErrorHandler(err, "Failed to Generate Token")
	}
	return token, nil
}
