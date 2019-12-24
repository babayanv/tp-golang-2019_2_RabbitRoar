#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
WORK_DIR="$SCRIPT_DIR/.."

echo "Running stop_prod in $WORK_DIR."

pushd $WORK_DIR

docker-compose -f deployments/docker-compose-prod.yml -p svoyak_backend down

popd
