#!/bin/bash

set -o errtrace
set -o errexit

docker run --rm -v $(pwd)/frontend:/data digitallyseamless/nodejs-bower-grunt:5 bower install
docker run --rm -w /go/src/github.com/Dataman-Cloud/crane -v $(pwd):/go/src/github.com/Dataman-Cloud/crane golang:1.5.4 make

if [ ! -f docker/docker ]; then
    curl https://get.docker.com/builds/Linux/x86_64/docker-latest.tgz | tar xzv
fi

export TAG=`git log --pretty=format:'%h' -n 1`
export CRANE_SWARM_MANAGER_IP=$CRANE_IP
export REGISTRY_PREFIX=""

docker-compose -p crane -f deploy/docker-compose.yml stop
docker-compose -p crane -f deploy/docker-compose.yml rm -f
docker-compose -p crane -f deploy/docker-compose.yml up -d
