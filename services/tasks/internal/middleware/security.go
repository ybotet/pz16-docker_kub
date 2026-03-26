package middleware

import "net/http"

// SecurityHeadersMiddleware añade headers de seguridad
func SecurityHeadersMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Prevenir MIME sniffing
        w.Header().Set("X-Content-Type-Options", "nosniff")
        
        // Prevenir clickjacking
        w.Header().Set("X-Frame-Options", "DENY")
        
        // Content Security Policy básica
        w.Header().Set("Content-Security-Policy", 
            "default-src 'self'; script-src 'self'")
        
        // Prevenir XSS en navegadores antiguos
        w.Header().Set("X-XSS-Protection", "1; mode=block")
        
        next.ServeHTTP(w, r)
    })
}