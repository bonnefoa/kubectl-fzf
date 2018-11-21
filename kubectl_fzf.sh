export KUBECTL_FZF_CACHE="/tmp/kubectl_fzf_cache"
eval "`declare -f __kubectl_parse_get | sed '1s/.*/_&/'`"
eval "`declare -f __kubectl_get_containers | sed '1s/.*/_&/'`"
KUBECTL_FZF_OPTIONS=(-1 --header-lines=1 --layout reverse)
KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:3 --preview "echo {} | fold -w \$COLUMNS")

_pod_selector()
{
    cut -d ' ' -f 1-8 "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $2}'
}

_replicaset_selector()
{
    cut -d ' ' -f 1-7 "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $2}'
}

_endpoints_selector()
{
    cut -d ' ' -f 1-4 "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $2}'
}

_statefulset_selector()
{
    cut -d ' ' -f 1-4 "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $2}'
}

_deployment_selector()
{
    cat "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $2}'
}

_namespace_selector()
{
    cat "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $1}'
}

_configmap_selector()
{
    cat "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $2}'
}

_pv_selector()
{
    cut -d ' ' -f 1-6 "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $1}'
}

_pvc_selector()
{
    cut -d ' ' -f 1-7 "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $2}'
}

_service_selector()
{
    cut -d ' ' -f 1-6 "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $2}'
}

_node_selector()
{
    cut -d ' ' -f 1-6 "${KUBECTL_FZF_CACHE}/$1" \
        | column -t \
        | sort \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $1}'
}

_flag_selector()
{
	local file="${KUBECTL_FZF_CACHE}/$1"
    awk '{print $NF}' "$file" \
        | paste -sd ',' - \
        | tr ',' '\n' \
        | grep -v None \
        | sort \
        | uniq \
        | fzf ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $1}'
}

__kubectl_get_containers()
{
	local pod=$(echo $COMP_LINE | awk '{print $(NF)}')
    containers=$(awk "(\$2 == \"$pod\") {print \$7}" ${KUBECTL_FZF_CACHE}/pods \
        | tr ',' '\n' \
        | sort)
    if [[ $containers == "" ]]; then
        ___kubectl_get_containers $*
        return
    fi
    { echo "ContainerName"; echo "$containers" } | fzf ${KUBECTL_FZF_OPTIONS[@]}
}

# $1 is the type of resource to get
__kubectl_parse_get()
{
	local penultimate=$(echo $COMP_LINE | awk '{print $(NF-1)}')
	local last_part=$(echo $COMP_LINE | awk '{print $(NF)}')

	local filename
	local autocomplete_fun
    local resource_name=$1

	case $resource_name in
		all )
			filename="pods"
            ;;
		pod | pods )
			filename="pods"
			autocomplete_fun=_pod_selector
			;;
		rs | resplicaset | replicasets )
			filename="replicasets"
			autocomplete_fun=_replicaset_selector
			;;
        configmap | configmaps )
			filename="configmaps"
			autocomplete_fun=_configmap_selector
			;;
        ns | namespace | namespaces )
			filename="namespaces"
			autocomplete_fun=_namespace_selector
			;;
		node | nodes )
			filename="nodes"
			autocomplete_fun=_node_selector
			;;
        deployment | deployments | deployments.apps | deployments.extensions  )
			filename="deployments"
			autocomplete_fun=_deployment_selector
			;;
		sts | statefulset | statefulsets | statefulsets.apps  )
			filename="statefulsets"
			autocomplete_fun=_statefulset_selector
			;;
		pv )
			filename="pvs"
			autocomplete_fun=_pv_selector
			;;
		pvc )
			filename="pvcs"
			autocomplete_fun=_pvc_selector
			;;
		endpoints )
			filename="endpoints"
			autocomplete_fun=_endpoints_selector
			;;
		svc | service | services )
			filename="services"
			autocomplete_fun=_service_selector
			;;
		* )
			___kubectl_parse_get $*
			return
			;;
	esac

	if [[ $penultimate == "--selector" || $penultimate == "-l" || $last_part == "--selector" || $last_part == "-l" ]]; then
        if [[ ($penultimate == "--selector" || $penultimate == "-l") && ${COMP_LINE: -1} == " " ]]; then
            return
        fi
		if [[ $penultimate == "--selector" || $penultimate == "-l" ]]; then
			query=$last_part
		fi
		flags=$(_flag_selector $filename $query)
		if [[ -n $flags ]]; then
			COMPREPLY=( "$flags" )
		fi
		return
	fi

    if [[ -z $autocomplete_fun ]]; then
        ___kubectl_parse_get $*
        return
    fi

	local query=""
	case $last_part in
        # Special cases: Some last words doesn't match the resource name, don't put them in the query
		exec )
            ;;
		cp )
            ;;
		logs | log )
            ;;
        *)
            if [[ $resource_name != $last_part && $last_part != -* && ${COMP_LINE: -1} != " " ]]; then
                query=$last_part
            fi
    esac

	results=$( $autocomplete_fun $filename $query )
	if [[ -n "$results" ]]; then
		COMPREPLY=( $results )
	fi
}
