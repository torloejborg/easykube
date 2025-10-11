#!/usr/bin/env bash

if [ ! -d node_modules ]; then
  npm init -y
  npm install @springio/antora-extensions
  touch docs/.nojekyll
fi
