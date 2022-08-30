# Build the kubectl_fzf_server binary
FROM golang:latest as builder

ARG GIT_COMMIT
ARG GIT_BRANCH
ARG GO_VERSION
ARG VERSION
ARG BUILD_DATE

WORKDIR /workspace
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
    go build -a -o kubectl-fzf-server \
        -ldflags "-X 'main.gitCommit=$GIT_COMMIT' -X 'main.gitBranch=$GIT_BRANCH' -X 'main.goVersion=$GO_VERSION' -X 'main.buildDate=$BUILD_DATE' -X 'main.version=$VERSION'" \
        cmd/kubectl-fzf-server/main.go

# Copy the kubectl_fzf_server into a thin image
FROM alpine:latest
WORKDIR /
COPY --from=builder /workspace/kubectl-fzf-server .

ENTRYPOINT ["/kubectl-fzf-server"]
