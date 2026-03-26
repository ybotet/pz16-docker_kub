package logger

import (
	"github.com/sirupsen/logrus"
)

// LogError registra un error con contexto pero DEVUELVE un mensaje seguro
func LogError(log *logrus.Logger, component string, err error, fields logrus.Fields) {
    if err == nil {
        return
    }
    
    // Asegurar campos mínimos
    if fields == nil {
        fields = logrus.Fields{}
    }
    
    fields[FieldComponent] = component
    fields[FieldError] = err.Error() // El error REAL va al log
    
    // Loggear con nivel ERROR
    log.WithFields(fields).Error("operation failed")
}

// SafeError es un error que TIENE un mensaje seguro para el cliente
type SafeError struct {
    Err        error  // Error real (solo logs)
    Message    string // Mensaje seguro para cliente
    StatusCode int    // Código HTTP recomendado
}

func (e SafeError) Error() string {
    return e.Message // Esto es lo que ve el cliente si se imprime
}

// NewSafeError crea un error seguro
func NewSafeError(err error, message string, statusCode int) SafeError {
    return SafeError{
        Err:        err,
        Message:    message,
        StatusCode: statusCode,
    }
}