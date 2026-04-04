FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod ./
# Descomentar al generar el go.sum:
# COPY go.sum ./
# RUN go mod download

COPY . .
# Binario estático puro
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dbp-api .

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/
COPY --from=builder /app/dbp-api .

# Exponemos el puerto de Gin
EXPOSE 8080

CMD ["./dbp-api"]
