// This is generated content, the scaffolding creates a rudimentary application, composed
// of the basic Kubernetes primitives; A Service, Configmap, Deployment and Ingress.

// This is a special variable, when easykube scans the addons, it will extract the
// configuration and add it to the overall cluster configuration. 'configuration' is a reserved word.
let configuration = {

        // Dependencies on other addons is declared here, for instance, if your application depends
        // on persistence via Postgres, easykube will install the dependency tree before app.
        "dependsOn" : ["ingress"]
         // Declare extra port mappings for your application here, you must have
         // a 'NodePort' service defined that exposes the required port.
         // "extraPorts" : [
         //   {
         //       "hostPort" : 9000,
         //       "containerPort" : 32525,
         //       "protocol": "TCP"
         //   }
         //],
         // You can also define custom storage locations that will mount a persistent-volume,
         // to a location on your local file system. If hostPath is relative it will be created in the
         // <UserConfigDir>/easykube/persistence, otherwise the absolute path is used.
         // "extraMounts" : [
         //   {
         //           "hostPath":"storage",
         //           "containerPath":"/var/openebs/local"
         //   }
         // ]
}

// You are free to define non-reserved variables that can be passed to other functions
let namespace = "default"
let deployment = "{{.DeploymentName}}"

// You can preload multiple images into the local registry, this will prevent Kind
// from fetching images from network, thus accelerating deployment times.
const images = {"nginx:latest":"localhost:5001/nginx:latest"}

// This invokes the 'kustomize' command, which assembles all your deployment yaml, and
// applies it Kind with 'kubectl'.
easykube
    .preload(images)
    .kustomize()
    .waitForDeployment(deployment,namespace)


