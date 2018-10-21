#!/usr/bin/env sh

NO_REPLICAS=${REPLICAS:-4}
HOST=${SERVICE:-"minio"}
VOLUME_PATH=${VOLUME:-"export"}

TASKS="tasks"
QUERY="${TASKS}.${HOST}"

until nslookup "${HOST}"; do
  echo "waiting for service discovery..."
  sleep 5
done

NO_HOSTS=0
while [ "${NO_HOSTS}" -lt "${NO_REPLICAS}" ]; do
  echo "waiting for all replicas to come online..."
  nslookup "${QUERY}"

  NO_HOSTS=$(nslookup "${QUERY}" 2>/dev/null | grep Address | wc -l)
  echo $NO_HOSTS
  sleep 5
done

HOSTNAMES=$(nslookup "${QUERY}" 2>/dev/null \
     | grep "Address" \
     | awk '{ print $3 }' \
     | sed -e 's/^/http:\/\//' \
     | sed -e "s/$/\/${VOLUME_PATH}/" \
     | tr '\n' ' ' \
     | sed -e 's/[ \t]*$//')

echo "found hosts: $HOSTNAMES"

# start server
eval "minio server" "${HOSTNAMES}"
