# Copyright 2018 Anthonin Bonnefoy
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

export KUBECTL_FZF_CACHE="/tmp/kubectl_fzf_cache"
eval "`declare -f __kubectl_parse_get | sed '1s/.*/_&/'`"

_pod_selector()
{
    cut -d ' ' -f 1,2,4-7 ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf --sync -m --header="Namespace Name IP Node Status Age" --layout reverse -q "$2" \
        | awk '{print $2}'
}

_replicaset_selector()
{
    cut -d ' ' -f 1,2,4-8 ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf --sync -m --header="Namespace Name Desired Current Ready LabelSelector Age" --layout reverse -q "$2" \
        | awk '{print $2}'
}

_endpoints_selector()
{
    cut -d ' ' -f 1,2,4-5 ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf --sync -m --header="Namespace Name Age ReadyIPs ReadyPods UnreadyIPs NotReadyPods" --layout reverse -q "$2" \
        | awk '{print $2}'
}

_statefulset_selector()
{
    cut -d ' ' -f 1,2,4-5 ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf --sync -m --header="Namespace Name Ready/Replicas LabelSelector Age" --layout reverse -q "$2" \
        | awk '{print $2}'
}

_deployment_selector()
{
    cat ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf -m --header="Namespace Name Age Label" --layout reverse -q "$2" \
        | awk '{print $2}'
}

_namespace_selector()
{
    cat ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf -m --header="Namespace" --layout reverse -q "$2" \
        | awk '{print $1}'
}

_configmap_selector()
{
    awk '{print $1" "$2" "$4" "$3}' ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf -m --header="Namespace Name Age Labels" --layout reverse -q "$2" \
        | awk '{print $2}'
}

_service_selector()
{
    cut -d ' ' -f 1,2,4-7 ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf -m --header="Namespace Service Type Ip Ports Selector" --layout reverse -q "$2" \
        | awk '{print $2}'
}

_node_selector()
{
    awk '{print $1 " " $6 " " $5 " " $4 " " $7 " " $3}' ${KUBECTL_FZF_CACHE}/$1 \
        | column -t \
        | sort \
        | fzf -m --header="Node InternalIp Zone InstanceType Age Roles" --layout reverse -q "$2" \
        | awk '{print $1}'
}

_flag_selector()
{
	declare -A resources_to_label
	resources_to_label[pods]='3'
	resources_to_label[services]='3'
	resources_to_label[deployments]='3'
	resources_to_label[nodes]='2'
	resources_to_label[statefulsets]='3'
	resources_to_label[replicasets]='3'
	resources_to_label[configmaps]='3'
	resources_to_label[endpoints]='3'

	local file="${KUBECTL_FZF_CACHE}/$1"
	local column="${resources_to_label[$1]}"
    cut -d ' ' -f $column "$file" \
        | paste -sd ',' \
        | tr ',' '\n' \
        | grep -v None \
        | sort \
        | uniq \
        | fzf -m --header="Label Value" --layout reverse -q "$2" \
        | awk '{print $1}'
}

__kubectl_parse_get()
{
	local penultimate=$(echo $COMP_LINE | awk '{print $(NF-1)}')
	local last_part=$(echo $COMP_LINE | awk '{print $(NF)}')

	local filename
	local autocomplete_fun

	case $1 in
		pod?(s) )
			filename="pods"
			autocomplete_fun=_pod_selector
			;;
		rs | resplicaset?(s) )
			filename="replicasets"
			autocomplete_fun=_replicaset_selector
			;;
		configmap )
			filename="configmaps"
			autocomplete_fun=_configmap_selector
			;;
        ns | namespace?(s) )
			filename="namespaces"
			autocomplete_fun=_namespace_selector
			;;
		node?(s) )
			filename="nodes"
			autocomplete_fun=_node_selector
			;;
		deployment )
			filename="deployments"
			autocomplete_fun=_deployment_selector
			;;
		sts | statefulset )
			filename="statefulsets"
			autocomplete_fun=_statefulset_selector
			;;
		endpoints )
			filename="endpoints"
			autocomplete_fun=_endpoints_selector
			;;
		svc | service )
			filename="services"
			autocomplete_fun=_service_selector
			;;
		* )
			___kubectl_parse_get $*
			return
			;;
	esac

	if [[ $penultimate == "--selector" || $penultimate == "-l" || $last_part == "--selector" || $last_part == "-l" ]]; then
		if [[ $penultimate == "--selector" || $penultimate == "-l" ]]; then
			query=$last_part
		fi
		flags=$(_flag_selector $filename $query)
		if [[ -n $flags ]]; then
			COMPREPLY=( "$flags" )
		fi
		return
	fi

	local query=""
	if [[ $1 != $last_part && $last_part != -* ]]; then
		query=$last_part
	fi

	results=$( $autocomplete_fun $filename $query )
	if [[ -n "$results" ]]; then
		COMPREPLY=( $results )
	fi
}
