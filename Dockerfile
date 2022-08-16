# Build the kubectl_fzf_server binary
FROM golang:latest as builder

WORKDIR /workspace
COPY .git .
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY cmd/ cmd/
COPY pkg/ pkg/

# Build
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GO111MODULE=on \
    GIT_COMMIT=$(git rev-parse --short HEAD) \
    GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD) \
    VERSION=$(git describe --tags) \
    GO_VERSION=$(go version) \
    BUILD_DATE=$(date) \
    go build -a -o kubectl-fzf-server -ldflags "-X main.GitCommit=$GIT_COMMIT -X main.GitBranch=$GitBranch -X main.GoVersion=$GoVersion -X main.BuildDate=$BUILD_DATE -X main.Version=$VERSION" cmd/kubectl-fzf-server/main.go

# Copy the kubectl_fzf_server into a thin image
FROM alpine:latest
WORKDIR /
COPY --from=builder /workspace/kubectl-fzf-server .

ENTRYPOINT ["/kubectl-fzf-server"]
