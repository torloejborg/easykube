#!/bin/bash

docker stop easykube-registry
docker rm easykube-registry

docker stop easykube-control-plane
docker rm easykube-control-plane

docker rmi -a
docker volume prune