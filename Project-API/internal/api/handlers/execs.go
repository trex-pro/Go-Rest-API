package handlers

import "net/http"

func ExecHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.Write([]byte("This is EXECS GET method route."))
	case http.MethodPost:
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
