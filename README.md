[![Go](https://github.com/torloejborg/easykube/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/torloejborg/easykube/actions/workflows/go.yml)
[![CodeQL Advanced](https://github.com/torloejborg/easykube/actions/workflows/codeql.yml/badge.svg)](https://github.com/torloejborg/easykube/actions/workflows/codeql.yml)
[![Dependabot Updates](https://github.com/torloejborg/easykube/actions/workflows/dependabot/dependabot-updates/badge.svg?branch=main)](https://github.com/torloejborg/easykube/actions/workflows/dependabot/dependabot-updates)

# Easykube

"A tool for learning kubernetes, run various stacks, develop complex applications locally"


This README covers the basics, to find out more, visit the [documentation site](https://torlojeborg.github.io/easykune)

## Download the binary for your platform

### [OSX/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-darwin-amd64.zip)
### [OSX/arm64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-darwin-arm64.zip)
### [Linux/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-linux-amd64.zip)
### [Windows/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-windows-amd64.zip)

## Or, build from source
You must install go, version 1.22.3 or newer should work.

Compile with ```go build```, go will pull in dependencies from github, and a binary ```easykube``` should appear in the project directory.

## Prerequisite dependent binaries

Next, you must have the follow set of programs installed, and available on your path.

* docker (windows and mac, could use docker desktop. Docker on WSL2 works ok)
* kustomize
* kubectl
* helm

Use your favourite package manager to install the binaries. As long they are in your path easykube should pick them up. Do not use snap packages on Linux
for the prerequisites. Easykube will create a kind cluster called `easykube-kind`

Once all dependencies are in place, a little configuration is required.

## Get some addons
By itself, easykube is not very exciting, it can only establish an empty cluster. Clone this repository somewhere,

`git@github.com:torloejborg/easykube-addons.git`

## Configuration

1. Set an environment variable VISUAL=<an editor> this could be `nano`,`vi`,`code` or whatever you prefer.
2. (Optional) Link the `easykube` binary to a place where the system can find it, such as /usr/local/bin, or add the easykube source tree to your PATH variable. Establishing a dev environment is covered [here](https://torloejborg.github.io/easykube/easykube/latest/install/#install-nix)
3. Now, invoke `easykube config` this starts the editor with a default configuration
    ```
   easykube:
    # location of easykube-addons dir
    addon-root: /home/user/code/research/easykube-addons
    # where configuration is stored
    config-dir: /home/user/.config/easykube
    # if an absolute path is not given, persistence will be located in config-dir
    persistence-dir: /home/user/.config/easykube/persistence
   ```
    The important part being the path to the addons dir, change to match the location of an easykube addon repository.

4. Use it; `easykube --help` prints out a summary of all commands, `easykube <command> --help` prints the summary for that command. 

6. `easykube create -s <your local properties file>` will create a kind cluster and import your *your.properties* as a secret which easykube will use to pull images from a private registry. It will also create a new kind-easykube cluster config.
 NOTE: If you are not using a private repository, the "-s" argument can be skipped, and images will be pulled from dockerhub.

## What it does

It basically wraps the awesome [Kind](https://kind.sigs.k8s.io/) and [Goja](https://github.com/dop251/goja) projects. It configures an opinionated default that works with a local docker registry.
It provides a method of orchestrating installation of applications that has some form
of dependency to other applications.

The addons directory is scanned, and a dependency-graph is created by
introspecting each *.ek.js file it locates. 

The javascript files are then executed in the correct order, carrying out the instructions in 
each file, such as fetching images, pushing to the local docker registry, and invoking Kustomize
to build and apply the manifests in each addon.

A simple set of command allows the user to perform rudimentary scripting
of the installation process. 


## Certificates
In the folder cacerts, there is a selfsigned CA certificate - Install this on your system to enjoy https browser connections (most addons require https)
