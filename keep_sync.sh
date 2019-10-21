#!/usr/bin/env bash
set -eu

. kubectl_fzf.sh

resources=("pods" "serviceaccounts" "daemonsets" "replicasets" "cronjobs" "horizontalpodautoscalers" "ingresses" "configmaps" "secrets" "namespaces" "nodes" "deployments" "statefulsets" "persistentvolumes" "persistentvolumeclaims" "endpoints" "services")

while true; do
    current_context=$(kubectl config current-context)
    local_port=$(_fzf_get_port_forward_port $current_context)
    kfzf_ns=($(kubectl get svc --all-namespaces -l app=kubectl-fzf -o=jsonpath='{.items[0].metadata.namespace}' 2> /dev/null))
    log_file="$KUBECTL_FZF_CACHE/${current_context}_port_forward_log"
    _fzf_fetch_rsynced_resource $current_context 15 ${resources[@]}
    sleep 15
done
