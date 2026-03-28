package router

import (
	"net/http"
	"project-api/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)

	mux.HandleFunc("/students/", handlers.StudentHandler)

	mux.HandleFunc("GET /teachers/", handlers.GETTeachersHandler)
	mux.HandleFunc("POST /teachers/", handlers.POSTTeachersHandler)
	mux.HandleFunc("PUT /teachers/", handlers.PUTTeachersHandler)
	mux.HandleFunc("PATCH /teachers/", handlers.PATCHTeachersHandler)
	mux.HandleFunc("DELETE /teachers/", handlers.DELETETeacherHandler)

	mux.HandleFunc("GET /teachers/{id}", handlers.GETTeacherByIDHandler)
	mux.HandleFunc("PUT /teachers/{id}", handlers.PUTTeachersHandler)
	mux.HandleFunc("PATCH /teachers/{id}", handlers.PATCHTeacherByIDHandler)
	mux.HandleFunc("DELETE /teachers/{id}", handlers.DELETETeacherByIDHandler)

	mux.HandleFunc("/execs/", handlers.ExecHandler)

	return mux
}
