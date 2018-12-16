kubectl-fzf provides a fast kubectl autocompletion with fzf.

Table of Contents
=================

   * [Requirements](#requirements)
   * [Pros](#pros)
   * [Installation](#installation)
      * [Using zplug](#using-zplug)
   * [Usage](#usage)
      * [cache_builder](#cache_builder)
         * [Watch a specific namespace](#watch-a-specific-namespace)
      * [kubectl_fzf](#kubectl_fzf)
         * [fzf options](#fzf-options)
   * [Debugging](#debugging)

# Requirements

- [fzf](https://github.com/junegunn/fzf)
- [kubectl shell autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion)

# Pros

- Seamless integration with kubectl autocomplete
- Scale with clusters > 1K pods
- Label autocompletion
- Automatic namespace switch

# Installation

Install `cache_builder`

```
go get -u github.com/bonnefoa/kubectl-fzf/cmd/cache_builder
```

Source the autocompletion functions
```
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
```
zplug "plugins/kubectl", from:oh-my-zsh, defer:2
zplug "bonnefoa/kubectl-fzf", defer:3
```

# Usage

## `cache_builder`

`cache_builder` will watch cluster resources and keep the current state of the cluster in local files.

By default, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

To create cache files necessary for `kubectl_fzf`, just run

```
cache_builder
```

It will watch the cluster in the current context. You can keep it running in a screen or a tmux.

### Watch a specific namespace

To create cache for a specific namespace, just run

```
cache_builder -n mynamespace
```

## `kubectl_fzf`

`kubectl_fzf.sh` overloads autocompletion function defined by `kubectl completion zsh` (or `kubectl completion bash`) to fzf with the local files to power autocompletion.

### fzf options

You can control used options for fzf with `KUBECTL_FZF_OPTIONS` variable.

For example, to force exact match in search, set the variable to the following value
```
export KUBECTL_FZF_OPTIONS=(-1 --header-lines=1 --layout reverse -e)
```

# Debugging

If files are not empty, you can activate debugging logs with

```
cache_builder -logtostderr -v 10
```
