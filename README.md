# easykube 1.0.0

"A tool for learning kubernetes, running various stacks, and hacking your applications locally"

## Download the binary for your platform

### [OSX/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-darwin-amd64.zip)
### [OSX/arm64 (Silicon)](https:github.com/torloejborg/easykube/releases/latest/download/easykube-darwin-arm64.zip)
### [Linux/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-linux-amd64.zip)
### [Windows/amd64](https://github.com/torloejborg/easykube/releases/latest/download/easykube-windows-amd64.zip)

## Or, build from source
You must install go, version 1.22.3 or newer should work.

Compile with ```go build```, go will pull in dependencies from github, and a binary ```easykube``` should appear in the project directory.

## Prerequisite dependent binaries

Next, you must have the follow set of programs installed, and available on your path.

* docker (windows and mac, should use docker desktop)
* kustomize
* kubectl
* helm

Use your favourite package manager to install the binaries. As long they are in your path easykube should pick them up. Do not use snap packages on Linux
for the prerequisites. Easykube will create a kind cluster called `easykube-kind`

Once all dependencies are in place, a little configuration is required.

## Get some addons!
By itself, easykube is not very exciting, it can only establish an empty cluster and load kustomize and helm files/charts. Clone this repository somewhere,

`git@github.com:torloejborg/easykube-addons.git`

## Configuration

1. Set an environment variable VISUAL=<an editor> this could be `nano`,`vi`,`code` or whatever you prefer.
2. (Optional) Link the `easykube` binary to a place where the system can find it, such as /usr/local/bin, or add the easykube source tree to your PATH variable.
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

6. `(unset KUBECONFIG && easykube create -s <your local properties file>)` will create a kind cluster and import your *your.properties* as a secret which easykube will use to pull images from a private registry. It will also create a new kind-easykube cluster config.
 NOTE: If you are not using a private repository, the "-s" argument can be skipped, and images will be pulled from dockerhub  HINT: Create an alias, like "ek", "easykube" gets tedious to type :)
`

## What it does

It wraps Kind, and configures an opinionated default that works with a local docker registry.
It provides a method of orchestrating installation of applications that has some form
of dependency to other applications.

The addons directory is scanned, and a dependency-graph is created by
introspecting each *.ek.js file it locates. 

The javascript files are then executed in the correct order, carrying out the instructions in 
each file, such as fetching images, pushing to the local docker registry, and invoking Kustomize
to build and apply the manifests in each addon.

A simple set of command allows the user to perform rudimentary scripting
of the installation process. 

## Creating a new addon

A template function is provided that enables you to start out with a minimal example. run `easykube skaffold --name new-project --location utils`. This will create directory `utils/new-project` in the root of the addons dir.

`easykube list` will display the new addon as uninstalled (installed have green checkmarks)

`easykube add new-project` installs the addon, after some time, you can visit http://new-project.localtest.me 

When changes are made to the new addon, you can iterate changes quickly by force-installing without touching any of the dependencies, by issuing `easykube add new-project --force --no-depends` or `easykube add new-project -fn` 

## DNS Trickery

When the cluster is created, Kubernetes CoreDNS is patched to 
understand the localtest.me domain. "localtest.me" is a domain someone created on the internet, itself, and all subdomains resolves to 127.0.0.1

This allows us to invent an arbitrary number of hostnames which is useful when you have many services. Editing (and forgetting entries) in the hosts file is a thing of the past.   

A note for advanced users; Let's say you have a service in the cluster that depends on the "outside" hostname of a service already running in the cluster.

The prime example of this is the Keycloak addon.

Keycloak is exposed at https://keycloak.localtest.me. When developing a service that depends on keycloak, the service can only look up keycloak from within the cluster on the address using a service name keycloak.default.cluster.svc.

In order to circumvent this a rewrite rule modifies an internal the DNS request for keycloak.localtest.me to point to a service postfixed with **"-ext"**

```
 rewrite stop {
       name regex (.*)\.localtest.me {1}-ext.default.svc.cluster.local
       answer auto
    }
```

So to enable "external" name resolution for a given service, in this case keycloak. Just create an extra 
service called keycloak-ext (which points to the pod/deployment). The service must be created in the default namespace.  

In your application the url to the oauth provider will simply be keycloak.localtest.me. Your app and keycloak will not complain about different origins.  

## Certificates
In the folder cacerts, there is a selfsigned CA certificate - Install this on your system to enjoy https browser connections (most addons require https)
