all: build test

build:
	go build ./cmd/kubectl-fzf-server
	go build ./cmd/kubectl-fzf-completion

install:
	go install ./cmd/kubectl-fzf-server
	go install ./cmd/kubectl-fzf-completion

docker:
	eval $$(minikube docker-env) && docker build . -t bonnefoa/kubectl-fzf:v2.0.0

test:
	go test ./...

graph:
	goda graph ./... | dot -Tsvg -o graph.svg

clean:
	go clean
