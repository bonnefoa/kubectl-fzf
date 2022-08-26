#!/bin/bash

__kubectl_fzf_debug()
{
    local file="$KUBECTL_FZF_COMP_DEBUG_FILE"
    if [[ -n ${file} ]]; then
        echo "$*" >> "${file}"
    fi
}

__kubectl_fzf_query_from_word()
{
    local currentWord query
    if [[ $currentWord != " " ]]; then
        query="$currentWord"
        query=${query#-l}
        query=${query#--field-selector}
        query=${query#=}
    fi
    echo "$query"
}

__kubectl_fzf_get_completions()
{
    local cmdArgs completionOutput requestComp lastChar
    cmdArgs="$1"
    lastChar="$2"
    # TODO Handle query
    query="$3"

    requestComp="$KUBECTL_FZF_COMPLETION_BIN k8s_completion $cmdArgs"
    if [ "${lastChar}" = "" ]; then
        __kubectl_fzf_debug "Adding extra empty parameter"
        requestComp="${requestComp} \"\""
    fi
    __kubectl_fzf_debug "About to call: eval '${requestComp}'"
    completionOutput=$(eval "${requestComp}")
    exitCode=$?
    __kubectl_fzf_debug "completion output: ${completionOutput}, exit code ${exitCode}"

    if [[ $exitCode == 5 ]]; then
        # No completion available
        echo "error: No completion available: $requestComp"
        return
    fi
    if [[ $exitCode == 6 ]]; then
        # Unknow resource type, fallback to default completion
        echo "fallback"
        return
    fi
    if [[ $exitCode != 0 ]]; then
        # Error on completion
        echo "error when calling kubectl-fzf-completion: $requestComp. Output: $completionOutput"
        return
    fi
    echo "$completionOutput"
}
