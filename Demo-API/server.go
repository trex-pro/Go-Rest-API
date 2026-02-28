package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		logRequestDetails(r)
		fmt.Fprintf(w, "Handling incoming orders.\n")
	})

	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		logRequestDetails(r)
		fmt.Fprintf(w, "Handling incoming users.\n")
	})

	port := 3000

	// Load TLS Certificate and key.
	cert := "cert.pem"
	key := "key.pem"

	// TLS Config.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Custom HTTP Server.
	server := http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      nil,
		TLSConfig:    tlsConfig,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	// Enable HTTP2 Package.
	// http2.ConfigureServer(&server, &http2.Server{})

	fmt.Println("Server is running in port:", port)
	err := server.ListenAndServeTLS(cert, key)
	if err != nil {
		log.Fatalln("Error starting server:", err)
	}

	// HTTP 1.1 without TLS.
	// err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	// if err != nil {
	// 	log.Fatalln("Error starting server:", err)
	// }
}

func logRequestDetails(r *http.Request) {
	httpVersion := r.Proto
	fmt.Println("Received request with HTTP version:", httpVersion)

	if r.TLS != nil {
		tlsVersion := getTLSVersion(r.TLS.Version)
		fmt.Println("Received request with TLS version:", tlsVersion)
	} else {
		fmt.Println("Received request without TLS.")
	}
}

func getTLSVersion(version uint16) string {
	switch version {
	case tls.VersionTLS10:
		return "TLS 1.0"
	case tls.VersionTLS11:
		return "TLS 1.1"
	case tls.VersionTLS12:
		return "TLS 1.2"
	case tls.VersionTLS13:
		return "TLS 1.3"
	default:
		return "Unknown TLS"
	}
}
