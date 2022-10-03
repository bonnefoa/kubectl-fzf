# Kubectl-fzf

kubectl-fzf provides a fast and powerful fzf autocompletion for kubectl.

[![asciicast](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja.png)](https://asciinema.org/a/yHKY5vQ40ZaOwMQnhLfYJ5Pja?t=01)

Table of Contents
=================

* [Kubectl-fzf](#kubectl-fzf)
* [Table of Contents](#table-of-contents)
* [Features](#features)
* [Requirements](#requirements)
* [Installation](#installation)
   * [kubectl-fzf binaries](#kubectl-fzf-binaries)
   * [Shell autocompletion](#shell-autocompletion)
      * [Zsh plugins: Antigen](#zsh-plugins-antigen)
   * [kubectl-fzf-server](#kubectl-fzf-server)
      * [Install kubectl-fzf-server as a pod](#install-kubectl-fzf-server-as-a-pod)
      * [Install kubectl-fzf-server as a systemd service](#install-kubectl-fzf-server-as-a-systemd-service)
* [Usage](#usage)
   * [kubectl-fzf-server: local version](#kubectl-fzf-server-local-version)
   * [kubectl-fzf-server: pod version](#kubectl-fzf-server-pod-version)
   * [Completion](#completion)
      * [Configuration](#configuration)
* [Troubleshooting](#troubleshooting)
   * [Debug kubectl-fzf-completion](#debug-kubectl-fzf-completion)
   * [Debug Tab Completion](#debug-tab-completion)
   * [Debug kubectl-fzf-server](#debug-kubectl-fzf-server)

# Features

- Seamless integration with kubectl autocompletion
- Fast completion
- Label autocompletion
- Automatic namespace switch

# Requirements

- go (minimum version 1.19)
- awk
- [fzf](https://github.com/junegunn/fzf)

# Installation

## kubectl-fzf binaries

```shell
# Completion binary called during autocompletion
go install github.com/bonnefoa/kubectl-fzf/v3/cmd/kubectl-fzf-completion@main
# If you want to run the kubectl-fzf server locally
go install github.com/bonnefoa/kubectl-fzf/v3/cmd/kubectl-fzf-server@main
```

`kubectl-fzf-completion` needs to be in you $PATH so make sure that your $GOPATH bin is included:
```
PATH=$PATH:$GOPATH/bin
```

## Shell autocompletion

Source the autocompletion functions:
```
# bash version
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/main/shell/kubectl_fzf.bash -O ~/.kubectl_fzf.bash
echo "source <(kubectl completion bash)" >> ~/.bashrc
echo "source ~/.kubectl_fzf.bash" >> ~/.bashrc

# zsh version
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/main/shell/kubectl_fzf.plugin.zsh -O ~/.kubectl_fzf.plugin.zsh
echo "source <(kubectl completion zsh)" >> ~/.zshrc
echo "source ~/.kubectl_fzf.plugin.zsh >> ~/.zshrc
```

### Zsh plugins: Antigen

You can use [antigen](https://github.com/zsh-users/antigen) to load it as a zsh plugin
```shell
antigen bundle robbyrussell/oh-my-zsh plugins/docker
antigen bundle bonnefoa/kubectl-fzf@main shell/
```

## kubectl-fzf-server

### Install kubectl-fzf-server as a pod

You can deploy `kubectl-fzf-server` as a pod in your cluster.

From the [k8s directory](https://github.com/bonnefoa/kubectl-fzf/tree/main/k8s):
```shell
helm template --namespace myns --set image.kubectl_fzf_server.tag=v3 --set toleration=aToleration . | kubectl apply -f -
```

You can check the latest image version [here](https://cloud.docker.com/repository/docker/bonnefoa/kubectl-fzf/general).

### Install kubectl-fzf-server as a systemd service

You can install `kubectl-fzf-server` as a systemd unit server.

```
# Create user systemd config
mkdir -p ~/.config/systemd/user
wget https://raw.githubusercontent.com/bonnefoa/kubectl-fzf/main/systemd/kubectl_fzf_server.service -O ~/.config/systemd/user/kubectl_fzf_server.service
# Set fullpath of kubectl-fzf-server
sed -i "s#INSTALL_PATH#$GOPATH/bin#" ~/.config/systemd/user/kubectl_fzf_server.service

# Reload to pick up new service
systemctl --user daemon-reload

# Start the server
systemctl --user start kubectl_fzf_server.service

# Automatically enable it at startup
systemctl --user enable kubectl_fzf_server.service

# Get log
journalctl --user-unit=kubectl_fzf_server.service
```

# Usage

## kubectl-fzf-server: local version

``` mermaid
flowchart TB
    subgraph TargetCluster
        k8s[api-server]
    end

    subgraph Laptop
        shell[Shell]
        fileNode([/tmp/kubectl_fzf_cache/TargetCluster/pods])
        comp[kubectl-fzf-completion]
        server[kubectl-fzf-server]
    end
    shell -- kubectl get pods TAB --> comp -- Read content and feed it to fzf --> fileNode
    server -- Write autocompletion informations --> fileNode

    k8s <-- Watch --o server
```

`kubectl-fzf-server` will watch cluster resources and keep the current state of the cluster in local files.
By default, files are written in `/tmp/kubectl_fzf_cache` (defined by `KUBECTL_FZF_CACHE`)

Advantages:
- Minimal setup needed.
- Local cache is maintained up to date.

Drawbacks:
- It can be CPU and memory intensive on big clusters.
- It also can be bandwidth intensive. The most expensive is the initial listing at startup and on error/disconnection. Big namespace can increase the probability of errors during initial listing.
- It can generate load on the kube-api servers if multiple user are running it.

To create cache files necessary for `kubectl_fzf`, just run in a tmux or a screen

```shell
kubectl-fzf-server
```

It will watch the cluster in the current context. If you switch context, `kubectl-fzf-server` will detect and start watching the new cluster.
The initial resource listing can be long on big clusters and autocompletion might need 30s+.

`connect: connection refused` or similar messages are expected if there's network issues/interruptions and `kubectl-fzf-server` will automatically reconnect.

## kubectl-fzf-server: pod version

``` mermaid
flowchart TB
    subgraph TargetCluster
        k8s[api-server]
        server[kubectl-fzf-server]
    end

    subgraph Laptop
        shell[Shell]
        comp[kubectl-fzf-completion]
    end


    shell -- kubectl get pods TAB --> comp 
    comp -- Through port forward\nGET /k8s/resources/pods --> server

    k8s <-- Watch --o server
```

If the pod is deployed in your cluster, the autocompletion will be fetched automatically fetched using port forward.

Advantages:
- No need to run a local `kubectl-fzf-server`
- Only a single instance of `kubectl-fzf-server` per cluster is needed, lowering the load on the `kube-api` servers.

Drawbacks:
- Resources need to be fetched remotely, this can increased the completion time. A local cache is maintained to lower this.

## Completion

Once `kubectl-fzf-server` is running, you will be able to use `kubectl_fzf` by calling the kubectl completion
```shell
# Get fzf completion on pods on all namespaces
kubectl get pod <TAB>

# Open fzf autocompletion on all available label
kubectl get pod -l <TAB>

# Open fzf autocompletion on all available field-selector. Usually much faster to list all pods running on an host compared to kubectl describe node.
kubectl get pod --field-selector <TAB>

# This will fallback to the normal kubectl completion (if sourced) 
kubectl <TAB>
```

### Configuration

By default, the local port used for the port-forward is 8080. You can override it through an environment variable:
```
KUBECTL_FZF_PORT_FORWARD_LOCAL_PORT=8081
```

# Troubleshooting

## Debug kubectl-fzf-completion

Build and test a completion with debug logs:
```
go build ./cmd/kubectl-fzf-completion && KUBECTL_FZF_LOG_LEVEL=debug ./kubectl-fzf-completion k8s_completion 'get pods '  
```

Force Tab completion to use the completion binary in the current directory:
```
export KUBECTL_FZF_COMPLETION_BIN=./kubectl-fzf-completion
```

## Debug Tab Completion

To debug Tab completion, you can activate the shell debug logs:
```
export KUBECTL_FZF_COMP_DEBUG_FILE=/tmp/debug
```

Check that the completion function is correctly sourced:
```
type kubectl_fzf_completion
kubectl_fzf_completion is a shell function from /home/bonnefoa/.antigen/bundles/kubectl-fzf-main/shell/kubectl_fzf.plugin.zsh
```

Use zsh completion debug:
```
kubectl get pods <C-X>?
Trace output left in /tmp/zsh497886kubectl1 (up-history to view)
```

## Debug kubectl-fzf-server

To launch kubectl-fzf-server with debug logs
```shell
kubectl-fzf-server --log-level debug
```
