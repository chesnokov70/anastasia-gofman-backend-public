# Build stage
FROM golang:1.24-alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

RUN go install github.com/swaggo/swag/cmd/swag@latest

COPY . .


RUN swag init -g ./cmd/api/main.go -o cmd/docs

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/api

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

RUN mkdir -p /root/uploads/arts_photos /root/uploads/events_photos /root/uploads/authors_photos

COPY docker.env docker.env
COPY .env .env

EXPOSE 8010

CMD ["./main"] 