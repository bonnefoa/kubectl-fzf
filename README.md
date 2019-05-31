# Kubectl-fzf

kubectl-fzf provides a fast and powerful fzf autocompletion for kubectl.

[![asciicast](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja.png)](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja?t=01)

Table of Contents
=================

   * [Features](#features)
   * [Requirements](#requirements)
   * [Installation](#installation)
      * [Using zplug](#using-zplug)
   * [Usage](#usage)
      * [cache_builder](#cache_builder)
         * [Watch a specific namespace](#watch-a-specific-namespace)
      * [kubectl_fzf](#kubectl_fzf)
         * [fzf options](#fzf-options)
   * [Caveats](#caveats)

# Features

- Seamless integration with kubectl autocompletion
- Sub second completion
- Label autocompletion
- Automatic namespace switch

# Requirements

[fzf](https://github.com/junegunn/fzf)

# Installation

Install `cache_builder`:
```shell
go get -u github.com/bonnefoa/kubectl-fzf/cmd/cache_builder
```

Source the autocompletion functions:
```shell
# kubectl_fzf.sh needs to be sourced after kubectl completion.

# bash version
echo "source <(kubectl completion bash)" >> ~/.bashrc
echo "source $GOPATH/src/github.com/bonnefoa/kubectl-fzf/kubectl_fzf.sh" >> ~/.bashrc

# zsh version
echo "source <(kubectl completion zsh)" >> ~/.zshrc
echo "source $GOPATH/src/github.com/bonnefoa/kubectl-fzf/kubectl_fzf.sh" >> ~/.zshrc
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
```

It will watch the cluster in the current context. If you switch context, `cache_builder` will detect and start watching the new cluster.
The initial resource listing can be long on big clusters and autocompletion might need 30s+.

`connect: connection refused` or similar messages are expected if there's network issues/interuptions and `cache_builder` will automatically reconnect.

### Troubleshooting

To launch with debug logs activated
```shell
cache_builder -logtostderr -v 14
```

### Watch a specific namespace

By default, all namespaces are watched. If you want to build the cache for a specific namespace, run
```shell
cache_builder -n mynamespace
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
| KUBECTL_FZF_ROLE_BLACKLIST  | List of roles to hide from node list (comma separated) | ""                                          |
| KUBECTL_FZF_EXCLUDE         | Exclusion patterns passed to the autocompletion        | ""                                          |
| KUBECTL_FZF_OPTIONS         | fzf parameters                                         | `-1 --header-lines=2 --layout reverse -e`   |

To turn down exact match in search:
```shell
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse)
```

To exclude all namespaces starting with "dev" and consul-agent resources:
```shell
export KUBECTL_FZF_EXCLUDE=("^dev" "consul-agent")
```

To hide `common_node` from the node's role list
```shell
export KUBECTL_FZF_ROLE_BLACKLIST="common_node"
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
