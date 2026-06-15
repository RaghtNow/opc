FROM golang:1.24.4-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/edu-insight-api ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /bin/edu-insight-api /app/edu-insight-api

EXPOSE 8088

CMD ["/app/edu-insight-api"]
