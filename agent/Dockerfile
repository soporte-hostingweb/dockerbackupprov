FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
# Recomendable descomentar la siguiente línea cuando se genere el go.sum localmente
# COPY go.sum ./
# RUN go mod download

COPY . .
# Compilamos el binario estático optimizado
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dbp-agent .

FROM alpine:latest
# Añadimos dependencias útiles: certificados para la API s3, restic para el backup real, y utilidades de base de datos para dumps locales
RUN apk --no-cache add ca-certificates restic mariadb-client postgresql-client tzdata

WORKDIR /root/
COPY --from=builder /app/dbp-agent .

# Ejecutamos el agente
CMD ["./dbp-agent"]
