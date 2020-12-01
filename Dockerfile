FROM golang:1.15-alpine3.12 as builder
WORKDIR /build
COPY . .
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -tags netgo -ldflags '-w' cmd/aws-ip-check/main.go

FROM alpine:3.12
WORKDIR /
COPY --from=builder /build/main ./aws-ip-check
ENTRYPOINT [ "/aws-ip-check" ]
