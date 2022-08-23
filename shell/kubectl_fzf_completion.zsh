KUBECTL_FZF_OPTIONS=(-1 --header-lines=1 --layout reverse -e --no-hscroll --no-sort --cycle)

__kubectl_fzf_debug()
{
    local file="$KUBECTL_FZF_COMP_DEBUG_FILE"
    if [[ -n ${file} ]]; then
        echo "$*" >> "${file}"
    fi
}

_kubectl_fzf_call_fzf()
{
    local fzfPreviewOptions numFields completionOutput query
    completionOutput="$1"
    query="$2"

    header=$(echo "$completionOutput" | head -n1)
    numFields=$(echo "$header" | wc -w | sed 's/  *//g')

    fzfPreviewOptions=(--preview-window=down:"$numFields" --preview "echo -e \"${header}\n{}\" | sed -e \"s/'//g\" | awk '(NR==1){for (i=1; i<=NF; i++) a[i]=\$i} (NR==2){for (i in a) {printf a[i] \": \" \$i \"\n\"} }' | column -t | fold -w \$COLUMNS" )

    (printf "%s" "$completionOutput") \
        | fzf "${fzfPreviewOptions[@]}" "${KUBECTL_FZF_OPTIONS[@]}" -q "$query" \
        | awk '{print $1,$2,$3}'
}

_kubectl_fzf_get_completions()
{
    local completionArgs completionOutput requestComp lastChar
    completionArgs="$1"
    lastChar="$2"
    # TODO: Check existence and pull it from $PATH
    requestComp="./kubectl-fzf-completion k8s_completion $completionArgs"
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

_kubectl_fzf_process_result()
{
    local processCommand outResult completionArgs fzfResults
    completionArgs="$1"
    fzfResults="$2"

    processCommand="./kubectl-fzf-completion process_result --source-cmd \"${completionArgs}\" --fzf-result \"$fzfResults\""
    __kubectl_fzf_debug "About to call: eval ${processCommand}"
    outResult=$(eval "${processCommand}")
    exitCode=$?
    __kubectl_fzf_debug "processed result code $exitCode: ${outResult}"

    if [[ $exitCode != 0 ]]; then
        echo "Error when calling process result. Command: $processCommand"
        echo "Output: $outResult"
        return
    else
        __kubectl_fzf_debug "Adding to the LBUFFER: '$outResult '"
        if [[ ${LBUFFER[-1]} != " " ]]; then
            zle backward-kill-word
        fi
        LBUFFER+="$outResult "
    fi
}

_kubectl_fzf_kubectl()
{
    local currentWord previousWord lastChar
    local requestComp completionArgs
    local fzfResults completionOutput

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
    # Last word is an unmanaged flag
    if [[ $currentWord == -* && $currentWord != -i && $currentWord != -t && $currentWord != --selector* && $currentWord != -l* && $currentWord != --field-selector* ]]; then
        zle "${kubectl_fzf_default_completion:-expand-or-complete}"
        return
    fi
    # Previous word is an unmanaged flag expecting a value
    if [[ $previousWord == -* && $previousWord != -i && $previousWord != -t && $previousWord != --all-namespaces && $previousWord != -l && $previousWord != --selector && $previousWord != --field-selector ]]; then
        zle "${kubectl_fzf_default_completion:-expand-or-complete}"
        return
    fi

    completionArgs="${words[2,-1]}"
    completionOutput=$(_kubectl_fzf_get_completions "$completionArgs" "$lastChar")
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

    if [[ $currentWord != " " ]]; then
        query="$currentWord"
        query=${query#-l}
        query=${query#--field-selector}
        query=${query#=}
    fi

    IFS=" " read -r -A fzfResults <<< "$(_kubectl_fzf_call_fzf "$completionOutput" "$query")"
    if [[ "$fzfResults" == "" ]]; then
        echo "Completion cancelled"
        return
    fi

    _kubectl_fzf_process_result "$completionArgs" "$fzfResults"
}


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

  #query=${words[${#words[@]}]}
  #if [[ $query == -* ]]; then
  #zle ${kubectl_fzf_default_completion:-expand-or-complete}
  #return
  #fi

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
  _kubectl_fzf_kubectl
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
