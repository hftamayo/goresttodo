FROM golang:1.22.2-alpine AS builder

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

EXPOSE 8001

CMD ["./gotodo"]
#how to run this file:
#docker build --no-cache --platform linux/amd64 -t hftamayo/goresttodo:0.1.3-experimental .
# docker network create godev_network.
#docker run --name goresttodo --network developer_network -p 8001:8001 -d --env-file .env hftamayo/goresttodo:0.2.0-experimental
