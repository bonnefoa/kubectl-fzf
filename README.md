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
  * [cache_builder](#cache_builder)
	 * [Configuration](#configuration)
  * [kubectl_fzf](#kubectl_fzf)
	 * [Options](#options)
* [Caveats](#caveats)
* [Troubleshooting](#troubleshooting)
  * [Debug logs](#debug-logs)
  * [The normal autocompletion is used](#the-normal-autocompletion-is-used)

# Features

- Seamless integration with kubectl autocompletion
- Sub second completion
- Label autocompletion
- Automatic namespace switch

# Requirements

- go (minimum version 1.12)
- [fzf](https://github.com/junegunn/fzf)
- coreutils `brew install coreutils`

# Installation

## Cache builder

### Local installation

Install `cache_builder`:
```shell
# Mac
FILE="kubectl-fzf_darwin_amd64.tar.gz"
# Linux
FILE="kubectl-fzf_linux_amd64.tar.gz"

cd /tmp
wget "https://github.com/bonnefoa/kubectl-fzf/releases/latest/download/$FILE"
tar -xf $FILE
install cache_builder ~/local/bin/cache_builder
```

### As a kubernetes deployment

You can deploy the cache builder as a pod in your cluster.

```shell

helm template --namespace myns --set image.cache_builder.tag=1.0.11 --set toleration=aToleration . | kubectl apply -f -
```

You can check the latest image version [here](https://cloud.docker.com/repository/docker/bonnefoa/kubectl-fzf/general).

## Shell autocompletion

Source the autocompletion functions:
```shell
# kubectl_fzf.sh needs to be sourced after kubectl completion.

# bash version
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/master/kubectl_fzf.sh -O ~/.kubectl_fzf.sh
echo "source <(kubectl completion bash)" >> ~/.bashrc
echo "source ~/.kubectl_fzf.sh" >> ~/.bashrc

# zsh version
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/master/kubectl_fzf.plugin.zsh -O ~/.kubectl_fzf.plugin.zsh
echo "source <(kubectl completion zsh)" >> ~/.zshrc
echo "source ~/.kubectl_fzf.plugin.zsh" >> ~/.zshrc
```

### Using zplug

You can use zplug to install the autocompletion functions
```shell
zplug "plugins/kubectl", from:oh-my-zsh, defer:2
zplug "bonnefoa/kubectl-fzf", defer:3
```

# Usage

## cache_builder: local version

`cache_builder` will watch cluster resources and keep the current state of the cluster in local files.
By default, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

To create cache files necessary for `kubectl_fzf`, just run in a tmux or a screen

```shell
cache_builder
```

It will watch the cluster in the current context. If you switch context, `cache_builder` will detect and start watching the new cluster.
The initial resource listing can be long on big clusters and autocompletion might need 30s+.

`connect: connection refused` or similar messages are expected if there's network issues/interruptions and `cache_builder` will automatically reconnect.

### Configuration

You can configure `cache_builder` with the configuration file `$HOME/.kubectl_fzf.yaml`

```yaml
# Role to hide from the role list in node completion
role-blacklist:
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

## cache_builder: pod version

If the pod is deployed in your cluster, the autocompletion fetch the cache files from the pod using rsync.

To check if you can reach the pod, try:

```shell
IP=$(k get endpoints -l app=kubectl-fzf --all-namespaces -o=jsonpath='{.items[*].subsets[*].addresses[*].ip}')
nc -v -z $IP 80
```

By default, it will use the port 80.
You can change it by deploying the chart with a different port value and using `KUBECTL_FZF_RSYNC_PORT`:

```shell
helm template --namespace myns --set port=873 . | kubectl apply -f -
export KUBECTL_FZF_RSYNC_PORT=873
```

If there's no direct access, a `port-forward` will be opened.
Opening a `port-forward` can take several seconds and thus, slow the autocompletion.
If you want to avoid this, you can keep the `open_port_forward.sh` script running which will keep the `port-forward` opened.

### Advantages

- No need to run a local `cache_builder`
- Bandwidth usage is limited to the rsync transfert which is low.

### Drawbacks

- Rsynced resources are cached for 30 seconds so the autocompletion can get outdated.

## kubectl_fzf

Once `cache_builder` is running, you will be able to use `kubectl_fzf` by calling the kubectl completion
```shell
kubectl get pod <TAB>
```

### Options

| Environment variable        | Description                                            | Default                                     |
| --------------------        | --------------------                                   | --------------------                        |
| KUBECTL_FZF_CACHE           | Cache files location                                   | `/tmp/kubectl_fzf_cache`                    |
| KUBECTL_FZF_EXCLUDE         | Exclusion patterns passed to the autocompletion        | ""                                          |
| KUBECTL_FZF_OPTIONS         | fzf parameters                                         | `-1 --header-lines=2 --layout reverse -e --no-hscroll --no-sort `   |

To turn down exact match in search:
```shell
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse)
```

To enable hscroll
```shell
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse -e)
```

To bind space to accept current result (You can check the list of available keys and actions in the fzf man)
```shell
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse -e --no-hscroll --no-sort --bind space:accept)
```


To exclude all namespaces starting with "dev" and consul-agent resources:
```shell
export KUBECTL_FZF_EXCLUDE=("^dev" "consul-agent")
```

# Caveats

With zsh, if the suggested completion doesn't match the start of the query, the completion will fail.

```shell
k get pod pr<TAB>
# result needs to start with `pr` otherwise autocompletion will fail
```

---

If you're using an out-of-the-box `oh-my-zsh` configuration, the default `matcher-list` zstyle (`zstyle ':completion:*' matcher-list 'r:|=*' 'l:|=* r:|=*'`) will interfere with the search. If fzf does not find any match, or if you interrupt it by pressing `Esc` or `Ctrl-C/Cmd-C`, zsh will see it as a failed completion and will restart it again.

Changing the zstyle to `zstyle ':completion:*' matcher-list 'r:|=*'` fixes the issue.

# Troubleshooting

## Debug logs

To launch cache_builder with debug logs
```shell
cache_builder -logtostderr -v 14
```

## The normal autocompletion is used

First, check if cache files are correctly generated in `/tmp/kubectl_fzf_cache`.
The autocompletion will fallback to normal method if cache files are absent.

If the files are present, check that the `__kubectl_get_containers` is correctly overloaded

```
# Incorrect type
type __kubectl_get_containers
__kubectl_get_containers is a shell function from /dev/fd/15

# Expected output
type __kubectl_get_containers
__kubectl_get_containers is a shell function from .../kubectl-fzf/kubectl_fzf.plugin.zsh
```

Be sure that `kubectl_fzf.plugin` is loaded after `kubectl completion zsh` in your bashrc/zshrc.
