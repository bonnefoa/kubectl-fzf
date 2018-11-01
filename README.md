kubectl-fzf overrides kubectl autocompletion functions with fzf using local cache for speed.

# Requirements

- [fzf](https://github.com/junegunn/fzf)
- [kubectl shell autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion)

# Installation

## kubectl_fzf_cache_builder

```
pip2 install
```

## kubectl_fzf

Add in your `.bashrc` or `.zshrc`

```
source <repository>/kubectl_fzf.sh
```

You can also use zplug
```
zplug "bonnefoa/kubectl-fzf"
```

# Usage

## kubectl_fzf_cache_builder

To create cache files necessary for `kubectl_fzf`, just run

```
kubectl_fzf_cache_builder
```

It will watch the cluster and namespace in the current context.

## Available autocompletions

- pod
- nodes
- deployment
- service
- labels
