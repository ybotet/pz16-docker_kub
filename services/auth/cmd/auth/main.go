package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/ybotet/pz6_csrf_xss/services/auth/internal/auth"
	grpcserver "github.com/ybotet/pz6_csrf_xss/services/auth/internal/grpc"
	httpserver "github.com/ybotet/pz6_csrf_xss/services/auth/internal/http"
	"github.com/ybotet/pz6_csrf_xss/shared/logger"
)

func main() {
    // 1. Inicializar logger (el de tu proyecto)
    log := logger.New(logger.Config{
        ServiceName: "auth",
        Environment: "development",
        LogLevel:    "debug",
        JSONFormat:  true,
    })

    // 2. Configuración desde variables de entorno
    grpcPort := getEnv("AUTH_GRPC_PORT", "50051")
    httpPort := getEnv("AUTH_HTTP_PORT", "8081")
    jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")

    if jwtSecret == "your-secret-key-change-in-production" {
        log.Warn("JWT_SECRET no configurado, usando clave por defecto (¡NO USAR EN PRODUCCIÓN!)")
    }

    // 3. Crear servicio auth (compartido entre HTTP y gRPC)
    authService := auth.NewService(jwtSecret)

    // 4. Crear canales para errores de servidores
    errChan := make(chan error, 2)

    // 5. Iniciar servidor HTTP
    httpSrv := httpserver.NewServer(authService, log, httpPort)
    go func() {
        log.WithField("port", httpPort).Info("📡 Iniciando servidor HTTP")
        if err := httpSrv.Start(); err != nil && err != http.ErrServerClosed {
            errChan <- err
        }
    }()

    // 6. Iniciar servidor gRPC
    lis, err := net.Listen("tcp", ":"+grpcPort)
    if err != nil {
        log.WithError(err).Fatalf("Error escuchando puerto %s", grpcPort)
    }

    grpcSrv := grpc.NewServer()
    grpcserver.Register(grpcSrv, authService)

    go func() {
        log.WithField("port", grpcPort).Info("Iniciando servidor gRPC")
        if err := grpcSrv.Serve(lis); err != nil {
            errChan <- err
        }
    }()

    // 7. Generar token de prueba
    testToken, _ := authService.GenerateToken("test_user", 24*time.Hour)
    log.WithField("token", testToken).Info(" Token de prueba generado (válido 24h)")
    log.WithFields(map[string]interface{}{
        "http_port": httpPort,
        "grpc_port": grpcPort,
    }).Info("Servicios auth iniciados correctamente")

    // 8. Graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

    select {
    case <-sigChan:
        log.Info("Señal de terminación recibida")
    case err := <-errChan:
        log.WithError(err).Error("Error en servidor")
    }

    // 9. Apagar servidores gracefulmente
    log.Info("Apagando servidores...")

    // Apagar HTTP
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    if err := httpSrv.Shutdown(ctx); err != nil {
        log.WithError(err).Error("Error apagando HTTP")
    }

    // Apagar gRPC
    grpcSrv.GracefulStop()
    
    log.Info("Servidores apagados correctamente")
}

// Helper para variables de entorno
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}