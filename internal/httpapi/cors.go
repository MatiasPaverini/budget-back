package httpapi

import "net/http"

// allowedOrigins covers the Angular dev server and the two schemes
// Capacitor's WebView runs under on iOS/Android. Update this once a real
// deployed frontend origin exists.
var allowedOrigins = map[string]bool{
	"http://localhost:4200": true,
	"capacitor://localhost": true,
	"ionic://localhost":     true,
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
			w.Header().Set("Vary", "Origin")
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
