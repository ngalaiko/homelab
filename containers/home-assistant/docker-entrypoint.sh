#!/usr/bin/env bash

CONFIG_PATH="/config/config"
CONFIG_TEMPLATE_PATH="/config/template"

mkdir -p "${CONFIG_PATH}"

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
    file_distination="$2"

    if [ ! -f ${file_name} ]; then
    	echo "File ${file_name} not found!"
		return
	fi

    for name_value in $(env); do
        name="${name_value%=*}"
        value="${name_value#*=}"

        sed "s#\${${name}}#${value}#g" ${file_name} > ${file_distination}
    done
}

for file in ${CONFIG_FILES[@]}; do
    if [ -f "${CONFIG_PATH}/${file}" ]; then
        echo "skipping ${CONFIG_PATH}/${file}, already exists"
        continue
    fi

    fill_from_env "${CONFIG_TEMPLATE_PATH}/${file}" "${CONFIG_PATH}/${file}"
done

python -m homeassistant --config "${CONFIG_PATH}"
