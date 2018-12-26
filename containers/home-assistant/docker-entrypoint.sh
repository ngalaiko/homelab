#!/usr/bin/env bash

CONFIG_FILES=(
    "automations.yaml"
    "configuration.yaml"
    "customize.yaml"
    "groups.yaml"
    "scripts.yaml"
    "secrets.yaml"
)

function fill_from_env() {
    file_name="$1"

    if [ ! -f ${file_name} ]; then
    	echo "File ${file_name} not found!"
		return
	fi

    for name_value in $(env); do
        name="${name_value%=*}"
        value="${name_value#*=}"

        sed -i -e "s#\${${name}}#${value}#g" ${file_name}
    done
}

for file in ${CONFIG_FILES[@]}; do
    fill_from_env "${file}"
done

python -m homeassistant --config /config
