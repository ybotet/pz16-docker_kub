FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copiar todo el código fuente
COPY . .

# Descargar dependencias
RUN go mod download

# Compilar auth service
RUN go build -o auth-service ./services/auth/cmd/auth

FROM alpine:latest
# Instalar netcat para healthchecks
RUN apk --no-cache add wget netcat-openbsd

# Health probe
RUN wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.37/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

WORKDIR /root/
COPY --from=builder /app/auth-service .
EXPOSE 50051
CMD ["./auth-service"]
