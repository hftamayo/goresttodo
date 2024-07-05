FROM golang:1.22.2-alpine

LABEL maintainer="Herbert Tamayo <hftamayo@gmail.com>"
WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8001

CMD ["./main"]