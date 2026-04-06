FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .

# git es necesario en Alpine para que go mod tidy resuelva dependencias externas
RUN apk add --no-cache git

# Automáticamente reconcilia las dependencias y crea el go.sum que faltaba
RUN go mod tidy

# Compilamos el binario estático optimizado
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dbp-agent .

FROM alpine:latest
# Añadimos dependencias útiles: certificados para la API s3, restic para el backup real, y utilidades de base de datos para dumps locales
RUN apk --no-cache add ca-certificates restic mariadb-client postgresql-client tzdata

# Crear usuario no-root 
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /home/appuser/
COPY --from=builder /app/dbp-agent .

# Ajustar permisos
RUN chown -R appuser:appgroup /home/appuser/

USER appuser

# Ejecutamos el agente
CMD ["./dbp-agent"]

