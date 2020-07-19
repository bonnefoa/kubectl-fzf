export KUBECTL_FZF_CACHE="/tmp/kubectl_fzf_cache"
eval "`declare -f __kubectl_parse_get | sed '1s/.*/_&/'`"
eval "`declare -f __kubectl_parse_resource | sed '1s/.*/_&/'`"
eval "`declare -f __kubectl_get_containers | sed '1s/.*/_&/'`"
eval "`declare -f __kubectl_get_resource | sed '1s/.*/_&/'`"
eval "`declare -f __kubectl_handle_filename_extension_flag | sed '1s/.*/_&/'`"
KUBECTL_FZF_EXCLUDE=${KUBECTL_FZF_EXCLUDE:-}
KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse -e --no-hscroll --no-sort)
# Cache time when no rsync service was detected
KUBECTL_FZF_RSYNC_NO_SERVICE_CACHE_TIME=${KUBECTL_FZF_RSYNC_NO_SERVICE_CACHE_TIME:-3600}
# Cache time of api resource list
KUBECTL_FZF_RSYNC_API_RESOURCE_CACHE_TIME=${KUBECTL_FZF_RSYNC_API_RESOURCE_CACHE_TIME:-3600}
# Cache time of every other resources
KUBECTL_FZF_RSYNC_RESOURCE_CACHE_TIME=${KUBECTL_FZF_RSYNC_RESOURCE_CACHE_TIME:-30}
KUBECTL_FZF_RSYNC_PORT=${KUBECTL_FZF_RSYNC_PORT:-80}
mkdir -p $KUBECTL_FZF_CACHE

# $1 is filename
# $2 is header
_fzf_get_header_position()
{
    awk "NR==1{ for(i = 1; i <= NF; i++){ if (\$i == \"$2\") {print i; } } } " $1
}

_fzf_file_mtime_older_than()
{
    local file=$1
    local cache_time=$2
    if [[ ! -f "$file" ]]; then
        return 0
    fi

    local mtime=$(date -r "$file" +%s)
    local current=$(date +%s)
    if [[ $((current - mtime)) -gt $cache_time ]]; then
        return 0
    fi

    return 1
}

# $1 is context
# $2 is cache time
# $3 is resource name
_fzf_fetch_resource()
{
    local context=$1
    local cache_time=$2
    shift 2
    local resources=($@)

    if pgrep -x cache_builder > /dev/null; then
        return
    fi

    mkdir -p "${KUBECTL_FZF_CACHE}/${context}"
    for resource in ${resources[@]} ; do
        local check_time_file="${KUBECTL_FZF_CACHE}/${context}/_${resource}"
        if ! $(_fzf_file_mtime_older_than $check_time_file $cache_time); then
            return
        fi
        touch "$check_time_file"
    done

    local include_param=()
    for resource_name in ${resources[@]} ; do
        include_param+="--include=${resource_name}* "
    done

    local rsync_endpoint=($(_fzf_check_direct_access $context))
    if [[ -z "${rsync_endpoint[@]}" ]]; then
        rsync_endpoint=($(_fzf_check_port_forward $context))
    fi

    if [[ -n "$rsync_endpoint" ]]; then
        rsync -qPrz --delete ${include_param[@]} --timeout=1 --exclude="*" "rsync://${rsync_endpoint[@]:0:1}:${rsync_endpoint[@]:1:1}/fzf_cache/" "${KUBECTL_FZF_CACHE}/${context}/"
    fi

#    mkdir -p "${KUBECTL_FZF_CACHE}/${context}"
#    for resource_name in ${resources[@]} ; do
#        local check_time_file="${KUBECTL_FZF_CACHE}/${context}/_${resource_name}"
#        if ! $(_fzf_file_mtime_older_than $check_time_file $cache_time); then
#            return
#        fi
#
#        local header_file="${KUBECTL_FZF_CACHE}/${context}/${resource_name}_header"
#        if [[ ! -f "$header_file" && $resource_name != "apiresources" ]]; then
#            curl -s ${rsync_endpoint[2]}://${rsync_endpoint[0]}:${rsync_endpoint[1]}/header/$resource_name > "$header_file"
#        fi
#
#        local resource_file="${KUBECTL_FZF_CACHE}/${context}/${resource_name}_resource"
#        curl -s ${rsync_endpoint[2]}://${rsync_endpoint[0]}:${rsync_endpoint[1]}/resource/$resource_name > "$resource_file"
#
#        local label_file="${KUBECTL_FZF_CACHE}/${context}/${resource_name}_label"
#        curl -s ${rsync_endpoint[2]}://${rsync_endpoint[0]}:${rsync_endpoint[1]}/label/$resource_name > "$label_file"
#        touch -m "$check_time_file"
#    done

}

_fzf_check_connection()
{
    local ip="$1"
    local port="$2"

    local nc_options=""
    if [ ! "`nc -G 1 2>&1 | grep invalid`" ]; then
        nc_options="-G 1"
    fi
    if nc $nc_options -w 1 -z $ip $port &>/dev/null; then
        return 0
    fi
    return 1
}

_fzf_check_direct_access()
{
    local context="$1"
    local endpoint_file="$KUBECTL_FZF_CACHE/${context}_cache_endpoint"
    if [[ -s "$endpoint_file" ]]; then
        local cached_ip=$(cat "$endpoint_file")
        if [[ "$cached_ip" == "No service" ]]; then
            if ! $(_fzf_file_mtime_older_than $endpoint_file $KUBECTL_FZF_RSYNC_NO_SERVICE_CACHE_TIME); then
                return
            fi
        fi

        if _fzf_check_connection $cached_ip ${KUBECTL_FZF_RSYNC_PORT}; then
            echo $cached_ip > "$endpoint_file"
            echo "$cached_ip $KUBECTL_FZF_RSYNC_PORT"
            return
        fi
    fi
    for ip in $(kubectl get endpoints -l app=kubectl-fzf --all-namespaces -o=jsonpath='{.items[*].subsets[*].addresses[*].ip}'); do
        if _fzf_check_connection $ip ${KUBECTL_FZF_RSYNC_PORT}; then
            echo $ip > "$endpoint_file"
            echo "$ip $KUBECTL_FZF_RSYNC_PORT"
            return
        fi
    done
    echo "No service" > "$endpoint_file"
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

_fzf_get_exclude_pattern()
{
    local grep_exclude=""
    for pattern in ${KUBECTL_FZF_EXCLUDE[@]}; do
        if [[ -z $grep_exclude ]]; then
            grep_exclude=$pattern
        else
            grep_exclude="$grep_exclude\|$pattern"
        fi
    done
    echo "$grep_exclude"
}

_fzf_get_node_to_pods()
{
    local context="$1"
    local pod_file="${KUBECTL_FZF_CACHE}/${context}/pods_resource"
    local pod_header_file="${KUBECTL_FZF_CACHE}/${context}/pods_header"
    local node_name_field=$(_fzf_get_header_position $pod_header_file "NodeName")
    local pod_name_field=$(_fzf_get_header_position $pod_header_file "Name")

    local daemonsets_file="${KUBECTL_FZF_CACHE}/${context}/daemonsets"
    local daemonsets=$(cut -d' ' -f2 "$daemonsets_file" | uniq)
    local exclude_pods=""
    for daemonset in $daemonsets ; do
        if [[ -z $exclude_pods ]]; then
            exclude_pods=$daemonset
        else
            exclude_pods="$exclude_pods\|$daemonset"
        fi
    done

    grep -v $exclude_pods $pod_file \
        | awk "{ if(a[\$$node_name_field]==\"\") {a[\$$node_name_field]=\$$pod_name_field} else { a[\$$node_name_field]=\$$pod_name_field \":\" a[\$$node_name_field] } } END { for (i in a) { print i \" \"  substr(a[i], 0, 1800) } } "
}

_fzf_kubectl_pv_complete()
{
    local pv_file="$1"
    local context="$2"
    local query="$3"

    local main_header=$(_fzf_get_main_header $context $namespace)

    local pv_header_file="${pv_file}_header"
    local label_field=$(_fzf_get_header_position $pv_header_file "Labels")
    local claim_field_pv_file=$(_fzf_get_header_position $pv_header_file "Claim")
    local end_field=$((label_field - 1))
    local header=$(cut -d ' ' -f 1-$end_field "$pv_header_file")
    header="$header MountedBy"

    local pod_file="${KUBECTL_FZF_CACHE}/${context}/pods_resource"
    local pod_header_file="${KUBECTL_FZF_CACHE}/${context}/pods_header"
    local claim_field_pod_file=$(_fzf_get_header_position $pod_header_file "Claims")
    local pod_name_field=$(_fzf_get_header_position $pod_header_file "Name")
    local pod_namespace_field=$(_fzf_get_header_position $pod_header_file "Namespace")

    local data=$(join -a1 -o'1.1,1.2,1.3,1.4,1.5,1.6,1.7,1.8,2.2' -1 $claim_field_pv_file -2 1 -e None \
        <(cut -d ' ' -f 1-$end_field "$pv_file" | sort -k $claim_field_pv_file) \
        <(awk "(\$$claim_field_pod_file != \"None\"){split(\$$claim_field_pod_file,c,\",\"); for (i in c) { print c[i] \" \" \$$pod_namespace_field\"/\"\$$pod_name_field } }" $pod_file | sort))
    local num_fields=$(echo $header | wc -w | sed 's/  *//g')

    KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:$num_fields --preview "echo -e \"${header}\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=\$i} (NR==2){for (i in a) {printf a[i] \": \" \$i \"\n\"} }' | column -t | fold -w \$COLUMNS" )
    (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | cut -d' ' -f1
}

_fzf_kubectl_node_complete()
{
    local node_file="$1"
    local context="$2"
    local query="$3"

    local main_header=$(_fzf_get_main_header $context $namespace)
    local node_header_file="${node_file}_header"
    local label_field=$(_fzf_get_header_position $node_header_file "Labels")
    local end_field=$((label_field - 1))
    local header=$(cut -d ' ' -f 1-$end_field "$node_header_file")
    header="$header Pods"

    local node_to_pods=$(_fzf_get_node_to_pods $context)
    local data=$(join -a1 -oauto -e None <(cut -d ' ' -f 1-$end_field "$node_file") <(echo "$node_to_pods"))
    local num_fields=$(echo $header | wc -w | sed 's/  *//g')
    KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:$num_fields --preview "echo -e \"${header}\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=\$i} (NR==2){for (i in a) {printf a[i] \": \" \$i \"\n\"} }' | column -t | fold -w \$COLUMNS" )
    (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | cut -d' ' -f1
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
    local prefix="$3"
    local resource_file="${prefix}_resource"
    local header_file="${prefix}_header"
    local label_file="${prefix}_label"
    local context="$4"
    local query=$5
    local namespace="$6"
    local label_field=$(_fzf_get_header_position $header_file "Labels")
    local end_field=$((label_field - 1))
    local main_header=$(_fzf_get_main_header $context $namespace)

    if [[ $is_flag == "with_namespace" ]]; then
        local header="Namespace Labels Occurrences"
        local data=$(cat $label_file)
    elif [[ $is_flag == "without_namespace" ]]; then
        local header="Labels Occurrences"
        local data=$(cat $label_file)
    else
        local header=$(cut -d ' ' -f 1-$end_field "$header_file")
        local data=$(cut -d ' ' -f 1-$end_field $resource_file)
    fi

    if [[ -n $namespace ]]; then
        data=$(echo "$data" | grep -w "^$namespace")
    fi

    local grep_exclude=$(_fzf_get_exclude_pattern)
    if [[ -n $grep_exclude ]]; then
        data=$(echo "$data" | grep -v $grep_exclude)
    fi
    local num_fields=$(echo $header | wc -w | sed 's/  *//g')

    KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:$num_fields --preview "echo -e \"${header}\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=\$i} (NR==2){for (i in a) {printf a[i] \": \" \$i \"\n\"} }' | column -t | fold -w \$COLUMNS" )
    (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | awk "$end_print"
}

# $1 is awk end print command
# $2 is filepath
# $3 is context
# $4 is query
# $5 optional namespace
_fzf_field_selector_complete()
{
    local end_print=$1
    local resource_prefix="$2"
    local header_file="${resource_prefix}_header"
    local context="$3"
    local query=$4
    local namespace="$5"
    local field_selector_field=$(_fzf_get_header_position $header_file "FieldSelectors")
    local main_header=$(_fzf_get_main_header $context $namespace)

    local header="Namespace FieldSelector Occurrences"
    local data=$(cut -d' ' -f 1,$field_selector_field $resource_prefix \
        | awk '{split($2,c,","); for (i in c){print $1,c[i]; print "all-namespaces",c[i]}}' | uniq -c | awk '{print $2,$3,$1}' | sort -k 3 -n -r)

    if [[ -n $namespace ]]; then
        data=$(echo "$data" | grep -w "^$namespace")
    fi

    (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | awk "$end_print"
}

# $1 is filepath
# $2 is context
# $3 is query
_fzf_with_namespace()
{
    local namespace_in_query=$(__get_parameter_in_query --namespace -n)
    _fzf_kubectl_complete '{print $1,$2}' "false" $1 "$2" "$3" "$namespace_in_query"
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
    local namespace_in_query=$(__get_parameter_in_query --namespace -n)
    _fzf_kubectl_complete '{print $1,$2}' "with_namespace" $1 "$2" "$3" "$namespace_in_query"
}

# $1 is filepath
# $2 is query
# $3 is context
_flag_selector_without_namespace()
{
    _fzf_kubectl_complete '{print $1}' "without_namespace" $1 "$2" "$3"
}

# $1 is filepath
# $2 is context
# $3 is query
_fzf_field_selector_with_namespace()
{
    local namespace_in_query=$(__get_parameter_in_query --namespace -n)
    _fzf_field_selector_complete '{print $1,$2}' $1 "$2" "$3" "$namespace_in_query"
}

__kubectl_get_containers()
{
    local pod=$(echo $COMP_LINE | awk '{print $(NF)}')
    local current_context=$(kubectl config current-context)
    local main_header=$(_fzf_get_main_header $current_context "")
    local data=$(awk "(\$2 == \"$pod\") {print \$7}" ${KUBECTL_FZF_CACHE}/${current_context}/pods \
        | tr ',' '\n')
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
    local i=0
    for word in ${COMP_WORDS[@]} ; do
        for parameter in $* ; do
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
        if [[ $namespace == "all-namespaces" ]]; then
            COMPREPLY=( "${result[1]} --all-namespaces" )
        elif [[ $namespace != $current_namespace && $COMP_LINE != *" -n"* && "$COMP_LINE" != *" --namespace"* ]]; then
            COMPREPLY=( "${result[1]} -n ${result[0]}" )
        else
            COMPREPLY=( ${result[1]} )
        fi
    else
        COMPREPLY=( $result )
    fi
}

__kubectl_get_resource()
{
    local current_context=$(kubectl config current-context)
    local apiresources_file="${KUBECTL_FZF_CACHE}/${current_context}/apiresources_resource"
    local header_file="${apiresources_file}_header"

    _fzf_fetch_resource $current_context $KUBECTL_FZF_RSYNC_API_RESOURCE_CACHE_TIME "apiresources"

    if [[ ! -f ${apiresources_file} ]]; then
        ___kubectl_get_resource $*
        return
    fi

    local last_part=$(echo $COMP_LINE | awk '{print $(NF)}')
    local last_char=${COMP_LINE: -1}
    local query=""

    if [[ ${#nouns[@]} -gt 0 ]]; then
        # 'k get pod <TAB>' completion
        __kubectl_parse_get "${nouns[${#nouns[@]} -1]}"
        return 0
    fi

    # 'k get p<TAB>' completion
    if [[ $last_char != " " ]]; then
        query=$last_part
    fi

    # 'k get <TAB>' completion
    local main_header=$(_fzf_get_main_header $context $namespace)

    local header=$(cat $header_file)
    local data=$(cat $apiresources_file)

    result=$( (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | cut -d' ' -f1)

    COMPREPLY=( $result )
    return 0
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
    local field_selector_autocomplete_fun
    local resource_name=$1

    case $resource_name in
        all )
            filename="pods"
            ;;
        po | pod | pods )
            filename="pods"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            field_selector_autocomplete_fun=_fzf_field_selector_with_namespace
            ;;
        sa | serviceaccount | serviceaccounts )
            filename="serviceaccounts"
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
        cronjob | cronjobs | cronjob. | cronjobs. | cronjobs.batch )
            filename="cronjobs"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        hpa | horizontalpodautoscaler | horizontalpodautoscalers | horizontalpodautoscalers.autoscaling)
            filename="horizontalpodautoscalers"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        ingress | ingresses | ingress. | ingresses. | ingresses.extensions )
            filename="ingresses"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        cm | configmap | configmaps )
            filename="configmaps"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        secret | secrets )
            filename="secrets"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        ns | namespace | namespaces )
            filename="namespaces"
            autocomplete_fun=_fzf_without_namespace
            flag_autocomplete_fun=_flag_selector_without_namespace
            ;;
        no | node | nodes )
            filename="nodes"
            autocomplete_fun=_fzf_kubectl_node_complete
            flag_autocomplete_fun=_flag_selector_without_namespace
            ;;
        deploy | deployment | deployments | deployments. | deployments.apps | deployments.extensions  )
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
            autocomplete_fun=_fzf_kubectl_pv_complete
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

    local query_context=$(__get_parameter_in_query --context)
    local context=$current_context
    if [[ -n $query_context && $query_context != $current_context ]]; then
        context=$query_context
    fi

    if [[ "$filename" == "nodes" ]]; then
        _fzf_fetch_resource $current_context $KUBECTL_FZF_RSYNC_RESOURCE_CACHE_TIME "pods" "nodes" "daemonsets"
    else
        _fzf_fetch_resource $current_context $KUBECTL_FZF_RSYNC_RESOURCE_CACHE_TIME $filename
    fi

    local filepath="${KUBECTL_FZF_CACHE}/${context}/${filename}"
    if [[ ! -f ${filepath}_header ]]; then
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
    elif [[ -n $field_selector_autocomplete_fun && ($penultimate == "--field-selector" || $last_part == "--field-selector") ]]; then
        if [[ ($penultimate == "--field-selector") && ${COMP_LINE: -1} == " " ]]; then
            return
        fi
        if [[ $penultimate == "--field-selector" ]]; then
            query=$last_part
        fi
        result=$($field_selector_autocomplete_fun $filepath $context $query)
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

__kubectl_handle_filename_extension_flag()
{
    local ext="$1"

    x="${COMP_WORDS[COMP_CWORD]}"
    if [[ "${x}" == "" ]]; then
        findNames_0=$( sed "s/|/' '*./g" <<< "'*.$ext'" )
        findNames_1=$( sed "s/|/ -o -name *./g" <<< "-name *.$ext" )
        COMPREPLY=$( ((eval "git diff HEAD --name-only --diff-filter ACMR --relative -- $findNames_0" | grep --color=always .) 2> /dev/null; (ag -g "$ext" || find . -type f $findNames_1) | sed 's|^./||g') 2> /dev/null | __kubectl_fzf_preview )
    else
        _filedir "@(${ext})"
    fi
}

__kubectl_fzf_preview()
{
    fzf --ansi --no-sort --reverse --preview '[[ $(file --mime {}) =~ binary ]] &&
        echo {} is a binary file ||
        (bat --style=numbers --color=always {} ||
        highlight -O ansi -l {} ||
        coderay {} ||
        rougify {} ||
        cat {}) 2> /dev/null | head -500'
}

# Reregister complete function without '-o default' as we don't want to
# fallback to files and dir completion
declare -f complete > /dev/null && complete -F __start_kubectl kubectl || true
