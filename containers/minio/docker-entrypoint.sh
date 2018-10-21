#!/usr/bin/env sh

NO_REPLICAS=${REPLICAS:-4}
HOST=${SERVICE:-"minio"}
VOLUME_PATH=${VOLUME:-"export"}

TASKS="tasks"
QUERY="${TASKS}.${HOST}"

until nslookup "${HOST}"
do
  echo "waiting for service discovery..."
done

NO_HOSTS=0
while [[ "${NO_HOSTS}" -lt "${NO_REPLICAS}" ]] 
do
  echo "waiting for all replicas to come online..."
  NO_HOSTS=$(nslookup "${QUERY}" | grep Address | wc -l)
  echo $NO_HOSTS
done

HOSTNAMES=$(nslookup "${QUERY}" | grep "Address" | awk '{ print $3 }' | sed -e 's/^/http:\/\//' | sed -e "s/$/\/${VOLUME_PATH}/" | tr '\n' ' ' | sed -e 's/[ \t]*$//')

# export secrets
export MINIO_ACCESS_KEY=$(cat /run/secrets/MINIO_ACCESS_KEY)
export MINIO_SECRET_KEY=$(cat /run/secrets/MINIO_SECRET_KEY)

# start server
eval "minio server" "${HOSTNAMES}"
