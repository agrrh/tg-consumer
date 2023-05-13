FROM golang:1.20 as builder

ENV CGO_ENABLED 0
ENV GOOS linux

WORKDIR /app

COPY ./go.mod ./
COPY ./go.sum ./
RUN go mod download

COPY ./ ./
RUN go build ./

FROM alpine:3.17

RUN apk update --no-cache && \
  apk add --no-cache \
    tzdata==2023c-r0 \
    ca-certificates==20230506-r0

WORKDIR /app

COPY --from=builder /app/tg-consumer ./

ENTRYPOINT ["/app/tg-consumer"]
CMD []
