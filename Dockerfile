FROM golang:1.21-alpine AS builder

WORKDIR /app

RUN apk add --no-cache bash gcc musl-dev

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -o gotodo cmd/todo/main.go

FROM alpine:3.18

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /app/gotodo .
COPY .env .

EXPOSE 8001

CMD ["./gotodo"]