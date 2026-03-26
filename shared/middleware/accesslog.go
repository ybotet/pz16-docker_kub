package middleware

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ybotet/pz6_csrf_xss/shared/logger"
)

// responseWriter es un wrapper que captura el código de estado
type responseWriter struct {
    http.ResponseWriter
    statusCode int
    written    int64
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    n, err := rw.ResponseWriter.Write(b)
    rw.written += int64(n)
    return n, err
}

// Logging middleware registra información de cada petición
func Logging(log *logrus.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 1. Obtener request_id del context
            requestID := GetRequestID(r.Context())
            
            // 2. Preparar logger con campos comunes
            entry := log.WithFields(logrus.Fields{
                logger.FieldRequestID: requestID,
                logger.FieldMethod:    r.Method,
                logger.FieldPath:      r.URL.Path,
                "remote_ip":           r.RemoteAddr,
                "user_agent":          r.UserAgent(),
            })
            
            // 3. Log de inicio (opcional - nivel DEBUG)
            entry.Debug("request started")
            
            // 4. Preparar para capturar el código de estado
            wrapper := &responseWriter{
                ResponseWriter: w,
                statusCode:     http.StatusOK, // Por defecto 200
            }
            
            // 5. Medir tiempo
            start := time.Now()
            
            // 6. Ejecutar el siguiente handler
            next.ServeHTTP(wrapper, r)
            
            // 7. Calcular duración
            duration := time.Since(start)
            durationMs := duration.Milliseconds()
            
            // 8. Log de finalización con TODOS los campos obligatorios
            entry = entry.WithFields(logrus.Fields{
                logger.FieldStatus:   wrapper.statusCode,
                logger.FieldDuration: durationMs,
            })
            
            // 9. Nivel de log según el código de estado
            if wrapper.statusCode >= 500 {
                entry.Error("request completed")
            } else if wrapper.statusCode >= 400 {
                entry.Warn("request completed")
            } else {
                entry.Info("request completed")
            }
        })
    }
}