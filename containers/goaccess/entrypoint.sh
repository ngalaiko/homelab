#!/usr/bin/env bash

tail -f -n +0 /srv/log/access.log \
    | grep --line-buffered 'Host-galayko-rocks' > /srv/log/access.filtered.log

/sbin/tini --  goaccess \
    --no-global-config \
    --config-file=/srv/data/goaccess.conf
