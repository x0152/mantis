FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
COPY . .
RUN CGO_ENABLED=0 go build -o /mantis ./cmd

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=builder /mantis /mantis
COPY --from=builder /go/bin/goose /goose
COPY migrations /migrations
EXPOSE 8080
ENTRYPOINT ["/mantis"]
