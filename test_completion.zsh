#!/usr/bin/env zsh

. kubectl_completion.zsh
BASH_COMP_DEBUG_FILE="/tmp/log"
rm $BASH_COMP_DEBUG_FILE

words=[]
_kubectl
