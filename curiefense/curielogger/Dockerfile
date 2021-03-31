FROM golang:1-alpine as builder

WORKDIR /app
COPY go.mod go.mod
COPY go.sum go.sum
RUN go mod download
COPY . .
RUN go build -o curielogger ./cmd/grpc

FROM alpine
COPY --from=builder /app/curielogger /bin
ENTRYPOINT ["/bin/curielogger"]