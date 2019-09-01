# Build the cache_builder binary
FROM golang:latest as builder

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o cache_builder cmd/cache_builder/main.go

# Copy the cache_builder into a thin image
FROM alpine:latest
WORKDIR /
COPY --from=builder /workspace/cache_builder .

RUN apk add --no-cache rsync

ENTRYPOINT ["/cache_builder"]
