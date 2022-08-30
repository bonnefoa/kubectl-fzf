all: build test

LD_FLAGS=-ldflags "-X 'main.gitCommit=$(shell git rev-parse --short HEAD)' -X 'main.gitBranch=$(shell git rev-parse --abbrev-ref HEAD)' -X 'main.goVersion=$(shell go version)' -X 'main.buildDate=$(shell date -Iseconds -u)' -X 'main.version=$(shell git describe --tags)'"

build:
	go build $(LD_FLAGS) ./cmd/kubectl-fzf-server
	go build $(LD_FLAGS) ./cmd/kubectl-fzf-completion

install:
	go install $(LD_FLAGS) ./cmd/kubectl-fzf-server
	go install $(LD_FLAGS) ./cmd/kubectl-fzf-completion

DOCKER_BUILD_ARGS=--build-arg GIT_COMMIT="$(shell git rev-parse --short HEAD)" \
				  --build-arg GIT_BRANCH="$(shell git rev-parse --abbrev-ref HEAD)" \
				  --build-arg VERSION="$(shell git describe --tags)" \
				  --build-arg BUILD_DATE="$(shell date -Iseconds -u)" \
				  --build-arg GO_VERSION="$(shell go version)"

DOCKER_TAGS=-t bonnefoa/kubectl-fzf:latest \
	-t bonnefoa/kubectl-fzf:$(shell git describe --tags)

docker:
	docker build . \
		$(DOCKER_TAGS) \
		$(DOCKER_BUILD_ARGS)

docker-minikube:
	eval $$(minikube docker-env) && docker build . \
		$(DOCKER_TAGS) \
		$(DOCKER_BUILD_ARGS)

test:
	go test ./...

snapshot:
	GO_VERSION="$(shell go version)" goreleaser release --snapshot --rm-dist

release:
	GO_VERSION="$(shell go version)" goreleaser release --rm-dist

graph:
	goda graph ./... | dot -Tsvg -o graph.svg

clean:
	go clean
