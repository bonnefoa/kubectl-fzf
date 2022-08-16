all: build test

build:
	go build ./cmd/kubectl-fzf-server
	go build ./cmd/kubectl-fzf-completion

docker:
	eval $$(minikube docker-env) && docker build . -t bonnefoa/kubectl-fzf:v2.0.0

test:
	go test ./...

clean:
	go clean
