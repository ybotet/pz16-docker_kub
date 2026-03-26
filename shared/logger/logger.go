package logger

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Config contiene la configuración del logger
type Config struct {
    ServiceName string
    Environment string // development, production
    LogLevel    string
    JSONFormat  bool
}

// New crea una instancia configurada de logrus.Logger
func New(cfg Config) *logrus.Logger {
    log := logrus.New()
    
    // Configurar nivel de log
    level, err := logrus.ParseLevel(cfg.LogLevel)
    if err != nil {
        level = logrus.InfoLevel // Valor por defecto
    }
    log.SetLevel(level)
    
    // Configurar formato
    if cfg.JSONFormat {
        log.SetFormatter(&logrus.JSONFormatter{
            TimestampFormat: time.RFC3339Nano,
            FieldMap: logrus.FieldMap{
                logrus.FieldKeyTime:  "ts",
                logrus.FieldKeyLevel: "level",
                logrus.FieldKeyMsg:   "message",
                logrus.FieldKeyFunc:  "caller",
            },
        })
    } else {
        // Formato texto bonito para desarrollo
        log.SetFormatter(&logrus.TextFormatter{
            FullTimestamp:   true,
            TimestampFormat: time.RFC3339,
        })
    }
    
    // Configurar salida (por defecto stdout)
    log.SetOutput(os.Stdout)
    
    return log
}

// Campos comunes que usaremos en todos los servicios
const (
    FieldService   = "service"
    FieldRequestID = "request_id"
    FieldMethod    = "method"
    FieldPath      = "path"
    FieldStatus    = "status"
    FieldDuration  = "duration_ms"
    FieldError     = "error"
    FieldComponent = "component"
)