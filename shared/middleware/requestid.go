package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// Definimos una clave tipo para evitar colisiones con otras librerías
type contextKey string

const (
    // RequestIDKey es la clave para guardar/recuperar el request ID del context
    RequestIDKey contextKey = "request_id"
    
    // HeaderXRequestID es el nombre estándar del header
    HeaderXRequestID = "X-Request-ID"
)

// RequestID es un middleware que asegura que toda petición tenga un request ID
func RequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // PASO 1: Obtener o generar el request ID
        requestID := r.Header.Get(HeaderXRequestID)
        
        // PASO 2: Si no existe, generar uno nuevo
        if requestID == "" {
            requestID = generateRequestID()
        }
        
        // PASO 3: Guardar en el context para que otros middlewares/handlers puedan acceder a él
        ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
        
        // PASO 4: Añadir al header de respuesta (para que el cliente lo vea)
        w.Header().Set(HeaderXRequestID, requestID)
        
        // PASO 5: Continuar con el siguiente middleware/handler con el nuevo context
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// generateRequestID crea un ID único simple
func generateRequestID() string {
    bytes := make([]byte, 16) // 16 bytes = 128 bits
    if _, err := rand.Read(bytes); err != nil {
        // Si falla la generación aleatoria, usamos un fallback
        // En realidad, esto casi nunca falla
        return "fallback-request-id"
    }
    return hex.EncodeToString(bytes) // Convertir a hexadecimal
}

// GetRequestID recupera el request ID del context
// Útil para otros middlewares y handlers
func GetRequestID(ctx context.Context) string {
    if id, ok := ctx.Value(RequestIDKey).(string); ok {
        return id
    }
    return "" // Si no hay ID, devolvemos vacío
}