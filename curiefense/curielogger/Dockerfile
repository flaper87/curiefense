FROM golang:1-alpine as builder

WORKDIR /app
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN go build -o curielogger ./cmd/grpc

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
RUN mkdir -p /etc/curielogger/
COPY --from=builder /app/curielogger /bin
COPY --from=builder /app/curielogger.yml /etc/curielogger/

ENTRYPOINT ["/bin/curielogger"]