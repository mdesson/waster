FROM golang:latest AS builder

WORKDIR /usr/src/waster
COPY . .
RUN go mod tidy
RUN go build -o app .

FROM alpine:latest

WORKDIR /waster
COPY --from=builder /usr/src/waster/app /waster/

CMD ["./app"]