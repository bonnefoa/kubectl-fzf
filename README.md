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
   * [Caveat](#caveat)

# Features

- Seamless integration with kubectl autocompletion
- Sub second completion even with clusters > 1K pods
- Label autocompletion
- Automatic namespace switch

# Requirements

[fzf](https://github.com/junegunn/fzf) needs to be installed

# Installation

Install `cache_builder`

```shell
go get -u github.com/bonnefoa/kubectl-fzf/cmd/cache_builder
```

Source the autocompletion functions

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
The initial resource listing can be long  on big clusters and autocompletion might need 30s+.

To launch with debug logs activated

```shell
cache_builder -logtostderr -v 14
```

### Watch a specific namespace

To create cache for a specific namespace, just run

```shell
cache_builder -n mynamespace
```

## kubectl_fzf

Once `cache_builder` is running, you will be able to use `kubectl_fzf` by calling the kubectl completion

```shell
kubectl get pod <TAB>
```

### fzf options

You can control the options used for fzf with `KUBECTL_FZF_OPTIONS` variable.

For example, to turn down exact match in search:

```shell
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse)
```

# Caveat

With zsh, if the suggested completion doesn't match the start of the query, the completion will fail.

```shell
k get pod pr<TAB>
# result needs to starts with pr, otherwise, it will fail
```
