package handlers

import "net/http"

func StudentHandler(w http.ResponseWriter, r *http.Request) {
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
