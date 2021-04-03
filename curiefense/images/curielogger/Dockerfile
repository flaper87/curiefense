FROM golang:1-alpine as builder

WORKDIR /app
COPY curielogger .
RUN go build -o build/curielogger ./cmd/grpc

FROM alpine
RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*
COPY --from=builder /app/build/curielogger /bin
ENTRYPOINT ["/bin/curielogger"]
