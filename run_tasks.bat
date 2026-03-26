@echo off
cd /d "%~dp0services\tasks"
set TASKS_PORT=8082
set AUTH_GRPC_ADDR=localhost:50051
echo Iniciando Tasks Service en puerto %TASKS_PORT%...
echo.
go run ./cmd/tasks
pause