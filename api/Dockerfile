FROM golang:1.25-alpine AS builder

WORKDIR /app
# Copiamos el módulo y el código
COPY go.mod ./
COPY . .

# git es necesario internamente para que go mod tidy descargue ciertos paquetes
RUN apk add --no-cache git

# Descargamos e instalamos dependencias generando el go.sum automáticamente
RUN go mod tidy

# Binario estático puro
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dbp-api .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata openssh-client sshpass

# Crear usuario no-root para seguridad (CIS Benchmark)
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser/
COPY --from=builder /app/dbp-api .
COPY --from=builder /app/lang ./lang

# Ajustar permisos
RUN chown -R appuser:appgroup /home/appuser/

USER appuser

# Exponemos el puerto de Gin
EXPOSE 8089

CMD ["./dbp-api"]


