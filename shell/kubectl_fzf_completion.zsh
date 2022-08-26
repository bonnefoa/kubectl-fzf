KUBECTL_FZF_COMPLETION_BIN=${KUBECTL_FZF_COMPLETION_BIN:-kubectl-fzf-completion}
. "${0:A:h}/kubectl_fzf.sh"

__kubectl_fzf_kubectl()
{
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

    query=$(__kubectl_fzf_query_from_word "$currentWord")

    cmdArgs="${words[2,-1]}"
    completionOutput=$(__kubectl_fzf_get_completions "$cmdArgs" "$lastChar" "$query")
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
kubectl_fzf_completion()
{
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
      for word in "${words[@]:1}"
      do
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
      IFS=" " read -r -A kubectl_fzf_default_completion <<< "$binding"
      kubectl_fzf_default_completion=${kubectl_fzf_default_completion[2]}
  fi
  unset binding
fi

zle     -N   kubectl_fzf_completion
bindkey '^I' kubectl_fzf_completion
