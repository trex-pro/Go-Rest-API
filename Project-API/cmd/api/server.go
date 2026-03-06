package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"project-api/internal/api/middlewares"
	"project-api/internal/api/router"
)

func main() {
	port := 3000

	// Load TLS Certificate and Key.
	cert := "cert.pem"
	key := "key.pem"

	// TLS Config.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Initialize Rate Limiter.
	// rl := middlewares.RateLimiter(5, time.Minute)

	// HPP Options.
	// hppOptions := middlewares.HPPOptions{
	// 	CheckBody:                true,
	// 	CheckBodyforConetentType: "application/x-www-form-urlencoded",
	// 	CheckQuery:               true,
	// 	WhiteList:                []string{"sortBy", "sortOrder", "name", "age", "class"},
	// }

	// Middlewares.
	// secureMux := applyMiddlewares(mux, middlewares.HPP(hppOptions),
	// 	middlewares.Compression,
	// 	middlewares.SecurityHeader,
	// 	middlewares.ResponseTimer,
	// 	rl.RateLimiterMiddleware,
	// 	middlewares.CORS)

	router := router.Router()
	secureMux := middlewares.SecurityHeader(router)

	// Custom HTTPS Server.
	server := http.Server{
		Addr:      fmt.Sprintf(":%d", port),
		Handler:   secureMux,
		TLSConfig: tlsConfig,
	}

	fmt.Println("Server is running on port:", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}
}
