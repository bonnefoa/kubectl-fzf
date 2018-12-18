# Kubectl-fzf

kubectl-fzf provides a fast and powerful fzf autocompletion for kubectl.

[![asciicast](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja)](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja?t=01)

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
   * [Debugging](#debugging)

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

## `cache_builder`

`cache_builder` will watch cluster resources and keep the current state of the cluster in local files.

By default, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

To create cache files necessary for `kubectl_fzf`, just run

```shell
cache_builder
```

It will watch the cluster in the current context.
You can keep it running in a screen or a tmux.

### Watch a specific namespace

To create cache for a specific namespace, just run

```shell
cache_builder -n mynamespace
```

## `kubectl_fzf`

`kubectl_fzf.sh` overloads autocompletion function defined by `kubectl completion zsh` (or `kubectl completion bash`) to fzf with the local files to power autocompletion.

### fzf options

You can control used options for fzf with `KUBECTL_FZF_OPTIONS` variable.

To turn down exact match in search, set the variable to the following value

```shell
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=1 --layout reverse)
```

# Debugging

If files are empty, you can activate debugging logs with

```shell
cache_builder -logtostderr -v 10
```

# Caveat

With zsh, if the suggested completion doesn't match the start of the query, the completion will fail.

```shell
k get pod pr<TAB>
# result needs to starts with pr, otherwise, it will fail
```
