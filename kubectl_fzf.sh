export KUBECTL_FZF_CACHE="/tmp/kubectl_fzf_cache"
eval "`declare -f __kubectl_parse_get | sed '1s/.*/_&/'`"
eval "`declare -f __kubectl_get_containers | sed '1s/.*/_&/'`"
KUBECTL_FZF_OPTIONS=(-1 --header-lines=1 --layout reverse -e)
KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:3 --preview "echo {} | fold -w \$COLUMNS")

# $1 is awk end print command
# $2 is filename
# $3 is query
# $4 is an optional filter
_fzf_kubectl_complete()
{
    local end_print=$1
    local file="${KUBECTL_FZF_CACHE}/$2"
    local query=$3
    local filter=$4
    local end_field=$(awk 'NR==1{ for(i = 1; i <= NF; i++){ if ($i == "Labels") {print i - 1; } } } ' $file)
    local header=$(head -n1 "$file" | cut -d ' ' -f 1-$end_field)
    local rest=$(tail -n +2 "$file" | cut -d ' ' -f 1-$end_field | sort)
    if [[ -n $filter ]]; then
        rest=$(echo "$rest" | grep -w $filter)
    fi
    printf "$header\n$rest\n" \
        | column -t \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | awk "$end_print"
}

# $1 is filename
# $2 is query
_fzf_with_namespace()
{
    local namespace_in_query=$(__get_namespace_in_query)
    _fzf_kubectl_complete '{print $1 " " $2}' $1 "$2" "$namespace_in_query"
}

# $1 is filename
# $2 is query
_fzf_without_namespace()
{
    _fzf_kubectl_complete '{print $1}' $1 "$2"
}

# $1 is filename
# $2 is query
_flag_selector_with_namespace()
{
	local file="${KUBECTL_FZF_CACHE}/$1"
    awk '{split($NF,a,","); for (i in a) print $1 " " a[i]}' "$file" \
        | sort \
        | uniq \
        | column -t \
        | fzf ${KUBECTL_FZF_OPTIONS[@]} -q "$2" \
        | awk '{print $1 " " $2}'
}

# $1 is filename
# $2 is query
_flag_selector_without_namespace()
{
	local file="${KUBECTL_FZF_CACHE}/$1"
    awk '{split($NF,a,","); for (i in a) print a[i]}' "$file" \
        | sort \
        | uniq \
        | column -t \
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
    { echo "ContainerName"; echo "$containers"; } | fzf ${KUBECTL_FZF_OPTIONS[@]}
}

__get_current_namespace()
{
    local namespace=$(kubectl config view --minify --output 'jsonpath={..namespace}')
    echo "${namespace:-default}"
}

__get_namespace_in_query()
{
    local i=0
    for word in ${COMP_WORDS[@]} ; do
        if [[ $word == "-n" || $word == "--namespace" ]]; then
            if [[ ${#COMP_WORDS[@]} -gt $i && -n ${COMP_WORDS[$i + 1]} ]]; then
                echo ${COMP_WORDS[$i + 1]}
            fi
        fi
        ((i++))
    done
}

# $1 is result
__build_namespaced_compreply()
{
    local result=("$@")
    result=($(echo $result | tr " " "\n"))
    if [[ ${#result[@]} -eq 2 ]]; then
        # We have namespace in first position
        local current_namespace=$(__get_current_namespace)
        local namespace=${result[0]}
        if [[ $namespace != $current_namespace && $COMP_LINE != *" -n"* && "$COMP_LINE" != *" --namespace"* ]]; then
            COMPREPLY=( "${result[1]} -n ${result[0]}" )
        else
            COMPREPLY=( ${result[1]} )
        fi
    else
        COMPREPLY=( $result )
    fi
}

# $1 is the type of resource to get
__kubectl_parse_get()
{
	local penultimate=$(echo $COMP_LINE | awk '{print $(NF-1)}')
	local last_part=$(echo $COMP_LINE | awk '{print $(NF)}')

	local filename
	local autocomplete_fun
	local flag_autocomplete_fun
    local resource_name=$1

	case $resource_name in
		all )
			filename="pods"
            ;;
		pod | pods )
			filename="pods"
			autocomplete_fun=_fzf_with_namespace
			flag_autocomplete_fun=_flag_selector_with_namespace
			;;
		rs | resplicaset | replicasets )
			filename="replicasets"
			autocomplete_fun=_fzf_with_namespace
			flag_autocomplete_fun=_flag_selector_with_namespace
			;;
        configmap | configmaps )
			filename="configmaps"
			autocomplete_fun=_fzf_with_namespace
			flag_autocomplete_fun=_flag_selector_with_namespace
			;;
        ns | namespace | namespaces )
			filename="namespaces"
			autocomplete_fun=_fzf_without_namespace
			flag_autocomplete_fun=_flag_selector_without_namespace
			;;
		node | nodes )
			filename="nodes"
			autocomplete_fun=_fzf_without_namespace
			flag_autocomplete_fun=_flag_selector_without_namespace
			;;
        deployment | deployments | deployments. | deployments.apps | deployments.extensions  )
			filename="deployments"
			autocomplete_fun=_fzf_with_namespace
			flag_autocomplete_fun=_flag_selector_with_namespace
			;;
		sts | statefulset | statefulsets | statefulsets.apps  )
			filename="statefulsets"
			autocomplete_fun=_fzf_with_namespace
			flag_autocomplete_fun=_flag_selector_with_namespace
			;;
		persistentvolumes | pv )
			filename="persistentvolumes"
			autocomplete_fun=_fzf_without_namespace
			flag_autocomplete_fun=_flag_selector_without_namespace
			;;
		persistentvolumeclaims | pvc )
			filename="persistentvolumeclaims"
			autocomplete_fun=_fzf_with_namespace
			flag_autocomplete_fun=_flag_selector_with_namespace
			;;
		endpoints )
			filename="endpoints"
			autocomplete_fun=_fzf_with_namespace
			flag_autocomplete_fun=_flag_selector_with_namespace
			;;
		svc | service | services )
			filename="services"
			autocomplete_fun=_fzf_with_namespace
			flag_autocomplete_fun=_flag_selector_with_namespace
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
		result=$($flag_autocomplete_fun $filename $query)
        __build_namespaced_compreply "${result[@]}"
		return
	fi

    if [[ -z $autocomplete_fun ]]; then
        ___kubectl_parse_get $*
        return
    fi

	local query=""
	case $last_part in
        # Special cases: Sometime the last word doesn't match the resource name, don't put them in the query
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

	result=$($autocomplete_fun $filename $query)
	if [[ -z "$result" ]]; then
        return
	fi

    __build_namespaced_compreply "${result[@]}"
}

# Reregister complete function without '-o default' as we don't want to
# fallback to files and dir completion
complete -o nospace -F __start_kubectl kubectl
