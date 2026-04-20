package middlewares

import (
	"context"
	"errors"
	"net/http"
	"os"
	"project-api/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := r.Cookie("JWT")
		if err != nil {
			utils.ErrorHandler(err, "")
			http.Error(w, "Session Expired, Please Login Again", http.StatusUnauthorized)
			return
		}
		jwtSecret := os.Getenv("JWT_SECRET")

		parsedToken, err := jwt.Parse(token.Value, func(token *jwt.Token) (any, error) {
			// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
			return []byte(jwtSecret), nil
		}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				utils.ErrorHandler(err, "")
				http.Error(w, "Token Expired", http.StatusUnauthorized)
				return
			} else if errors.Is(err, jwt.ErrTokenMalformed) {
				utils.ErrorHandler(err, "")
				http.Error(w, "Token Malformed", http.StatusUnauthorized)
				return
			}
			utils.ErrorHandler(err, "")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if !parsedToken.Valid {
			utils.ErrorHandler(err, "")
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		claims, ok := parsedToken.Claims.(jwt.MapClaims)
		if !ok {
			utils.ErrorHandler(err, "")
			http.Error(w, "Invalid Token", http.StatusUnauthorized)
			return
		}

		// Using Context for carrying claims across the API.
		ctx := context.WithValue(r.Context(), "role", claims["role"])
		ctx = context.WithValue(ctx, "id", claims["id"])
		ctx = context.WithValue(ctx, "username", claims["username"])
		ctx = context.WithValue(ctx, "expiry", claims["expiry"])

		next.ServeHTTP(w, r)
	})
}
