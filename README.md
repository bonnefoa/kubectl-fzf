# kubectl-fzf

kubectl-fzf overrides completion functions of kubectl with fzf to search for kubernetes resources using a local cache for speed.

## Requirements

- [fzf](https://github.com/junegunn/fzf)
- [kubectl shell autocompletion](https://kubernetes.io/docs/tasks/tools/install-kubectl/#enabling-shell-autocompletion)

## Installation

### Sourcing

Just source

```
source <repository>/kubectl_fzf.sh
```

### Zsh via zplug

To install kubectl-fzf via zplug. Add the following content to ~/.zshrc:
```
zplug "bonnefoa/kubectl-fzf"
```

## Usage

### kube_watcher.py

`kube_watcher.py` will build and maintain the local resource cache.

