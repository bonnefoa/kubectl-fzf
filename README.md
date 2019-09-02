# Kubectl-fzf

kubectl-fzf provides a fast and powerful fzf autocompletion for kubectl.

[![asciicast](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja.png)](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja?t=01)

Table of Contents
=================

- [Kubectl-fzf](#kubectl-fzf)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
  - [Using zplug](#using-zplug)
- [Usage](#usage)
  - [cache_builder](#cachebuilder)
    - [Configuration](#configuration)
  - [kubectl_fzf](#kubectlfzf)
    - [Options](#options)
- [Caveats](#caveats)
- [Troubleshooting](#troubleshooting)
  - [Debug logs](#debug-logs)
  - [The normal autocompletion is used](#the-normal-autocompletion-is-used)

# Features

- Seamless integration with kubectl autocompletion
- Sub second completion
- Label autocompletion
- Automatic namespace switch

# Requirements

- go (minimum version 1.12)
- [fzf](https://github.com/junegunn/fzf)

# Installation

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

Source the autocompletion functions:
```shell
# kubectl_fzf.sh needs to be sourced after kubectl completion.

# bash version
echo "source <(kubectl completion bash)" >> ~/.bashrc
echo "source $GOPATH/src/github.com/bonnefoa/kubectl-fzf/kubectl_fzf.sh" >> ~/.bashrc

# zsh version
echo "source <(kubectl completion zsh)" >> ~/.zshrc
echo "source $GOPATH/src/github.com/bonnefoa/kubectl-fzf/kubectl_fzf.plugin.zsh" >> ~/.zshrc
```

## Using zplug

You can use zplug to install the autocompletion functions
```shell
zplug "plugins/kubectl", from:oh-my-zsh, defer:2
zplug "bonnefoa/kubectl-fzf", defer:3
```

# Usage

## cache_builder

`cache_builder` will watch cluster resources and keep the current state of the cluster in local files.
By default, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

To create cache files necessary for `kubectl_fzf`, just run in a tmux or a screen

```shell
cache_builder
# If $GOPATH/bin is not in you $PATH
$GOPATH/bin/cache_builder
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
| KUBECTL_FZF_OPTIONS         | fzf parameters                                         | `-1 --header-lines=2 --layout reverse -e`   |

To turn down exact match in search:
```shell
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse)
```

To enable hscroll
```shell
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse -e)
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
