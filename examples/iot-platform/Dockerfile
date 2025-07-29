FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/examples/iot-platform
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o iot-platform .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN addgroup -g 1001 appgroup && \
    adduser -D -u 1001 -G appgroup appuser

WORKDIR /app

RUN mkdir -p /app/config

COPY --from=builder /app/examples/iot-platform/iot-platform .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 9000

CMD ["./iot-platform"]