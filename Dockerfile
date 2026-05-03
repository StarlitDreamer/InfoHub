FROM golang:1.23-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/infohub-agent ./cmd

FROM alpine:3.20

WORKDIR /app

RUN addgroup -S infohub && adduser -S infohub -G infohub

COPY --from=builder /out/infohub-agent /app/infohub-agent
COPY configs /app/configs

RUN mkdir -p /app/data && chown -R infohub:infohub /app

USER infohub

EXPOSE 8080

ENV INFOHUB_CONFIG_PATH=/app/configs/config.example.json
ENV INFOHUB_HTTP_ADDR=:8080

CMD ["/app/infohub-agent", "serve"]
