package middleware

import (
	"log" // AÑADIR temporalmente
	"net/http"
)

func CSRFMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // === DEBUG TEMPORAL ===
        log.Println("========== CSRF DEBUG ==========")
        log.Printf("Method: %s", r.Method)
        log.Printf("URL: %s", r.URL.Path)
        log.Printf("Cookies recibidas:")
        for _, cookie := range r.Cookies() {
            log.Printf("  - %s = %s", cookie.Name, cookie.Value)
        }
        log.Printf("Headers:")
        log.Printf("  X-CSRF-Token: %s", r.Header.Get("X-CSRF-Token"))
        log.Println("=================================")
        // === FIN DEBUG ===

        // Solo proteger métodos que modifican estado
        if r.Method == http.MethodGet || 
           r.Method == http.MethodHead || 
           r.Method == http.MethodOptions {
            next(w, r)
            return
        }

        // 1. Obtener token de la cookie
        csrfCookie, err := r.Cookie("csrf_token")
        if err != nil {
            log.Printf("ERROR: Cookie csrf_token no encontrada: %v", err) // DEBUG
            http.Error(w, `{"error":"CSRF token cookie no encontrada"}`, http.StatusForbidden)
            return
        }

        // 2. Obtener token del header
        csrfHeader := r.Header.Get("X-CSRF-Token")
        if csrfHeader == "" {
            log.Printf("ERROR: Header X-CSRF-Token vacío") // DEBUG
            http.Error(w, `{"error":"Header X-CSRF-Token requerido"}`, http.StatusForbidden)
            return
        }

        // 3. Comparar tokens
        if csrfCookie.Value != csrfHeader {
            log.Printf("ERROR: Tokens no coinciden: cookie=%s, header=%s", 
                csrfCookie.Value, csrfHeader) // DEBUG
            http.Error(w, `{"error":"CSRF token inválido"}`, http.StatusForbidden)
            return
        }

        log.Printf("✅ CSRF validación exitosa") // DEBUG
        next(w, r)
    }
}