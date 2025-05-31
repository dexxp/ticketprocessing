FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN if [ -f cmd/server/main.go ]; then go build -o server ./cmd/server; else go build -o server ./main.go; fi

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/server ./server
COPY --from=builder /app/config.yaml ./config.yaml
EXPOSE 8080
CMD ["./server"] 