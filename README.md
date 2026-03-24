[![Go](https://github.com/torloejborg/easykube/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/torloejborg/easykube/actions/workflows/go.yml)
[![CodeQL Advanced](https://github.com/torloejborg/easykube/actions/workflows/codeql.yml/badge.svg)](https://github.com/torloejborg/easykube/actions/workflows/codeql.yml)
[![Dependabot Updates](https://github.com/torloejborg/easykube/actions/workflows/dependabot/dependabot-updates/badge.svg?branch=main)](https://github.com/torloejborg/easykube/actions/workflows/dependabot/dependabot-updates)

# Easykube

"A tool for learning kubernetes, run various stacks, develop complex applications locally"

This README covers the basics, to find out more, visit the [documentation site](https://torloejborg.github.io/easykube/easykube/latest)

## What it does

Easykube wraps the awesome [Kind](https://kind.sigs.k8s.io/) and [Goja](https://github.com/dop251/goja) projects. It configures an opinionated default that uses a zot container registry as pull through cache for kubernetes.

With easykube and it's companion addon repository, you get dependency management for k8s apps and applications with preconfigured defaults. 

This is ideal for team based local development, no more fighting over that dev cluster.


## Download the binary for your platform
### [Linux/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-linux-amd64.zip)
### [Windows/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-windows-amd64.zip)
### [OSX/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-darwin-amd64.zip) *
### [OSX/arm64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-darwin-arm64.zip) *

\* osx binaries remain unsigned as I'm not paying for the privilege of building free software for apple.  

## Build from source
Go version 1.22.3 or newer should work.

Compile with ```go build```, go will pull in dependencies from github, and a binary ```easykube``` should appear in the project directory.

## Prerequisites

Next, you must have the following set of programs installed, and available on your path.

* docker (windows and mac, could use docker desktop. Docker on WSL2 works ok)
* kustomize
* kubectl
* helm

Use your favourite package manager to install the binaries. As long they are in your path easykube should pick them up. Do not use snap packages on Linux
for the prerequisites. Easykube will create a kind cluster called `easykube-kind`

Once all dependencies are in place, a little configuration is required.

![Tip](https://img.shields.io/badge/💡_Tip-green?style=for-the-badge)
> Easykube will not touch your existing kubeconfigs, it will
create a new configuration file in ~/.kube/easykube - Refer to [here](https://torloejborg.github.io/easykube/easykube/latest/install/#install-create) for more information.


## Get some addons
By itself, easykube is not very exciting, it can only establish an empty cluster. Clone this repository somewhere,

`git@github.com:torloejborg/easykube-addons.git`

## Configuration

1. Set an environment variable VISUAL=<an editor> this could be `nano`,`vi`,`code` or whatever you prefer.
3. Invoking `easykube config --use-defaults` will generate a default configuration

4. To inspect the configuration issue `easykube config --edit`

```
version: 1
easykube:
# location of addons dir
addon-root: /home/user/addons <-- change this! 

# where local configuration is stored
config-dir: /home/user/.config/easykube

# if an absolute path is not given, persistence will be located in config-dir
# this is used by kind to store persistent data, it will survive cluster deletion
persistence-dir: /home/user/.config/easykube/persistence

# Container Runtime docker or podman
container-runtime: docker

# use pull through caching on these container registries
mirror-registries:
  - registry-url: https://my-internal-registry.domain.name  <-- private registry
      username: <username>                                  <-- private registry 
      password: <password or access token>                  <-- private registry
  - registry-url: https://registry-1.docker.io
  - registry-url: https://ghcr.io
  - registry-url: https://quay.io
  - registry-url: https://registry.k8s.io
```

The important part being the path to the `addon-root` dir, change to match the location of an easykube addon repository.  

`easykube config` starts an interactive configuration session

5. Once a configuration is created `easykube boot` will bootstrap a kind cluster and a companion zot container registry. The registry is configured as a pull through cache for kubernetes. Configuring private registries are easy, just add username and password keys under registry-url




## Certificates
In the folder cacerts, there is a selfsigned CA certificate - Install this on your system to enjoy https browser connections (most addons require https)
