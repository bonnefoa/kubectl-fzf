kubectl-fzf overrides kubectl autocompletion functions with fzf using a local cache for speed.

# Requirements

- [fzf](https://github.com/junegunn/fzf)
- [kubectl shell autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion)

# Installation

## kubectl_fzf_cache_builder

```
pip2 install .
```

## kubectl_fzf

Add in your `.bashrc` or `.zshrc`

```
source <repository>/kubectl_fzf.sh
```

You can also use zplug
```
zplug "plugins/kubectl", from:oh-my-zsh, defer:2
zplug "bonnefoa/kubectl-fzf", defer:3
```

# Usage

## kubectl_fzf_cache_builder

To create cache files necessary for `kubectl_fzf`, just run

```
kubectl_fzf_cache_builder
```

It will watch the cluster and namespace in the current context.

If you have a custom login script, you can use

```
kubectl_fzf_cache_builder --refresh-command <script>
```

The script will be called to refresh oidc token when necessary.

## Available autocompletions

- deployments
- endpoints
- labels
- namespace
- nodes
- pods
- replicaset
- service
- statefulset
