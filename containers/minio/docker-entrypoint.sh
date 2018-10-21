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
while [ "${NO_HOSTS}" -ne "${NO_REPLICAS}" ]; do
  echo "waiting for all replicas to come online..."

  HOSTS=$(nslookup "${QUERY}" 2>/dev/null | grep Address | grep -v '#')
  echo "Found: \n${HOSTS}"

  NO_HOSTS=$(nslookup "${QUERY}" 2>/dev/null | grep Address | grep -v '#' | wc -l)

  sleep 5
done

HOSTNAMES=`nslookup "${QUERY}" 2>/dev/null \
     | grep "Address" \
     | grep -v "#" \
     | awk '{ print $2 }' \
     | sed -e 's/^/http:\/\//' \
     | sed -e "s/$/\/${VOLUME_PATH}/" \
     | tr '\n' ' ' \
     | sed -e 's/[ \t]*$//'
`

echo "found hosts: $HOSTNAMES"

# start server
eval "minio server" "${HOSTNAMES}"
