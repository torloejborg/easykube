#!/usr/bin/env bash
# This script requires the github commandline client https://cli.github.com/
# As argument, pass the desired version, gh will create a tag.

export VERSION=$1
make dist

gh release create $1 \
  dist/easykube-linux-amd64.zip \
  dist/easykube-windows-amd64.zip \
  dist/easykube-darwin-arm64.zip \
  dist/easykube-darwin-amd64.zip \

git pull