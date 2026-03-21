#!/usr/bin/env bash

if [ ! -d node_modules ]; then
  npm init -y
  npm install @springio/antora-extensions
  npm install @asciidoctor/diagram
  npm install asciidoctor-kroki
  touch docs/.nojekyll



fi
