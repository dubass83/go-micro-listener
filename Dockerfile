FROM golang:1.23.3-alpine3.20 as builder

WORKDIR /app
COPY . /app/
RUN go build -o main main.go

FROM alpine:3.20

WORKDIR /app
COPY --from=builder /app/main /app/
COPY --from=builder /app/conf.env /app/conf.env

CMD [ "/app/main" ]
