# Dockerfile
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copiar todo el código fuente primero
COPY . .

# Descargar dependencias (necesita ver los replace locales)
RUN go mod download

# Compilar para Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o tasks ./services/task/cmd/task

# Imagen final
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copiar el binario
COPY --from=builder /app/tasks .

EXPOSE 8082

CMD ["./tasks"]