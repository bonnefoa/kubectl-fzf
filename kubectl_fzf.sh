export KUBECTL_FZF_CACHE="/tmp/kubectl_fzf_cache"
eval "`declare -f __kubectl_parse_get | sed '1s/.*/_&/'`"
eval "`declare -f __kubectl_get_containers | sed '1s/.*/_&/'`"
KUBECTL_FZF_EXCLUDE=${KUBECTL_FZF_EXCLUDE:-}
KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse -e)
KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:3 --preview "echo {} | tr -s '\t ' | fold -s -w \$COLUMNS")

# $1 is filename
_fzf_get_label_field()
{
    awk 'NR==1{ for(i = 1; i <= NF; i++){ if ($i == "Labels") {print i; } } } ' $1
}

# $1 is context
# $2 is namespace
_fzf_get_main_header()
{
    local context="$1"
    local namespace="$2"
    local main_header="Context:$context"
    if [[ -n $namespace ]]; then
        main_header="$main_header, Namespace:$namespace"
    fi
    echo $main_header
}

# $1 is awk end print command
# $2 isFlag
# $3 is filepath
# $4 is context
# $5 is query
# $6 optional namespace
_fzf_kubectl_complete()
{
    local end_print=$1
    local is_flag="$2"
    local file="$3"
    local header_file="$3_header"
    local context="$4"
    local query=$5
    local namespace="$6"
    local label_field=$(_fzf_get_label_field $header_file)
    local end_field=$((label_field - 1))
    local main_header=$(_fzf_get_main_header $context $namespace)

    if [[ $is_flag == "true" ]]; then
        local header=$(cat "$header_file" | cut -d ' ' -f 1,$label_field)
        local data=$(cat "$file" | awk '{split($NF,a,","); for (i in a) print $1 " " a[i]}' | sort | uniq)
    else
        local header=$(cat "$header_file" | cut -d ' ' -f 1-$end_field)
        local data=$(cat "$file" | cut -d ' ' -f 1-$end_field)
    fi
    if [[ -n $namespace ]]; then
        data=$(echo "$data" | grep -w "^$namespace")
    fi
    for pattern in ${KUBECTL_FZF_EXCLUDE[@]}; do
        data=$(echo "$data" | grep -v "$pattern")
    done
    data=$(printf "$header\n$data\n" | column -t)

    printf "${main_header}\n$data" \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | awk "$end_print"
}

# $1 is filepath
# $2 is context
# $3 is query
_fzf_with_namespace()
{
    local namespace_in_query=$(__get_parameter_in_query "--namespace -n")
    _fzf_kubectl_complete '{print $1 " " $2}' "false" $1 "$2" "$3" "$namespace_in_query"
}

# $1 is filepath
# $2 is context
# $3 is query
_fzf_without_namespace()
{
    _fzf_kubectl_complete '{print $1}' "false" $1 "$2" "$3"
}

# $1 is filepath
# $2 is context
# $3 is query
_flag_selector_with_namespace()
{
    local namespace_in_query=$(__get_parameter_in_query "--namespace -n")
    _fzf_kubectl_complete '{print $1 " " $2}' "true" $1 "$2" "$3" "$namespace_in_query"
}

# $1 is filepath
# $2 is query
# $3 is context
_flag_selector_without_namespace()
{
    _fzf_kubectl_complete '{print $2}' "true" $1 "$2" "$3"
}

__kubectl_get_containers()
{
    local pod=$(echo $COMP_LINE | awk '{print $(NF)}')
    local current_context=$(kubectl config current-context)
    local main_header=$(_fzf_get_main_header $current_context "")
    local data=$(awk "(\$2 == \"$pod\") {print \$7}" ${KUBECTL_FZF_CACHE}/${current_context}/pods \
        | tr ',' '\n' \
        | sort)
    if [[ $data == "" ]]; then
        ___kubectl_get_containers $*
        return
    fi
    printf "ContainerName\n${main_header}\n${data}" \
        | fzf ${KUBECTL_FZF_OPTIONS[@]}
}

__get_current_namespace()
{
    local namespace=$(kubectl config view --minify --output 'jsonpath={..namespace}')
    echo "${namespace:-default}"
}

__get_parameter_in_query()
{
    local parameter_names="$1"
    local i=0
    for word in ${COMP_WORDS[@]} ; do
        for parameter in $parameter_names ; do
            if [[ $word == $parameter ]]; then
                if [[ ${#COMP_WORDS[@]} -gt $i && -n ${COMP_WORDS[$i + 1]} ]]; then
                    echo ${COMP_WORDS[$i + 1]}
                fi
            fi
        done
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
    local current_context=$(kubectl config current-context)

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
        ds | daemonset | daemonsets | daemonsets.apps | daemonsets.extensions | daemonsets. )
            filename="daemonsets"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        rs | resplicaset | replicasets )
            filename="replicasets"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        ingress | ingresses | ingresses. | ingresses.extensions )
            filename="ingresses"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        cm | configmap | configmaps )
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

    local query_context=$(__get_parameter_in_query "--context")
    local context=$current_context
    if [[ -n $query_context && $query_context != $current_context ]]; then
        context=$query_context
    fi

    filepath="${KUBECTL_FZF_CACHE}/${context}/${filename}"

    if [[ ! -f $filepath ]]; then
        ___kubectl_parse_get $*
        return
    fi

    if [[ $penultimate == "--selector" || $penultimate == "-l" || $last_part == "--selector" || $last_part == "-l" ]]; then
        if [[ ($penultimate == "--selector" || $penultimate == "-l") && ${COMP_LINE: -1} == " " ]]; then
            return
        fi
        if [[ $penultimate == "--selector" || $penultimate == "-l" ]]; then
            query=$last_part
        fi
        result=$($flag_autocomplete_fun $filepath $context $query)
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

    result=$($autocomplete_fun $filepath $context $query)
    if [[ -z "$result" ]]; then
        return
    fi

    __build_namespaced_compreply "${result[@]}"
}

# Reregister complete function without '-o default' as we don't want to
# fallback to files and dir completion
complete -F __start_kubectl kubectl
