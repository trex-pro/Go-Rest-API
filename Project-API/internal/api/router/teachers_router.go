package router

import (
	"net/http"
	"project-api/internal/api/handlers"
)

func teachersRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Teachers.
	mux.HandleFunc("GET /teachers", handlers.GETTeachersHandler)
	mux.HandleFunc("POST /teachers", handlers.POSTTeachersHandler)
	mux.HandleFunc("PUT /teachers", handlers.PUTTeachersHandler)
	mux.HandleFunc("PATCH /teachers", handlers.PATCHTeachersHandler)
	mux.HandleFunc("DELETE /teachers", handlers.DELETETeachersHandler)

	mux.HandleFunc("GET /teachers/{id}", handlers.GETTeacherByIDHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.PUTTeachersHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PATCHTeacherByIDHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DELETETeacherByIDHandler)

	mux.HandleFunc("GET /teachers/{id}/students", handlers.GETStudentsByTeacherIDHandler)

	return mux
}
