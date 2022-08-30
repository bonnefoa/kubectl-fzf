KUBECTL_FZF_COMPLETION_BIN=${KUBECTL_FZF_COMPLETION_BIN:-kubectl-fzf-completion}

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
    currentWord="$3"

    __kubectl_fzf_debug "Get completions: cmdArgs: '$cmdArgs', lastChar: '$lastChar', currentWord: '$currentWord'"
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

__kubectl_fzf_kubectl() {
    local currentWord previousWord lastChar
    local cmdArgs
    local completionOutput

    __kubectl_fzf_debug "CURRENT: ${CURRENT}, words[*]: '${words[*]}', ${#words[@]}"
    words=("${=words[1,CURRENT]}")
    __kubectl_fzf_debug "Truncated words[*]: ${words[*]},"
    currentWord=${words[CURRENT]}
    previousWord=${words[CURRENT-1]}
    lastChar=${words[-1][-1]}
    __kubectl_fzf_debug "Current word: ${currentWord}, previous word: ${previousWord}, lastChar: '${lastChar}'"

    # We only have 'kubectl g#', fallback to default completion
    if [[ ${#words[@]} -le 2 ]]; then
        zle "${kubectl_fzf_default_completion:-expand-or-complete}"
        return
    fi

    cmdArgs="${words[2, -1]}"
    completionOutput=$(__kubectl_fzf_get_completions "$cmdArgs" "$lastChar" "$currentWord")
    if [[ "$completionOutput" == "" ]]; then
        return
    fi
    if [[ "$completionOutput" == "fallback" ]]; then
        zle "${kubectl_fzf_default_completion:-expand-or-complete}"
        return
    fi
    if [[ "$completionOutput" == error* ]]; then
        zle "${kubectl_fzf_default_completion:-expand-or-complete}"
        return
    fi

    __kubectl_fzf_debug "Adding to the LBUFFER: '$completionOutput '"
    if [[ ${LBUFFER[-1]} != " " ]]; then
        zle backward-kill-word
    fi
    LBUFFER+="$completionOutput "
}

# Completion entry point
kubectl_fzf_completion() {
    local words firstWord
    setopt localoptions noshwordsplit noksh_arrays noposixbuiltins
    words=(${(z)LBUFFER})
    __kubectl_fzf_debug "\n========= starting completion logic =========="
    __kubectl_fzf_debug "LBUFFER: '$LBUFFER', words: '${words[*]}', ${#words}"

    firstWord=${words[1]}

    if [[ ${#words[@]} -le 1 && ${LBUFFER[-1]} != " " ]]; then
        zle "${kubectl_fzf_default_completion:-expand-or-complete}"
        return
    fi

    # We only care about kubectl completion
    if [[ $firstWord != k* ]]; then
        zle "${kubectl_fzf_default_completion:-expand-or-complete}"
        return
    fi

    if [[ $RBUFFER != "" ]]; then
        # TODO Handle right buffer
        zle "${kubectl_fzf_default_completion:-expand-or-complete}"
        return
    fi

    if [[ "$firstWord" != "kubectl" ]]; then
        # Try to resolve alias
        expanded=(${(z)aliases[$firstWord]})
        if [ ${#expanded} -lt 1 ]; then
            zle "${kubectl_fzf_default_completion:-expand-or-complete}"
            return
        fi
        if [ "${expanded[1]}" != "kubectl" ]; then
            zle "${kubectl_fzf_default_completion:-expand-or-complete}"
            return
        fi
        # We have resolved a kubectl alias
        for word in "${words[@]:1}"; do
            expanded+=("$word")
        done
        words=("${expanded[@]}")
    fi
    if [[ ${LBUFFER[-1]} == " " ]]; then
        words+=(" ")
    fi
    CURRENT=${#words[@]}
    __kubectl_fzf_kubectl
}

if [[ -z "$kubectl_fzf_default_completion" ]]; then
    binding=$(bindkey '^I')
    if [[ $binding =~ 'undefined-key' ]]; then
        IFS=" " read -r -A kubectl_fzf_default_completion <<<"$binding"
        kubectl_fzf_default_completion=${kubectl_fzf_default_completion[2]}
    fi
    unset binding
fi

zle -N kubectl_fzf_completion
bindkey '^I' kubectl_fzf_completion
