FROM golang:1.23-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux

RUN go build -ldflags="-s -w" -o /build/deadlockd .

FROM alpine:3.19

RUN addgroup -S app && adduser -S app -G app

COPY --from=builder /build/deadlockd /usr/local/bin/deadlockd

USER app

EXPOSE 8080

ENTRYPOINT ["deadlockd"]
