# Kubectl-fzf

kubectl-fzf provides a fast and powerful fzf autocompletion for kubectl.

[![asciicast](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja.png)](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja?t=01)

Table of Contents
=================

* [Kubectl-fzf](#kubectl-fzf)
* [Table of Contents](#table-of-contents)
* [Features](#features)
* [Requirements](#requirements)
* [Installation](#installation)
   * [kubectl-fzf binaries](#kubectl-fzf-binaries)
   * [Shell autocompletion](#shell-autocompletion)
      * [Zsh plugins: Antigen](#zsh-plugins-antigen)
   * [Deploy kubectl-fzf-server as a pod](#deploy-kubectl-fzf-server-as-a-pod)
* [Usage](#usage)
   * [kubectl-fzf-server: local version](#kubectl-fzf-server-local-version)
      * [Configuration](#configuration)
   * [kubectl-fzf-server: pod version](#kubectl-fzf-server-pod-version)
   * [Completion](#completion)
      * [Configuration](#configuration-1)
* [Troubleshooting](#troubleshooting)
   * [Debug Completion](#debug-completion)
   * [Debug Server](#debug-server)

# Features

- Seamless integration with kubectl autocompletion
- Fast completion
- Label autocompletion
- Automatic namespace switch

# Requirements

- go (minimum version 1.19)
- [fzf](https://github.com/junegunn/fzf)

# Installation

## kubectl-fzf binaries

```shell
# Completion binary called during autocompletion
go install -u github.com/bonnefoa/kubectl-fzf/v3/cmd/kubectl-fzf-completion@main
# If you want to run the kubectl-fzf server locally
go install -u github.com/bonnefoa/kubectl-fzf/v3/cmd/kubectl-fzf-server@main
```

`kubectl-fzf-completion` needs to be in you $PATH so make sure that your $GOPATH bin is included:
```
PATH=$PATH:$GOPATH/bin
```

## Shell autocompletion

Source the autocompletion functions:
```
# bash version
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/master/shell/kubectl_fzf.bash -O ~/.kubectl_fzf.bash
echo "source <(kubectl completion bash)" >> ~/.bashrc
echo "source ~/.kubectl_fzf.bash" >> ~/.bashrc

# zsh version
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/master/shell/kubectl_fzf.plugin.zsh -O ~/.kubectl_fzf.plugin.zsh
echo "source <(kubectl completion zsh)" >> ~/.zshrc
echo "source ~/.kubectl_fzf.plugin.zsh >> ~/.zshrc
```

### Zsh plugins: Antigen

You can use [antigen](https://github.com/zsh-users/antigen) to load it as a zsh plugin
```shell
antigen bundle robbyrussell/oh-my-zsh plugins/docker
antigen bundle bonnefoa/kubectl-fzf@main shell/
```

## Deploy kubectl-fzf-server as a pod

You can deploy `kubectl-fzf-server` as a pod in your cluster.

From the [k8s directory](https://github.com/bonnefoa/kubectl-fzf/tree/main/k8s):
```shell
helm template --namespace myns --set image.kubectl_fzf_server.tag=v3 --set toleration=aToleration . | kubectl apply -f -
```

You can check the latest image version [here](https://cloud.docker.com/repository/docker/bonnefoa/kubectl-fzf/general).

# Usage

## kubectl-fzf-server: local version

`kubectl-fzf-server` will watch cluster resources and keep the current state of the cluster in local files.
By default, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

Advantages:
- Minimal setup needed.
- Local cache is maintained up to date.

Drawbacks:
- It can be CPU and memory intensive on big clusters.
- It also can be bandwidth intensive. The most expensive is the initial listing at startup and on error/disconnection. Big namespace can increase the probability of errors during initial listing.
- It can generate load on the kube-api servers if multiple user are running it.

To create cache files necessary for `kubectl_fzf`, just run in a tmux or a screen

```shell
kubectl-fzf-server
```

It will watch the cluster in the current context. If you switch context, `kubectl-fzf-server` will detect and start watching the new cluster.
The initial resource listing can be long on big clusters and autocompletion might need 30s+.

`connect: connection refused` or similar messages are expected if there's network issues/interruptions and `kubectl-fzf-server` will automatically reconnect.

### Configuration

You can configure `kubectl-fzf-server` with the configuration file `$HOME/.kubectl_fzf.yaml`

```yaml
# Role to hide from the role list in node completion
ignored-node-roles:
  - common-node
# Namespaces to exclude for configmap and pod listing
# Regexps are accepted
excluded-namespaces:
  - consul-agent
  - datadog-agent
  - coredns
  - kube-system
  - kube2iam
  - dev-.*
```

## kubectl-fzf-server: pod version

If the pod is deployed in your cluster, the autocompletion will be fetched automatically fetched using port forward.

Advantages:
- No need to run a local `kubectl-fzf-server`
- Only a single instance of `kubectl-fzf-server` per cluster is needed, lowering the load on the `kube-api` servers.

Drawbacks:
- Resources need to be fetched remotely, this can increased the completion time. A local cache is maintained to lower this.

## Completion

Once `kubectl-fzf-server` is running, you will be able to use `kubectl_fzf` by calling the kubectl completion
```shell
# Get fzf completion on pods on all namespaces
kubectl get pod <TAB>

# Open fzf autocompletion on all available label
kubectl get pod -l <TAB>

# Open fzf autocompletion on all available field-selector. Usually much faster to list all pods running on an host compared to kubectl describe node.
kubectl get pod --field-selector <TAB>

# This will fallback to the normal kubectl completion (if sourced) 
kubectl <TAB>
```

### Configuration

By default, the local port used for the port-forward is 8080. You can override it through an environment variable:
```
KUBECTL_FZF_PORT_FORWARD_LOCAL_PORT=8081
```

# Troubleshooting

## Debug kubectl-fzf-completion

To directly call completion with debug logs, run: 
```
KUBECTL_FZF_LOG_LEVEL=debug kubectl-fzf-completion k8s_completion get pods ""
```

## Debug Tab Completion

To debug Tab completion, you can activate debug logs:
```
export KUBECTL_FZF_COMP_DEBUG_FILE=/tmp/debug
```

Check that the completion function is correctly sourced:
```
type kubectl_fzf_completion
kubectl_fzf_completion is a shell function from /home/bonnefoa/.antigen/bundles/kubectl-fzf-main/shell/kubectl_fzf.plugin.zsh
```

Use zsh completion debug:
```
kubectl get pods <C-X>?
Trace output left in /tmp/zsh497886kubectl1 (up-history to view)
```

## Debug kubectl-fzf-server

To launch kubectl-fzf-server with debug logs
```shell
kubectl-fzf-server --log-level debug
```
