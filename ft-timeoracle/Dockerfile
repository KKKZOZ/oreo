FROM golang:1.23-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

WORKDIR /app/ft-timeoracle
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ft-timeoracle .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/ft-timeoracle/ft-timeoracle .

RUN chmod +x ft-timeoracle

EXPOSE 8012 8013

CMD ["./ft-timeoracle"]