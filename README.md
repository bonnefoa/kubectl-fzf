kubectl-fzf provides a fast kubectl autocompletion with fzf.

Table of Contents
=================

   * [Requirements](#requirements)
   * [Pros](#pros)
   * [Table of Contents](#table-of-contents)
   * [Installation](#installation)
      * [Using zplug](#using-zplug)
   * [Usage](#usage)
      * [kubectl_fzf_cache_builder](#kubectl_fzf_cache_builder)
         * [Watch all namespaces](#watch-all-namespaces)
         * [Refresh](#refresh)
      * [kubectl_fzf](#kubectl_fzf)

# Requirements

- [fzf](https://github.com/junegunn/fzf)
- [kubectl shell autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion)

# Pros

- Seamless integration
- Scale with clusters > 1K pods
- Provide label autocompletion

# Installation

Install `kubectl_fzf_cache_builder` script

```
git clone --depth 1 https://github.com/bonnefoa/kubectl-fzf
pip2 install -U kubectl-fzf/
```

Source the autocompletion functions
```
# kubectl_fzf.sh needs to be sourced after kubectl completion.

# bash version
echo "source <(kubectl completion bash)" >> ~/.bashrc
echo "source $PWD/kubectl-fzf/kubectl_fzf.sh" >> ~/.bashrc

# zsh version
echo "source <(kubectl completion zsh)" >> ~/.zshrc
echo "source $PWD/kubectl-fzf/kubectl_fzf.sh" >> ~/.zshrc
```

## Using zplug

You can use zplug to install the autocompletion functions
```
zplug "plugins/kubectl", from:oh-my-zsh, defer:2
zplug "bonnefoa/kubectl-fzf", defer:3
```

# Usage

## `kubectl_fzf_cache_builder`

`kubectl_fzf_cache_builder` will watch cluster resources and keep the current state of the cluster in local files.

By default, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

To create cache files necessary for `kubectl_fzf`, just run

```
kubectl_fzf_cache_builder
```

It will watch the cluster and namespace in the current context.

### Watch all namespaces

To create cache for all namespaces of the current cluster, just run

```
kubectl_fzf_cache_builder --all-namespaces
```

### Refresh

If you have a custom login script, you can use

```
kubectl_fzf_cache_builder --refresh-command <script>
```

The script will be called to refresh oidc token when necessary.

## `kubectl_fzf`

`kubectl_fzf.sh` overloads autocompletion function defined by `kubectl completion zsh` (or `kubectl completion bash`) to fzf with the local files to power autocompletion.
