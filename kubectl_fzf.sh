# Copyright 2018 Anthonin Bonnefoy
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

export KUBECTL_FZF_CACHE="/tmp/kubectl_fzf_cache"
eval "`declare -f __kubectl_parse_get | sed '1s/.*/_&/'`"

_pod_selector()
{
	res=$(awk '{print $1 " " $2 " " $4 " " $5 " " $6 }' ${KUBECTL_FZF_CACHE}/pods \
		| column -t \
		| fzf -m --header="Namespace Name IP Node Status" --layout reverse -q "$2" \
		| awk '{print $2}')
	echo $res
}

_deployment_selector()
{
	res=$(cat ${KUBECTL_FZF_CACHE}/deployments \
		| column -t \
		| fzf -m --header="Deployment" --layout reverse -q "$2" \
		| awk '{print $1}')
	echo $res
}

_node_selector()
{
	res=$(awk '{print $5 }' ${KUBECTL_FZF_CACHE}/pods \
		| column -t \
		| fzf -m --header="Node" --layout reverse -q "$2" \
		| awk '{print $1}')
	echo $res
}

_flag_selector()
{
	res=$(awk '{print $3 }' ${KUBECTL_FZF_CACHE}/pods \
		| paste -sd ',' \
		| tr ',' '\n' \
		| grep -v None \
		| sort \
		| uniq \
		| column -t -s'=' \
		| fzf -m --header="Label Value" --layout reverse -q "$2" \
		| awk '{print $1 "=" $2}')
	echo $res
}

__kubectl_parse_get()
{
	local penultimate=$(echo $COMP_LINE | awk '{print $(NF-1)}')
	local last_part=$(echo $COMP_LINE | awk '{print $(NF)}')

	if [[ $penultimate == "--selector" || $penultimate == "-l" || $last_part == "--selector" || $last_part == "-l" ]]; then
		local query=""
		if [[ $penultimate == "--selector" || $penultimate == "-l" ]]; then
			query=$last_part
		fi
		flags=$(_flag_selector $1 $query)
		COMPREPLY=( "$flags" )
	elif [[ $1 == "pod" || $1 == "logs" ]]; then
		pods=$(_pod_selector $1 $2)
		COMPREPLY=( "$pods" )
	elif [[ $1 == "node" ]]; then
		nodes=$(_node_selector $1 $2)
		COMPREPLY=( "$nodes" )
	elif [[ $1 == "deployment" ]]; then
		nodes=$(_deployment_selector $1 $2)
		COMPREPLY=( "$nodes" )
	else
		___kubectl_parse_get $*
	fi
}
