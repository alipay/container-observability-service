# Build the aggregator binary
FROM golang:1.20.3 as builder

# Copy in the go src
WORKDIR /go/src/github.com/alipay/container-observability-service
COPY cmd/    cmd/
COPY internal internal
COPY vendor/ vendor/
COPY pkg/    pkg/
COPY statics/ statics/

# Build
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o aggregator github.com/alipay/container-observability-service/cmd/aggregator
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o auditinstaller github.com/alipay/container-observability-service/cmd/audit_init
RUN GO111MODULE=off CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -a -o grafanadi github.com/alipay/container-observability-service/cmd/grafanadi

# Copy the aggregator binary into a thin image
FROM ubuntu:devel
WORKDIR /
COPY --from=builder /go/src/github.com/alipay/container-observability-service/aggregator .
COPY --from=builder /go/src/github.com/alipay/container-observability-service/statics ./statics
COPY --from=builder /go/src/github.com/alipay/container-observability-service/auditinstaller .
COPY --from=builder /go/src/github.com/alipay/container-observability-service/grafanadi .
