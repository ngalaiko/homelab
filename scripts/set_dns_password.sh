#!/usr/bin/env sh

if [ -z "$1" ]; then
	echo "please provide password"
	exit 2
fi

docker exec dns pihole -a -p "$1"
