package main

import (
	"crypto/tls"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"project-api/internal/api/middlewares"
	"project-api/internal/api/router"
	"project-api/pkg/utils"
	"time"

	"github.com/joho/godotenv"
)

//go:embed .env
var envFile embed.FS

func loadEnvEmbeddedFile() {
	content, err := envFile.ReadFile(".env")
	if err != nil {
		log.Fatalf("Error Reading .env File: %v", err)
	}

	// Creating temp file.
	temp, err := os.CreateTemp("", ".env")
	if err != nil {
		log.Fatalf("Error Creating temp File: %v", err)
	}
	defer os.Remove(temp.Name())

	// Wrte contents of env variables to temp file.
	_, err = temp.Write(content)
	if err != nil {
		log.Fatalf("Error Writing to temp File: %v", err)
	}
	err = temp.Close()
	if err != nil {
		log.Fatalf("Error Closing temp File: %v", err)
	}

	// Loading env variables to temp file.
	err = godotenv.Load(temp.Name())
	if err != nil {
		log.Fatalf("Error Loading .env File: %v", err)
	}
}

func main() {
	// Only in Development Stage to run source code.
	loadEnvEmbeddedFile()

	port := os.Getenv("SERVER_PORT")

	// Load TLS Certificate and Key.
	cert := os.Getenv("CERT_FILE")
	key := os.Getenv("KEY_FILE")

	// TLS Config.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	rl := middlewares.RateLimiter(5, time.Minute)
	hppOptions := middlewares.HPPOptions{
		CheckBody:                true,
		CheckBodyforConetentType: "application/x-www-form-urlencoded",
		CheckQuery:               true,
		WhiteList:                []string{"sortBy", "sortOrder", "name", "class"},
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
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}
}
