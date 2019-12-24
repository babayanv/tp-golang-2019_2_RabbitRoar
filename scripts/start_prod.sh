#!/bin/bash

# Set work dir
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PROJ_DIR="$SCRIPT_DIR/.."

echo "Running start_prod $PROJ_DIR."

pushd $PROJ_DIR

docker-compose -f deployments/docker-compose-prod.yml pull
docker-compose -f deployments/docker-compose-prod.yml -p svoyak_backend up -d

popd
