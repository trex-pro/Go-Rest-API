package router

import (
	"net/http"
	"project-api/internal/api/handlers"
)

func studentsRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Students.
	mux.HandleFunc("GET /students", handlers.GETStudentsHandler)
	mux.HandleFunc("POST /students", handlers.POSTStudentsHandler)
	mux.HandleFunc("PUT /students", handlers.PUTStudentsHandler)
	mux.HandleFunc("PATCH /students", handlers.PATCHStudentsHandler)
	mux.HandleFunc("DELETE /students", handlers.DELETEStudentsHandler)

	mux.HandleFunc("GET /students/{id}", handlers.GETStudentByIDHandler)
	mux.HandleFunc("PUT /students/{id}", handlers.PUTStudentsHandler)
	mux.HandleFunc("PATCH /students/{id}", handlers.PATCHStudentByIDHandler)
	mux.HandleFunc("DELETE /students/{id}", handlers.DELETEStudentByIDHandler)

	return mux
}
