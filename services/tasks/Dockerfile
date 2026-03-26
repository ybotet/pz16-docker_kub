FROM golang:1.25-alpine AS builder
WORKDIR /app

# Copiar todo el código fuente
COPY . .

# Descargar dependencias
RUN go mod download

# Compilar tasks service
RUN go build -o tasks-service ./services/tasks/cmd/tasks

FROM alpine:latest
RUN apk --no-cache add wget
WORKDIR /root/
COPY --from=builder /app/tasks-service .
EXPOSE 8082
CMD ["./tasks-service"]
