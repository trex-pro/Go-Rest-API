package utils

import "net/http"

type MiddleWare func(http.Handler) http.Handler

func applyMiddlewares(handler http.Handler, middlewares ...MiddleWare) http.Handler {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}
	return handler
}
