#shellcheck shell=bash
compdef _kfzf_kubectl kubectl

KUBECTL_FZF_OPTIONS=(-1 --header-lines=2 --layout reverse -e --no-hscroll --no-sort)

__kubectl_debug()
{
    local file="$BASH_COMP_DEBUG_FILE"
    if [[ -n ${file} ]]; then
        echo "$*" >> "${file}"
    fi
}

# $1 is context
# $2 is namespace
_kfzf_get_main_header()
{
    local context namespace mainHeader
    context="$1"
    namespace="${2:-}"
    mainHeader="Context:$context"
    if [[ -n $namespace ]]; then
        mainHeader="$mainHeader, Namespace:$namespace"
    fi
    echo "$mainHeader"
}

_kfzf_get_current_namespace()
{
    local context namespace
    context=$1
    namespace=$(kubectl config --context "$context" view --minify --output 'jsonpath={..namespace}')
    echo "${namespace:-default}"
}

_kfzf_call_fzf()
{
    local mainHeader fzfPreviewOptions numFields
    local context namespace
    context=$1
    namespace=$2

    mainHeader=$(_kfzf_get_main_header "$context" "$namespace")

    header=$(echo "$out" | head -n1)
    numFields=$(echo "$header" | wc -w | sed 's/  *//g')

    fzfPreviewOptions=(--preview-window=down:"$numFields" --preview "echo -e \"${header}\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=\$i} (NR==2){for (i in a) {printf a[i] \": \" \$i \"\n\"} }' | column -t | fold -w \$COLUMNS" )

    (printf "%s\n%s" "$mainHeader" "$out" | column -t) \
        | fzf "${fzfPreviewOptions[@]}" "${KUBECTL_FZF_OPTIONS[@]}" \
        | awk '{print $1,$2,$3}'
}

_kfzf_kubectl()
{
    local lastParam lastChar flagPrefix requestComp out
    local context namespace results

    __kubectl_debug "\n========= starting completion logic =========="
    __kubectl_debug "CURRENT: ${CURRENT}, words[*]: ${words[*]}"

    words=("${=words[1,CURRENT]}")
    __kubectl_debug "Truncated words[*]: ${words[*]},"

    currentWord=${words[CURRENT]}
    previousWord=${words[CURRENT-1]}
    __kubectl_debug "Current word: ${currentWord}, Previous word: ${previousWord}"

    if [[ $currentWord == -* ]]; then
        _kubectl
        return
    fi

    if [[ $previousWord == -* ]]; then
        _kubectl
        return
    fi


    lastParam=${words[-1]}
    lastChar=${lastParam[-1]}
    __kubectl_debug "lastParam: ${lastParam}, lastChar: ${lastChar}"

    requestComp="./kubectl_completion ${words[2,-1]}"
    if [ "${lastChar}" = "" ]; then
        __kubectl_debug "Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi
    __kubectl_debug "About to call: eval ${requestComp}"
    out=$(eval "${requestComp}" 2>/dev/null)

    __kubectl_debug "completion output: ${out}"
    __kubectl_debug "completions: ${out}"
    __kubectl_debug "flagPrefix: ${flagPrefix}"

    context=$(kubectl config current-context)
    namespace=$(_kfzf_get_current_namespace "$context")

    IFS=" " read -r -A results <<< "$(_kfzf_call_fzf "$context" "$namespace")"

    resultContext="${results[*]:0:1}"
    resultNamespace="${results[*]:1:1}"
    resultName="${results[*]:2:1}"

    __kubectl_debug "completion results - Context: ${resultContext}"
    __kubectl_debug "completion results - Namespace: ${resultNamespace}"
    __kubectl_debug "completion results - Name: ${resultName}"

    compadd -Q -- "--namespace ${resultNamespace} $resultName"
}
