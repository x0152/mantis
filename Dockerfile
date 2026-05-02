FROM golang:1.25-alpine

ARG INFERENCED_VERSION=v0.2.11
ARG TARGETARCH

RUN apk add --no-cache git ca-certificates curl unzip libc6-compat

WORKDIR /app

RUN go install github.com/air-verse/air@latest \
 && go install github.com/pressly/goose/v3/cmd/goose@latest

RUN ARCH="${TARGETARCH:-$(uname -m | sed -e 's/x86_64/amd64/' -e 's/aarch64/arm64/')}" \
 && curl -fsSL -o /tmp/inferenced.zip \
      "https://github.com/gonka-ai/gonka/releases/download/release/${INFERENCED_VERSION}/inferenced-linux-${ARCH}.zip" \
 && unzip -p /tmp/inferenced.zip inferenced > /usr/local/bin/inferenced \
 && chmod +x /usr/local/bin/inferenced \
 && rm /tmp/inferenced.zip

ENV GONKA_INFERENCED_BIN=/usr/local/bin/inferenced

COPY go.mod go.sum ./
RUN go mod download

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]
