#!/bin/bash

sed \
        -e 's/declare -F/whence -w/' \
        -e 's/_get_comp_words_by_ref "\$@"/_get_comp_words_by_ref "\$*"/' \
        -e 's/local \([a-zA-Z0-9_]*\)=/local \1; \1=/' \
        -e 's/flags+=("\(--.*\)=")/flags+=("\1"); two_word_flags+=("\1")/' \
        -e 's/must_have_one_flag+=("\(--.*\)=")/must_have_one_flag+=("\1")/' \
        -e "s/${LWORD}_filedir${RWORD}/__kubectl_filedir/g" \
        -e "s/${LWORD}_get_comp_words_by_ref${RWORD}/__kubectl_get_comp_words_by_ref/g" \
        -e "s/${LWORD}__ltrim_colon_completions${RWORD}/__kubectl_ltrim_colon_completions/g" \
        -e "s/${LWORD}compgen${RWORD}/__kubectl_compgen/g" \
        -e "s/${LWORD}compopt${RWORD}/__kubectl_compopt/g" \
        -e "s/${LWORD}declare${RWORD}/builtin declare/g" \
        -e "s/\\\$(type${RWORD}/\$(__kubectl_type/g" kubectl_fzf.sh > kubectl_fzf.plugin.zsh