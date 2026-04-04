FROM golang:1.21-alpine AS builder

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
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/dbp-api .

# Exponemos el puerto de Gin
EXPOSE 8080

CMD ["./dbp-api"]
