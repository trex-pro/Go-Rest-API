package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"project-api/internal/api/middlewares"
	"strings"
	"time"
)

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("This is ROOT route."))
}

func studentHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("This is STUDENTS GET method route."))
	case http.MethodPost:
		w.Write([]byte("This is STUDENTS POST method route."))
	case http.MethodPut:
		w.Write([]byte("This is STUDENTS PUT method route."))
	case http.MethodPatch:
		w.Write([]byte("This is STUDENTS PATCH method route."))
	case http.MethodDelete:
		w.Write([]byte("This is STUDENTS DELETE method route."))
	default:
		w.Write([]byte("This is STUDENTS route."))
	}
}

func teacherHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		path := strings.TrimPrefix(r.URL.Path, "/teachers/")
		userID := strings.TrimSuffix(path, "/")
		fmt.Println("User ID:", userID)

		queryParams := r.URL.Query()
		sortby := queryParams.Get("sortby")
		sortorder := queryParams.Get("sortorder")
		fmt.Printf("SortBy: %v, SortOrder: %v\n", sortby, sortorder)

		w.Write([]byte("This is TEACHERS GET method route."))
	case http.MethodPost:
		w.Write([]byte("This is TEACHERS POST method route."))
	case http.MethodPut:
		w.Write([]byte("This is TEACHERS PUT method route."))
	case http.MethodPatch:
		w.Write([]byte("This is TEACHERS PATCH method route."))
	case http.MethodDelete:
		w.Write([]byte("This is TEACHERS DELETE method route."))
	default:
		w.Write([]byte("This is TEACHERS route."))
	}
}

func execHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("This is EXECS GET method route."))
	case http.MethodPost:
		fmt.Println("Query:", r.URL.Query())
		fmt.Println("Name:", r.URL.Query().Get("name"))
		fmt.Println("Age:", r.URL.Query().Get("age"))
		fmt.Println("City:", r.URL.Query().Get("city"))

		// Parse From Data for x-www-form-urlencoded.
		err := r.ParseForm()
		if err != nil {
			return
		}
		w.Write([]byte("This is EXECS POST method route."))
	case http.MethodPut:
		w.Write([]byte("This is EXECS PUT method route."))
	case http.MethodPatch:
		w.Write([]byte("This is EXECS PATCH method route."))
	case http.MethodDelete:
		w.Write([]byte("This is EXECS DELETE method route."))
	default:
		w.Write([]byte("This is EXECS route."))
	}
}

func main() {
	port := 3000
	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/students/", studentHandler)
	mux.HandleFunc("/teachers/", teacherHandler)
	mux.HandleFunc("/execs/", execHandler)

	// Load TLS Certificate and Key.
	cert := "cert.pem"
	key := "key.pem"

	// TLS Config.
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
	}

	// Initialize Rate Limiter.
	rl := middlewares.RateLimiter(5, time.Minute)

	// HPP Options.
	hppOptions := middlewares.HPPOptions{
		CheckBody:                true,
		CheckBodyforConetentType: "application/x-www-form-urlencoded",
		CheckQuery:               true,
		WhiteList:                []string{"sortBy", "sortOrder", "name", "age", "city", "class"},
	}

	// Middlewares.
	secureMux := middlewares.HPP(hppOptions)(rl.RateLimiterMiddleware(middlewares.Compression(middlewares.ResponseTimer(middlewares.SecurityHeader(middlewares.CORS(mux))))))

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
