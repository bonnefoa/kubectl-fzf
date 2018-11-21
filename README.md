`kubectl-fzf` provides a fast kubectl autocompletion with fzf.

`kubectl_fzf_cache_builder` will watch cluster resources and keep the current state of the cluster in local files. By defaults, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

`kubectl_fzf.sh` overloads `__kubectl_parse_get` function defined by `kubectl completion zsh` (or `kubectl completion bash`) to fzf with the local files to power autocompletion.

# Requirements

# Requirements

- [fzf](https://github.com/junegunn/fzf)
- [kubectl shell autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion)

# Installation

## kubectl_fzf_cache_builder

```
pip2 install .
```

## kubectl_fzf

Source the `kubectl_fzf.sh` file in your `.bashrc` or `.zshrc`

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
