@echo off
cd /d "%~dp0services\auth"
set AUTH_GRPC_PORT=50051
set JWT_SECRET=my-super-secret-key-for-auth-service-2024
echo Iniciando Auth Service en puerto %AUTH_GRPC_PORT%...
echo.
go run ./cmd/auth
pause