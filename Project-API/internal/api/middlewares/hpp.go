package middlewares

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
)

type HPPOptions struct {
	CheckQuery               bool
	CheckBody                bool
	CheckBodyforConetentType string
	WhiteList                []string
}

func HPP(options HPPOptions) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if options.CheckBody && r.Method == http.MethodPost && isCorrectContentType(r, options.CheckBodyforConetentType) {
				// Filter Body Parameters.
				filterBodyParams(r, options.WhiteList)
				// Filter Qury Parameters.
				if options.CheckQuery && r.URL.Query() != nil {
					filterOueryParams(r, options.WhiteList)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isCorrectContentType(r *http.Request, contentType string) bool {
	return strings.Contains(r.Header.Get("Content-Type"), contentType)
}

func filterBodyParams(r *http.Request, whiteList []string) {
	err := r.ParseForm()
	if err != nil {
		fmt.Println(err)
		return
	}

	for key, value := range r.Form {
		if len(value) > 1 {
			r.Form.Set(key, value[0]) // For first value.
			// r.Form.Set(key, value[len(value)-1]) // For last value
		}
		if !isWhiteListed(key, whiteList) {
			delete(r.Form, key)
		}
	}
}

func filterOueryParams(r *http.Request, whiteList []string) {
	query := r.URL.Query()

	for key, value := range query {
		if len(value) > 1 {
			query.Set(key, value[0]) // For first value.
			// query.Set(key, value[len(value)-1]) // For last value
		}
		if !isWhiteListed(key, whiteList) {
			query.Del(key)
		}
	}
	r.URL.RawQuery = query.Encode()
}

func isWhiteListed(param string, whiteList []string) bool {
	return slices.Contains(whiteList, param)
}
