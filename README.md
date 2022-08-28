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
  * [Cache builder](#cache-builder)
	 * [Local installation](#local-installation)
	 * [As a kubernetes deployment](#as-a-kubernetes-deployment)
  * [Shell autocompletion](#shell-autocompletion)
	 * [Using zplug](#using-zplug)
* [Usage](#usage)
  * [kubectl-fzf-server](#kubectl-fzf-server)
	 * [Configuration](#configuration)
  * [kubectl_fzf](#kubectl_fzf)
	 * [Options](#options)
* [Caveats](#caveats)
* [Troubleshooting](#troubleshooting)
  * [Debug logs](#debug-logs)
  * [The normal autocompletion is used](#the-normal-autocompletion-is-used)

# Features

- Seamless integration with kubectl autocompletion
- Fast completion
- Label autocompletion
- Automatic namespace switch

# Requirements

- go (minimum version 1.18)
- [fzf](https://github.com/junegunn/fzf)

# Installation

## kubectl-fzf-server

### Local installation

Install `kubectl-fzf-server`:
```shell
# Mac
FILE="kubectl-fzf_darwin_amd64.tar.gz"
# Linux
FILE="kubectl-fzf_linux_amd64.tar.gz"

cd /tmp
wget "https://github.com/bonnefoa/kubectl-fzf/releases/latest/download/$FILE"
tar -xf $FILE
install kubectl-fzf-server ~/local/bin/kubectl-fzf-server
```

### As a kubernetes deployment

You can deploy `kubectl-fzf-server` as a pod in your cluster.

```shell

helm template --namespace myns --set image.kubectl_fzf_server.tag=1.0.11 --set toleration=aToleration . | kubectl apply -f -
```

You can check the latest image version [here](https://cloud.docker.com/repository/docker/bonnefoa/kubectl-fzf/general).

## Shell autocompletion

Source the autocompletion functions:
```shell
# kubectl_fzf.sh needs to be sourced after kubectl completion.

# bash version
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/master/shell/kubectl_fzf.sh -O ~/.kubectl_fzf.sh
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/master/shell/kubectl_fzf_completion.bash -O ~/.kubectl_fzf_completion.bash
echo "source <(kubectl completion bash)" >> ~/.bashrc
echo "source ~/.kubectl_fzf_completion.bash" >> ~/.bashrc

# zsh version
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/master/shell/kubectl_fzf.sh -O ~/.kubectl_fzf.sh
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/master/shell/kubectl_fzf_completion.zsh -O ~/.kubectl_fzf_completion.zsh
echo "source <(kubectl completion zsh)" >> ~/.zshrc
echo "source ~/.kubectl_fzf_completion.zsh" >> ~/.zshrc
```

### Using zplug

You can use zplug to install the autocompletion functions
```shell
zplug "plugins/kubectl", from:oh-my-zsh, defer:2
zplug "bonnefoa/kubectl-fzf", defer:3
```

# Usage

## kubectl-fzf-server: local version

`kubectl-fzf-server` will watch cluster resources and keep the current state of the cluster in local files.
By default, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

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
ignored-node-role:
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

### Advantages

- Minimal setup needed.

### Drawbacks

- It can be CPU and memory intensive on big clusters
- It also can be bandwidth intensive. The most expensive is the initial listing at startup and on error/disconnection. Big namespace can increase the probability of errors during initial listing.
- It can generate load on the kube-api servers if multiple user are running it

## kubectl-fzf-server: pod version

If the pod is deployed in your cluster, the autocompletion will be fetched with port forward.

### Advantages

- No need to run a local `kubectl-fzf-server`
- Only a single instance of `kubectl-fzf-server` per cluster is needed, making it more optimal on kube-api servers.

### Drawbacks

- Resources need to be fetched remotely, this can increased the completion time. A local cache is maintained to lower this.

## kubectl_fzf

Once `kubectl-fzf-server` is running, you will be able to use `kubectl_fzf` by calling the kubectl completion
```shell
kubectl get pod <TAB>
```

# Troubleshooting

## Debug Completion

To directly call completion with debug logs, run: 
```
KUBECTL_FZF_LOG_LEVEL=debug kubectl-fzf-completion k8s_completion get pods ""
```

## Debug Server

To launch kubectl-fzf-server with debug logs
```shell
kubectl-fzf-server --log-level debug
```
