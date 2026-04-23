FROM golang:1.25-alpine

RUN apk add --no-cache git ca-certificates

WORKDIR /app

RUN go install github.com/air-verse/air@latest \
 && go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
