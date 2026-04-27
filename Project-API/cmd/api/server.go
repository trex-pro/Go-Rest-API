package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"project-api/internal/api/middlewares"
	"project-api/internal/api/router"
	"project-api/internal/repositories/sqlconnect"
	"project-api/pkg/utils"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		return
	}
	_, err = sqlconnect.ConnectDB()
	if err != nil {
		utils.ErrorHandler(err, "Error Connecting to DB")
		return
	}

	port := os.Getenv("SERVER_PORT")

	// Load TLS Certificate and Key.
	cert := "cert.pem"
	key := "key.pem"

	// TLS Config.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	rl := middlewares.RateLimiter(5, time.Minute)
	hppOptions := middlewares.HPPOptions{
		CheckBody:                true,
		CheckBodyforConetentType: "application/x-www-form-urlencoded",
		CheckQuery:               true,
		WhiteList:                []string{"sortBy", "sortOrder", "name", "age", "class"},
	}
	const (
		PathLogin    = "/execs/login"
		PathForgotPW = "/execs/forgotpassword"
		PathResetPW  = "/execs/resetpassword/reset"
	)

	routers := router.MainRouter()
	jwtMiddleware := middlewares.ExcludePathMiddleware(middlewares.JWTMiddleware, PathLogin, PathForgotPW, PathResetPW)

	// CORS → RateLimiter → JWT → SecurityHeader → HPP → Compression → ResponseTimer
	secureMux := utils.ApplyMiddlewares(routers,
		middlewares.ResponseTimer,
		middlewares.Compression,
		middlewares.HPP(hppOptions),
		middlewares.SecurityHeader,
		jwtMiddleware,
		rl.RateLimiterMiddleware,
		middlewares.CORS,
	)

	// Custom HTTPS Server.
	server := http.Server{
		Addr:      port,
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port", port)
	err = server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}
}
