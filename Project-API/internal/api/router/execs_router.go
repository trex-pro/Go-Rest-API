package router

import (
	"net/http"
	"project-api/internal/api/handlers"
)

func execsRouter() *http.ServeMux {
	mux := http.NewServeMux()

	// Execs.
	mux.HandleFunc("GET /execs", handlers.GETExecsHandler)
	mux.HandleFunc("POST /execs", handlers.POSTExecsHandler)
	mux.HandleFunc("PATCH /execs", handlers.PATCHExecsHandler)

	mux.HandleFunc("GET /execs/{id}", handlers.GETExecByIDHandler)
	mux.HandleFunc("PATCH /execs/{id}", handlers.PATCHExecByIDHandler)
	mux.HandleFunc("DELETE /execs/{id}", handlers.DELETEExecByIDHandler)
	// mux.HandleFunc("POST /execs/{id}/updatepassword", handlers.ExecsHandler)

	mux.HandleFunc("POST /execs/login", handlers.LoginHandler)
	mux.HandleFunc("POST /execs/logout", handlers.LogoutHandler)
	// mux.HandleFunc("POST /execs/forgotpassword", handlers.ExecsHandler)
	// mux.HandleFunc("POST /execs/resetpassword/{resetcode}", handlers.ExecsHandler)

	return mux
}
