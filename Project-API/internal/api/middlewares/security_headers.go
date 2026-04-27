package middlewares

import "net/http"

func SecurityHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Use OWASP HTTP Header Cheat Sheet for Recommendations.
		w.Header().Set("X-DNS-Prefetch-Control", "off")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "0")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		w.Header().Set("Strict-Transport-Policy", "max-age=63072000;includeSubDomains;preload")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Cross-Origin-Resource-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Opener-Policy", "same-origin")
		w.Header().Set("Cross-Origin-Embedder-Policy", "require-corp")
		w.Header().Set("Permission-Policy", "geolocation=(self), microphone=()")

		next.ServeHTTP(w, r)
	})
}
