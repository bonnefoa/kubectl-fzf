#!/bin/bash
set -eu

_fzf_get_port_forward_port()
{
    local context="$1"
    local port_file="$KUBECTL_FZF_CACHE/${context}_port"
    local global_port_file="$KUBECTL_FZF_CACHE/port"
    local local_port=$KUBECTL_FZF_PORT_FORWARD_START
    if [[ -f "$port_file" ]]; then
        local_port=$(cat $port_file)
    else
        if [[ -f "$global_port_file" ]]; then
            local_port=$(cat $global_port_file)
        fi
        echo $local_port > $port_file
        echo $((local_port + 1)) > $global_port_file
    fi
    echo $local_port
}

_fzf_check_port_forward_running()
{
    local local_port=$1
    if ! nc -G 1 -z localhost $local_port &> /dev/null; then
        return 1
    fi
    return 0
}

while true; do
    KUBECTL_FZF_PORT_FORWARD_START=${KUBECTL_FZF_PORT_FORWARD_START:-9873}
    KUBECTL_FZF_RSYNC_PORT=${KUBECTL_FZF_RSYNC_PORT:-80}
    context=$(kubectl config current-context)
    local_port=$(_fzf_get_port_forward_port $context)
    kfzf_ns=($(kubectl get svc --all-namespaces -l app=kubectl-fzf -o=jsonpath='{.items[0].metadata.namespace}' 2> /dev/null))
    log_file="$KUBECTL_FZF_CACHE/${context}_port_forward_log"

    if ! _fzf_check_port_forward_running $local_port; then
        echo "port forward not running, opening with port $local_port"
        nohup kubectl port-forward svc/kubectl-fzf -n ${kfzf_ns} ${local_port}:${KUBECTL_FZF_RSYNC_PORT} &> $log_file &
    fi
    sleep 10
done
