#!/bin/bash
. $(dirname "${BASH_SOURCE[0]}")/kubectl_fzf.sh

__kubectl_fzf_get_completion_results() {
    local lastParam lastChar cmdArgs

    # Prepare the command to request completions for the program.
    # Calling ${words[0]} instead of directly kubectl allows to handle aliases
    cmdArgs="${words[*]:1}"

    lastParam=${words[$((${#words[@]}-1))]}
    lastChar=${lastParam:$((${#lastParam}-1)):1}
    __kubectl_fzf_debug "lastParam ${lastParam}, lastChar ${lastChar}"

    # When completing a flag with an = (e.g., kubectl -n=<TAB>)
    # bash focuses on the part after the =, so we need to remove
    # the flag part from $cur
    if [[ "${cur}" == -*=* ]]; then
        cur="${cur#*=}"
    fi

    completionOutput=$(__kubectl_fzf_get_completions "$cmdArgs" "$lastChar" "$cur")
    COMPREPLY=("$completionOutput")
}

__kubectl_fzf_kubectl()
{
    local cur words cword

    COMPREPLY=()

    # Call _init_completion from the bash-completion package
    # to prepare the arguments properly
    if declare -F _init_completion >/dev/null 2>&1; then
        _init_completion -n "=:" || return
    else
        __kubectl_init_completion -n "=:" || return
    fi

    __kubectl_fzf_debug
    __kubectl_fzf_debug "========= starting completion logic =========="
    __kubectl_fzf_debug "cur is ${cur}, words[*] is ${words[*]}, #words[@] is ${#words[@]}, cword is $cword"

    # The user could have moved the cursor backwards on the command-line.
    # We need to trigger completion from the $cword location, so we need
    # to truncate the command-line ($words) up to the $cword location.
    words=("${words[@]:0:$cword+1}")
    __kubectl_fzf_debug "Truncated words[*]: ${words[*]},"

    __kubectl_fzf_get_completion_results
}

if [[ $(type -t compopt) = "builtin" ]]; then
    complete -o default -F __kubectl_fzf_kubectl kubectl
else
    complete -o default -o nospace -F __kubectl_fzf_kubectl kubectl
fi

# ex: ts=4 sw=4 et filetype=sh
