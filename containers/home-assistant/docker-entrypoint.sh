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
    "zones.yaml"
    "trackers.yaml"
    "known_devices.yaml"
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
    fill_from_env "${CONFIG_TEMPLATE_PATH}/${file}"
done

mv ${CONFIG_TEMPLATE_PATH}/* ${CONFIG_PATH}/

python -m homeassistant --config "${CONFIG_PATH}"
