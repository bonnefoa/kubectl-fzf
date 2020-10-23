export KUBECTL_FZF_CACHE="/tmp/kubectl_fzf_cache"
eval "`builtin declare -f __kubectl_parse_get | sed '1s/.*/_&/'`"
eval "`builtin declare -f __kubectl_parse_resource | sed '1s/.*/_&/'`"
eval "`builtin declare -f __kubectl_get_containers | sed '1s/.*/_&/'`"
eval "`builtin declare -f __kubectl_get_resource | sed '1s/.*/_&/'`"
eval "`builtin declare -f __kubectl_handle_filename_extension_flag | sed '1s/.*/_&/'`"
KUBECTL_FZF_CONF=${KUBECTL_FZF_CONF:-$HOME/.kubectl_fzf.sh}
KUBECTL_FZF_EXCLUDE=${KUBECTL_FZF_EXCLUDE:-}
KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse -e --no-hscroll --no-sort)
# Cache time when no rsync service was detected
KUBECTL_FZF_RSYNC_NO_SERVICE_CACHE_TIME=${KUBECTL_FZF_RSYNC_NO_SERVICE_CACHE_TIME:-3600}
# Cache time of api resource list
KUBECTL_FZF_RSYNC_API_RESOURCE_CACHE_TIME=${KUBECTL_FZF_RSYNC_API_RESOURCE_CACHE_TIME:-3600}
# Cache time of every other resources
KUBECTL_FZF_RSYNC_RESOURCE_CACHE_TIME=${KUBECTL_FZF_RSYNC_RESOURCE_CACHE_TIME:-60}
KUBECTL_FZF_RSYNC_PORT=${KUBECTL_FZF_RSYNC_PORT:-80}
KUBECTL_FZF_PORT_FORWARD_START=${KUBECTL_FZF_PORT_FORWARD_START:-9873}
mkdir -p $KUBECTL_FZF_CACHE

_fzf_init_context_mapping()
{
    if [[ -f "$KUBECTL_FZF_CONF" ]]; then
        source "$KUBECTL_FZF_CONF"
    fi

	for cluster_group in ${cluster_groups[@]} ; do
        local clusters; clusters=(${cluster_group//,/ })
        for cluster in ${clusters[@]}; do
            context_to_cluster_group[$cluster]="$cluster_group"
        done
	done
}

# $1 is filename
# $2 is header
_fzf_get_header_position()
{
    awk "NR==1{ for(i = 1; i <= NF; i++){ if (\$i == \"$2\") {print i; } } } " $1
}

_fzf_file_mtime_older_than()
{
    local file; file=$1
    local cache_time; cache_time=$2
    if [[ ! -f "$file" ]]; then
        return 0
    fi

    local mtime; mtime=$(date -r "$file" +%s)
    local current; current=$(date +%s)
    if [[ $((current - mtime)) -gt $cache_time ]]; then
        return 0
    fi

    return 1
}

_fzf_get_filepath()
{
    local context; context="$1"
    local resource_name; resource_name="$2"
    local suffix; suffix="${3:-}"
    echo "${KUBECTL_FZF_CACHE}/${context}/${resource_name}${suffix}"
}

_fzf_get_filepaths()
{
    local contexts; contexts=(${1//,/ })
    local resource_name; resource_name="$2"
    local suffix; suffix="$3"
    local files; files=""
    for context in ${contexts[@]}; do
        files="$files ${KUBECTL_FZF_CACHE}/${context}/${resource_name}${suffix}"
    done
    echo ${files:1}
}

# $1 is contexts
# $2 is cache time
# $3 is resource name
_fzf_fetch_rsynced_resource()
{
    local context; context=$1
    local cache_time; cache_time=$2
    shift 2
    local resources; resources=($@)

    if pgrep -x cache_builder > /dev/null; then
        return
    fi

    mkdir -p "${KUBECTL_FZF_CACHE}/${context}"
    for resource in ${resources[@]} ; do
        local check_time_file; check_time_file="${KUBECTL_FZF_CACHE}/${context}/_${resource}"
        if ! $(_fzf_file_mtime_older_than $check_time_file $cache_time); then
            return
        fi
    done

    local include_param; include_param=()
    for resource_name in ${resources[@]} ; do
        include_param+="--include=${resource_name}* "
    done

    local rsync_endpoint; rsync_endpoint=($(_fzf_check_direct_access $context))
    if [[ -z "${rsync_endpoint[@]}" ]]; then
        rsync_endpoint=($(_fzf_check_port_forward $context))
    fi

    if [[ -n "$rsync_endpoint" ]]; then
        rsync -qPrz --delete ${include_param[@]} --timeout=1 --exclude="*" "rsync://${rsync_endpoint[@]:0:1}:${rsync_endpoint[@]:1:1}/fzf_cache/" "${KUBECTL_FZF_CACHE}/${context}/" 2> /dev/null

        if [[ $? -eq 0 ]]; then
            for resource in ${resources[@]} ; do
                local check_time_file; check_time_file="${KUBECTL_FZF_CACHE}/${context}/_${resource}"
                touch "$check_time_file"
            done
        fi

    fi
}

_fzf_get_port_forward_port()
{
    local context; context="$1"
    local port_file; port_file="$KUBECTL_FZF_CACHE/${context}_port"
    local global_port_file; global_port_file="$KUBECTL_FZF_CACHE/port"
    local local_port; local_port=$KUBECTL_FZF_PORT_FORWARD_START
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

_fzf_check_connection()
{
    local ip; ip="$1"
    local port; port="$2"

    local nc_options; nc_options=""
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
    local context; context="$1"
    local endpoint_file; endpoint_file="$KUBECTL_FZF_CACHE/${context}_cache_endpoint"
    if [[ -s "$endpoint_file" ]]; then
        local cached_ip; cached_ip=$(cat "$endpoint_file")
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
    for ip in $(kubectl get endpoints --context $context -l app=kubectl-fzf --all-namespaces -o=jsonpath='{.items[*].subsets[*].addresses[*].ip}'); do
        if _fzf_check_connection $ip ${KUBECTL_FZF_RSYNC_PORT}; then
            echo $ip > "$endpoint_file"
            echo "$ip $KUBECTL_FZF_RSYNC_PORT"
            return
        fi
    done
    echo "No service" > "$endpoint_file"
}

_fzf_check_port_forward_running()
{
    local local_port; local_port=$1
    if ! _fzf_check_connection localhost $local_port; then
        return 1
    fi
    return 0
}

_fzf_get_service_namespace()
{
    local context; context="$1"
    local service_file; service_file="$KUBECTL_FZF_CACHE/${context}_service"
    local kfzf_ns; kfzf_ns=""
    if [[ -s "$service_file" ]]; then
        local cached_service; cached_service=$(cat "$service_file")
        if [[ "$cached_service" == "No service" ]]; then
            if ! $(_fzf_file_mtime_older_than $service_file $KUBECTL_FZF_RSYNC_NO_SERVICE_CACHE_TIME); then
                return 1
            fi
        else
            echo "$cached_service"
            return 0
        fi
    fi
    kfzf_ns=($(kubectl get svc --context $context --all-namespaces -l app=kubectl-fzf -o=jsonpath='{.items[0].metadata.namespace}' 2> /dev/null))
    if [[ "$kfzf_ns" == "" ]]; then
        echo "No service" > "$service_file"
        return 1
    fi
    echo "${kfzf_ns}" > "$service_file"
    echo "${kfzf_ns}"
    return 0
}

# $1 is context
_fzf_check_port_forward()
{
    local context; context="$1"
    local port_file; port_file="$KUBECTL_FZF_CACHE/${context}_port"
    local log_file; log_file="$KUBECTL_FZF_CACHE/${context}_port_forward_log"

    local local_port; local_port=$(_fzf_get_port_forward_port $context)
    if _fzf_check_port_forward_running $local_port; then
        echo "localhost $local_port"
        return 0
    fi

    local kfzf_ns; kfzf_ns=$(_fzf_get_service_namespace $context)
    if [[ "$kfzf_ns" == "" ]]; then
        return 1
    fi

    (nohup kubectl port-forward svc/kubectl-fzf --context $context -n ${kfzf_ns} ${local_port}:${KUBECTL_FZF_RSYNC_PORT} &> $log_file &)

    for (( i = 0; i < 10; i++ )); do
        if _fzf_check_connection localhost $local_port; then
            echo "localhost $local_port"
            return 0
        fi
        sleep 1
    done

    if _fzf_check_port_forward_running $local_port; then
        echo "localhost $local_port"
        return 0
    fi
    return 1
}

# $1 is context
# $2 is namespace
_fzf_get_main_header()
{
    local active_context; active_context="$1"
    local contexts; contexts="$2"
    local namespace; namespace="${3:-}"
    local main_header; main_header="Active Context:$active_context"
    if [[ $active_context != $contexts ]]; then
        main_header="$main_header, Search contexts:$contexts"
    fi
    if [[ -n $namespace ]]; then
        main_header="$main_header, Namespace:$namespace"
    fi
    echo $main_header
}

_fzf_get_exclude_pattern()
{
    local grep_exclude; grep_exclude=""
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
    local context; context="$1"
    local pod_file; pod_file=$(_fzf_get_filepath $current_context "pods" "_resource")
    local pod_header_file; pod_header_file=$(_fzf_get_filepath $current_context "pods" "_header")
    local node_name_field; node_name_field=$(_fzf_get_header_position $pod_header_file "NodeName")
    local pod_name_field; pod_name_field=$(_fzf_get_header_position $pod_header_file "Name")

    local daemonsets_file; daemonsets_file=$(_fzf_get_filepath $current_context "daemonsets" "")
    local daemonsets; daemonsets=""
    if [[ -f $daemonsets_file ]]; then
        daemonsets=$(cut -d' ' -f2 "$daemonsets_file" | uniq)
    fi
    local exclude_pods; exclude_pods=""
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
    local resource_name; resource_name="$1"
    local contexts; contexts="$2"
    local query; query="$3"
    local namespace; namespace=$(__get_parameter_in_query --namespace -n)

    local current_context; current_context=$(kubectl config current-context)
    local main_header; main_header=$(_fzf_get_main_header $current_context $contexts $namespace)

    local pv_header_file; pv_header_file=$(_fzf_get_filepath $current_context $resource_name "_header")
    local pv_resource_files; pv_resource_files=($(_fzf_get_filepaths $contexts $resource_name "_resource"))
    local pv_label_files; pv_label_files=($(_fzf_get_filepaths $contexts $resource_name "_label"))

    local label_field; label_field=$(_fzf_get_header_position $pv_header_file "Labels")
    local claim_field_pv_file; claim_field_pv_file=$(_fzf_get_header_position $pv_header_file "Claim")
    local end_field; end_field=$((label_field - 1))
    local header; header=$(cut -d ' ' -f 1-$end_field "$pv_header_file")
    header="$header MountedBy"

    local pod_files; pod_files=($(_fzf_get_filepaths $contexts "pods" "_resource"))
    local pod_header_file; pod_header_file=$(_fzf_get_filepath $current_context "pods" "_header")

    local claim_field_pod_file; claim_field_pod_file=$(_fzf_get_header_position $pod_header_file "Claims")
    local pod_name_field; pod_name_field=$(_fzf_get_header_position $pod_header_file "Name")
    local pod_namespace_field; pod_namespace_field=$(_fzf_get_header_position $pod_header_file "Namespace")

    local data; data=$(join -a1 -o'1.1,1.2,1.3,1.4,1.5,1.6,1.7,1.8,2.2' -1 $claim_field_pv_file -2 1 -e None \
        <(cut -d ' ' -f 1-$end_field "$pv_resource_files" | sort -k $claim_field_pv_file) \
        <(awk "(\$$claim_field_pod_file != \"None\"){split(\$$claim_field_pod_file,c,\",\"); for (i in c) { print c[i] \" \" \$$pod_namespace_field\"/\"\$$pod_name_field } }" $pod_files | sort))
    local num_fields; num_fields=$(echo $header | wc -w | sed 's/  *//g')

    KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:$num_fields --preview "echo -e \"${header}\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=\$i} (NR==2){for (i in a) {printf a[i] \": \" \$i \"\n\"} }' | column -t | fold -w \$COLUMNS" )
    (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | awk '{print $2}'
}

_fzf_kubectl_node_complete()
{
    local resource_name; resource_name="$1"
    local contexts; contexts="$2"
    local query; query="$3"

    local current_context; current_context=$(kubectl config current-context)
    local main_header; main_header=$(_fzf_get_main_header $current_context $contexts $namespace)
    local node_resource_file; node_resource_file=$(_fzf_get_filepaths $contexts $resource_name "_resource")
    local node_header_file; node_header_file=$(_fzf_get_filepaths $contexts $resource_name "_header")

    local label_field; label_field=$(_fzf_get_header_position $node_header_file "Labels")
    local end_field; end_field=$((label_field - 1))
    local header; header=$(cut -d ' ' -f 1-$end_field "$node_header_file")
    header="$header Pods"

    local node_to_pods; node_to_pods=$(_fzf_get_node_to_pods $context)
    local data; data=$(join -a1 -oauto -e None <(cut -d ' ' -f 1-$end_field "$node_resource_file") <(echo "$node_to_pods"))
    local num_fields; num_fields=$(echo $header | wc -w | sed 's/  *//g')
    KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:$num_fields --preview "echo -e \"${header}\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=\$i} (NR==2){for (i in a) {printf a[i] \": \" \$i \"\n\"} }' | column -t | fold -w \$COLUMNS" )
    (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | awk '{print $1,$2}'
}

# $1 is awk end print command
# $2 isFlag
# $3 is resource_name
# $4 is contexts
# $5 is query
# $6 optional namespace
_fzf_kubectl_complete()
{
    local end_print; end_print=$1
    local is_flag; is_flag="$2"
    local resource_name; resource_name="$3"
    local contexts; contexts=$4
    local query; query=$5
    local namespace; namespace="$6"

    local current_context; current_context=$(kubectl config current-context)
    local header_file; header_file=$(_fzf_get_filepath $current_context $resource_name "_header")
    local resource_files; resource_files=($(_fzf_get_filepaths $contexts $resource_name "_resource"))
    local label_files; label_files=($(_fzf_get_filepaths $contexts $resource_name "_label"))

    local label_field; label_field=$(_fzf_get_header_position $header_file "Labels")
    local end_field; end_field=$((label_field - 1))
    local main_header; main_header=$(_fzf_get_main_header $current_context $current_context $namespace)

    if [[ $is_flag == "with_namespace" ]]; then
        local header; header="Cluster Namespace Labels Occurrences"
        local data; data=$(cat ${label_files[@]})
    elif [[ $is_flag == "without_namespace" ]]; then
        local header; header="Cluster Labels Occurrences"
        local data; data=$(cat ${label_files[@]})
    else
        local header; header=$(cut -d ' ' -f 1-$end_field "$header_file")
        local data; data=$(cut -d ' ' -f 1-$end_field ${resource_files[@]})
    fi

    if [[ -n $namespace ]]; then
        data=$(echo "$data" | grep -w "^$namespace")
    fi

    local grep_exclude; grep_exclude=$(_fzf_get_exclude_pattern)
    if [[ -n $grep_exclude ]]; then
        data=$(echo "$data" | grep -v $grep_exclude)
    fi
    local num_fields; num_fields=$(echo $header | wc -w | sed 's/  *//g')

    KUBECTL_FZF_PREVIEW_OPTIONS=(--preview-window=down:$num_fields --preview "echo -e \"${header}\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=\$i} (NR==2){for (i in a) {printf a[i] \": \" \$i \"\n\"} }' | column -t | fold -w \$COLUMNS" )
    (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf "${KUBECTL_FZF_PREVIEW_OPTIONS[@]}" ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | awk "$end_print"
}

# $1 is awk end print command
# $2 is resource_name
# $3 is contexts
# $4 is query
# $5 optional namespace
_fzf_field_selector_complete()
{
    local end_print; end_print=$1
    local resource_prefix; resource_prefix="$2"
    local contexts; contexts="$3"
    local query; query=$4
    local namespace; namespace="$5"

    local current_context; current_context=$(kubectl config current-context)
    local header_file; header_file=$(_fzf_get_filepath $current_context $resource_name "_header")
    local resource_files; resource_files=($(_fzf_get_filepaths $contexts $resource_name "_resource"))

    local field_selector_field; field_selector_field=$(_fzf_get_header_position $header_file "FieldSelectors")
    local main_header; main_header=$(_fzf_get_main_header $current_context $contexts $namespace)

    local header; header="Cluster Namespace FieldSelector Occurrences"
    local data; data=$(cut -d' ' -f 1,2,$field_selector_field ${resource_files[@]} \
        | awk '{split($3,c,","); for (i in c){print $1,$2,c[i]; print $1,"all-namespaces",c[i]}}' | sort | uniq -c | awk '{print $2,$3,$4,$1}' | sort -k 4 -n -r)

    if [[ -n $namespace ]]; then
        data=$(echo "$data" | grep -w "^$namespace")
    fi

    (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | awk "$end_print"
}

# $1 is resource_name
# $2 is contexts
# $3 is query
_fzf_with_namespace()
{
    local namespace_in_query; namespace_in_query=$(__get_parameter_in_query --namespace -n)
    _fzf_kubectl_complete '{print $1,$2,$3}' "false" $1 "$2" "${3:-}" "$namespace_in_query"
}

# $1 is resource_name
# $2 is contexts
# $3 is query
_fzf_without_namespace()
{
    _fzf_kubectl_complete '{print $1,$2}' "false" $1 "$2" "$3"
}

# $1 is resource_name
# $2 is contexts
# $3 is query
_flag_selector_with_namespace()
{
    local namespace_in_query; namespace_in_query=$(__get_parameter_in_query --namespace -n)
    _fzf_kubectl_complete '{print $1,$2,$3}' "with_namespace" $1 "$2" "$3" "$namespace_in_query"
}

# $1 is resource_name
# $2 is query
# $3 is contexts
_flag_selector_without_namespace()
{
    _fzf_kubectl_complete '{print $1,$2}' "without_namespace" $1 "$2" "$3"
}

# $1 is resource_name
# $2 is contexts
# $3 is query
_fzf_field_selector_with_namespace()
{
    local namespace_in_query; namespace_in_query=$(__get_parameter_in_query --namespace -n)
    _fzf_field_selector_complete '{print $1,$2,$3}' $1 "$2" "$3" "$namespace_in_query"
}

__kubectl_get_containers()
{
    local pod; pod=$(echo $COMP_LINE | awk '{print $(NF)}')
    local current_context; current_context=$(kubectl config current-context)
    local main_header; main_header=$(_fzf_get_main_header $current_context)
    local data; data=$(awk "(\$2 == \"$pod\") {print \$7}" ${KUBECTL_FZF_CACHE}/${current_context}/pods \
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
    local context; context=$1
    local namespace; namespace=$(kubectl config --context $context view --minify --output 'jsonpath={..namespace}')
    echo "${namespace:-default}"
}

__get_parameter_in_query()
{
    local i; i=0
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
    local result; result=("$@")
    result=($(echo $result | tr " " "\n"))
    local context; context=${result[0]}
    local current_context; current_context=$(kubectl config current-context)

    context_option=""
    if [[ $context != $current_context && $COMP_LINE != *" --context"* ]]; then
        context_option="--context $context"
    fi

    if [[ ${#result[@]} -eq 3 ]]; then
        # We have namespace in first position
        local namespace; namespace=${result[1]}
        local current_namespace; current_namespace=$(__get_current_namespace $context)

        local namespace_option; namespace_option=""
        if [[ $namespace == "all-namespaces" ]]; then
            namespace_option="--all-namespaces"
        elif [[ $namespace != $current_namespace && $COMP_LINE != *" -n"* && "$COMP_LINE" != *" --namespace"* ]]; then
            namespace_option="-n ${result[1]}"
        fi

        local cluster_option; cluster_option=""
        if [[ $context != $current_context && $COMP_LINE != *" --context"* ]]; then
            # Always put the namespace if the context is different
            if [[ $namespace == "all-namespaces" ]]; then
                namespace_option="--all-namespaces"
            elif [[ $namespace != $current_namespace && $COMP_LINE != *" -n"* && "$COMP_LINE" != *" --namespace"* ]]; then
                namespace_option="-n ${result[1]}"
            fi
        fi
        COMPREPLY=( "${result[2]} $context_option $namespace_option" )
    else
        COMPREPLY=( ${result[1]} $context_option )
    fi
}

__kubectl_get_resource()
{
    #set -ue

    local current_context; current_context=$(kubectl config current-context)
    local header_file; header_file=$(_fzf_get_filepath $current_context "apiresources" "_header")
    local apiresources_file; apiresources_file=$(_fzf_get_filepath $current_context "apiresources" "_resource")

    _fzf_fetch_rsynced_resource $current_context $KUBECTL_FZF_RSYNC_API_RESOURCE_CACHE_TIME "apiresources"
    if [[ ! -f "${apiresources_file}" ]]; then
        ___kubectl_get_resource $*
        return
    fi

    local last_part; last_part=$(echo $COMP_LINE | awk '{print $(NF)}')
    local last_char; last_char=${COMP_LINE: -1}
    local query; query=""

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
    local main_header; main_header=$(_fzf_get_main_header $current_context $current_context $namespace)

    local header; header=$(cat $header_file)
    local data; data=$(cat $apiresources_file)

    result=$( (printf "${main_header}\n"; printf "${header}\n${data}\n" | column -t) \
        | fzf ${KUBECTL_FZF_OPTIONS[@]} -q "$query" \
        | cut -d' ' -f1)

    COMPREPLY=( $result )
    return 0
}

# $1 is the type of resource to get
__kubectl_parse_get()
{
    local penultimate; penultimate=$(echo $COMP_LINE | awk '{print $(NF-1)}')
    local last_part; last_part=$(echo $COMP_LINE | awk '{print $(NF)}')
    local current_context; current_context=$(kubectl config current-context)

    local resource_name
    local autocomplete_fun
    local flag_autocomplete_fun
    local field_selector_autocomplete_fun
    local resource_arg; resource_arg=$1

    case $resource_arg in
        all )
            resource_name="pods"
            ;;
        po | pod | pods )
            resource_name="pods"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            field_selector_autocomplete_fun=_fzf_field_selector_with_namespace
            ;;
        sa | serviceaccount | serviceaccounts )
            resource_name="serviceaccounts"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        ds | daemonset | daemonsets | daemonsets.apps | daemonsets.extensions | daemonsets. )
            resource_name="daemonsets"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        rs | resplicaset | replicasets )
            resource_name="replicasets"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        cronjob | cronjobs | cronjob. | cronjobs. | cronjobs.batch )
            resource_name="cronjobs"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        hpa | horizontalpodautoscaler | horizontalpodautoscalers | horizontalpodautoscalers.autoscaling)
            resource_name="horizontalpodautoscalers"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        ingress | ingresses | ingress. | ingresses. | ingresses.extensions )
            resource_name="ingresses"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        cm | configmap | configmaps )
            resource_name="configmaps"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        secret | secrets )
            resource_name="secrets"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        ns | namespace | namespaces )
            resource_name="namespaces"
            autocomplete_fun=_fzf_without_namespace
            flag_autocomplete_fun=_flag_selector_without_namespace
            ;;
        no | node | nodes )
            resource_name="nodes"
            autocomplete_fun=_fzf_kubectl_node_complete
            flag_autocomplete_fun=_flag_selector_without_namespace
            ;;
        deploy | deployment | deployments | deployments. | deployments.apps | deployments.extensions  )
            resource_name="deployments"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        sts | statefulset | statefulsets | statefulsets.apps  )
            resource_name="statefulsets"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        persistentvolumes | pv )
            resource_name="persistentvolumes"
            autocomplete_fun=_fzf_kubectl_pv_complete
            flag_autocomplete_fun=_flag_selector_without_namespace
            ;;
        persistentvolumeclaims | pvc )
            resource_name="persistentvolumeclaims"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        endpoints )
            resource_name="endpoints"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        svc | service | services )
            resource_name="services"
            autocomplete_fun=_fzf_with_namespace
            flag_autocomplete_fun=_flag_selector_with_namespace
            ;;
        * )
            ___kubectl_parse_get $*
            return
            ;;
    esac

    builtin declare -A cluster_groups
    builtin declare -A context_to_cluster_group
    _fzf_init_context_mapping

    local contexts; contexts=($current_context)
    local cluster_group
    if [[ -v "context_to_cluster_group[$current_context]" ]] ; then
        cluster_group=${context_to_cluster_group[$current_context]}
    fi
    local context_array; context_array=($current_context)
    if [[ -n $cluster_group ]]; then
        context_array=(${cluster_group//,/ })
        contexts=$cluster_group
    fi

    for context in ${context_array[@]}; do
        if [[ "$resource_name" == "nodes" ]]; then
            _fzf_fetch_rsynced_resource $context $KUBECTL_FZF_RSYNC_RESOURCE_CACHE_TIME "pods" "nodes" "daemonsets" &
        else
            _fzf_fetch_rsynced_resource $context $KUBECTL_FZF_RSYNC_RESOURCE_CACHE_TIME $resource_name &
        fi
    done
    wait

    local header_path; header_path=$(_fzf_get_filepath ${current_context} $resource_name "_header")
    if [[ ! -f ${header_path} ]]; then
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
        result=$($flag_autocomplete_fun $resource_name $contexts $query)
        __build_namespaced_compreply "${result[@]}"
        return
    elif [[ -n $field_selector_autocomplete_fun && ($penultimate == "--field-selector" || $last_part == "--field-selector") ]]; then
        if [[ ($penultimate == "--field-selector") && ${COMP_LINE: -1} == " " ]]; then
            return
        fi
        if [[ $penultimate == "--field-selector" ]]; then
            query=$last_part
        fi
        result=$($field_selector_autocomplete_fun $resource_name $contexts $query)
        __build_namespaced_compreply "${result[@]}"
        return
    fi

    if [[ -z $autocomplete_fun ]]; then
        ___kubectl_parse_get $*
        return
    fi

    local query; query=""
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

    result=$($autocomplete_fun $resource_name $contexts $query)
    if [[ -z "$result" ]]; then
        return
    fi

    __build_namespaced_compreply "${result[@]}"
}

__kubectl_handle_filename_extension_flag()
{
    local ext; ext="$1"

    x="${COMP_WORDS[COMP_CWORD]}"
    if [[ "${x}" == "" ]]; then
        findNames_0=$( sed "s/|/' '*./g" <<< "'*.$ext'" )
        findNames_1=$( sed "s/|/ -o -name *./g" <<< "-name *.$ext" )
        COMPREPLY=$( ((eval "git diff HEAD --name-only --diff-filter ACMR --relative -- $findNames_0" | grep --color=always .) 2> /dev/null; (ag -g "$ext" || find . -type f $findNames_1) | sed 's|^./||g') 2> /dev/null | __kubectl_fzf_preview )
    else
        __kubectl_filedir "@(${ext})"
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
builtin declare -f complete > /dev/null && complete -F __start_kubectl kubectl || true
