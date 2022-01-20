FROM ghcr.io/packagrio/packagr:latest-golang AS builder
WORKDIR /go/src/github.com/analogj/hatchet
COPY . .

RUN go mod vendor && \
    go build -ldflags="-extldflags=-static" -o hatchet-linux-amd64 -tags "static,netgo" $(go list ./cmd/...)

FROM scratch
# copy the ca-certificate.crt from the build stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/src/github.com/analogj/hatchet/hatchet-linux-amd64 /hatchet
ENTRYPOINT ["/hatchet"]