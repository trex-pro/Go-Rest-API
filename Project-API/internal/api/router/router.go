package router

import (
	"net/http"
	"project-api/internal/api/handlers"
)

func Router() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler)
	mux.HandleFunc("/students/", handlers.StudentHandler)
	mux.HandleFunc("/teachers/", handlers.TeacherHandler)
	mux.HandleFunc("/execs/", handlers.ExecHandler)

	return mux
}
