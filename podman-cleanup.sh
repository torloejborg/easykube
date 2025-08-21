#!/bin/bash

podman stop kind-registry
podman rm kind-registry

podman stop kind-control-plane
podman rm kind-control-plane

podman rmi -a
podman volume prune 