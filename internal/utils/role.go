package utils

import "net/http"

func RequireEmployee(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Role") != "employee" {
			EncodeJson(w, r, http.StatusForbidden, map[string]any{
				"error":   true,
				"code":    "FORBIDDEN",
				"message": "employee role required",
			})
			return
		}
		next.ServeHTTP(w, r)
	})
}
