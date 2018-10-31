# Copyright 2018 Anthonin Bonnefoy
#
# This Source Code Form is subject to the terms of the Mozilla Public
# License, v. 2.0. If a copy of the MPL was not distributed with this
# file, You can obtain one at http://mozilla.org/MPL/2.0/.

export KUBECTL_FZF_CACHE="/tmp/kubectl_fzf_cache"
eval "`declare -f __kubectl_parse_get | sed '1s/.*/_&/'`"

_pod_selector()
{
	res=$(awk '{print $1 " " $2 " " $4 " " $5 " " $6 " " $7 }' ${KUBECTL_FZF_CACHE}/pods \
		| column -t \
		| sort \
		| fzf --sync -m --header="Namespace Name IP Node Status Age" --layout reverse -q "$2" \
		| awk '{print $2}')
	echo $res
}

_deployment_selector()
{
	res=$(cat ${KUBECTL_FZF_CACHE}/deployments \
		| column -t \
		| sort \
		| fzf -m --header="Deployment" --layout reverse -q "$2" \
		| awk '{print $1 " " $3}')
	echo $res
}

_service_selector()
{
	res=$(awk '{print $2 " " $3 " " $4 " " $5 " " $6}' ${KUBECTL_FZF_CACHE}/services \
		| column -t \
		| sort \
		| fzf -m --header="Service Type Ip Ports Selector" --layout reverse -q "$2" \
		| awk '{print $1}')
	echo $res
}

_node_selector()
{
	res=$(awk '{print $5 " " $7 }' ${KUBECTL_FZF_CACHE}/pods \
		| column -t \
		| grep -v None \
		| sort \
		| fzf -m --header="Node Age" --layout reverse -q "$2" \
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
		| fzf -m --header="Label Value" --layout reverse -q "$2" \
		| awk '{print $1}')
	echo $res
}

__kubectl_parse_get()
{
	local penultimate=$(echo $COMP_LINE | awk '{print $(NF-1)}')
	local last_part=$(echo $COMP_LINE | awk '{print $(NF)}')

	if [[ $penultimate == "--selector" || $penultimate == "-l" || $last_part == "--selector" || $last_part == "-l" ]]; then
		if [[ $penultimate == "--selector" || $penultimate == "-l" ]]; then
			query=$last_part
		fi
		flags=$(_flag_selector $1 $query)
		if [[ -n $flags ]]; then
			COMPREPLY=( "$flags" )
		fi
		return
	fi

	local query=""
	if [[ $1 != $last_part ]]; then
		query=$last_part
	fi

	if [[ $1 =~ pods? ]]; then
		results=$(_pod_selector $1 $query)
	elif [[ $1 =~ nodes? ]]; then
		results=$(_node_selector $1 $query)
	elif [[ $1 == deployment ]]; then
		results=$(_deployment_selector $1 $query)
	elif [[ $1 == svc || $1 == service ]]; then
		results=$(_service_selector $1 $query)
	else
		___kubectl_parse_get $*
	fi

	if [[ -n "$results" ]]; then
		COMPREPLY=( $results )
	fi
}
