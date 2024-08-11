FROM golang:1.22-bookworm AS builder
LABEL maintainer="Nico Wehmöller" \
        org.label-schema.schema-version="1.0" \
        org.label-schema.description="DynDNS Server to work with Google Cloud DNS" \
        org.label-schema.name="dyndns"
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o dyndns.bin cmd/server/main.go

FROM alpine:latest AS tls
RUN  apk --no-cache add ca-certificates

FROM scratch
COPY --from=tls /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /build/dyndns.bin /dyndns.bin

EXPOSE 8080

ENTRYPOINT ["/dyndns.bin"]