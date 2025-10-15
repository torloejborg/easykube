/**
 * Type-hintable wrapper for the exposed Easykube Go functions
 */
class Easykube {

    /**
     * Defines semver constraint for the easykube go-binary, patch versions are free to roll.
     * The go binary will extract this, and compare it to its own version, if there is a version
     * mismatch, the status command will complain.
     */
    EASYKUBE_BINARY_COMPAT="~1.1.6";

    /**
     * Creates a new Easykube JavaScript handler. This class is for your IDE's enjoyment.
     * It wraps exported Go functions to provide type hints and documentation.
     */
    constructor() {
    }

    /**
     * Runs the kustomize tool, saves the rendered manifests into a .out file, and applies it with kubectl
     * @returns {Easykube}
     */
    kustomize() {
        _ek.kustomize();
        return this;
    }

    /**
     * Renders a Helm chart to .out, remember to apply the result
     * @param {string} chart relative path to the chart directory
     * @param {string} values the values file to use
     * @param {string} destination output file for rendered template
     * @param {string} namespace sets namespace in templated output
     */
    helmTemplate(chart, values, destination, namespace = 'default') {
        _ek.helmTemplate(chart, values, destination, namespace);
        return this;
    }

    /**
     * Waits until the specified deployment reaches ready state in the cluster.
     * @param {string} deployment - Deployment to wait for
     * @param {string} namespace - Where the deployment lives
     * @returns {Easykube}
     */
    waitForDeployment(deployment, namespace) {
        _ek.waitForDeployment(deployment, namespace);
        return this;
    }

    /**
     * Applies a single yaml resource with kubectl
     * @param {string} manifest - The yaml file to apply the path must be relative to the addon script
     * @returns {Easykube}
     */
    thenApply(manifest) {
        _ek.thenApply(manifest);
        return this;
    }

    /**
     * Preloads images by pulling them retagging and pushing them to your local registry,
     * this saves bandwidth, and will make subsequent installations faster
     * @param {Map<string, string>} images
     * @returns {Easykube}
     */
    preload(images) {
        _ek.preload(images);
        return this;
    }

    /**
     * Runs a command in a container     * @returns {Postgres}

     * @param {string} deployment - Name of the deployment (If A deployment has more than one container, the first discovered becomes the target)
     * @param {string} namespace , args) {
     return _ek.execInContainer(deployment, namespace, - Namespace of container
     * @param {string} command - The command to run, example "ls" or "/usr/local/bin/whatever"
     * @param {string[]} args - Arguments to the command, example ["-la","-v"]
     * @returns {string}
     */
    runCommand(deployment, namespace, command, args) {
        return _ek.execInContainer(deployment, namespace, command, args);
    }

    /**
     * Waits until a custom resource definition appears (Operators exemplifies this issue, it can
     * take some time for an operator to apply its own resources)
     * @param {string} group group name of crd 'acme.io'
     * @param {string} version V1
     * @param {string} kind the kind of resource, 'Widget'
     * @param {number} timeout duration in seconds to wait before giving up
     * @returns {Easykube}
     */
    waitForCustomResourceDefinition(group, version, kind, timeout) {
        _ek.waitForCRD(group, version, kind, timeout);
        return this;
    }

    /**
     * Copies a local file to a container in a pod. This is done by best effort. A pod can have many containers. If
     * two (or more) containers exist which matches the partial container name, the first becomes the target.
     *
     * Limitation; The destination MUST be aimed at a deployment, you cannot target a bare container. This is
     * why you must specify a deployment name
     *
     * @param {string} deployment Name of deployment
     * @param {string} namespace Where the target deployment lives
     * @param {string} containerLike The partial name of the target container
     * @param {string} localPath Source file, relative to the *.ek.js module definition
     * @param {string} remotePath Destination file in container, destination must be writeable
     * @returns {Easykube}
     */
    copyFile(deployment, namespace, containerLike, localPath, remotePath) {
        _ek.copyTo(deployment, namespace, containerLike, localPath, remotePath)
        return this;
    }

    /**
     * Creates a secret in a given namespace
     * @param {string} namespace create the secret in this namespace
     * @param {string} name what to call the secret
     * @param {{[key:string]: string}} data A map, All key and values are strings
     */
    createSecret(namespace, name, data) {
        _ek.createSecret(namespace, name, data)
    }

    /**
     *
     * @param namespace
     * @param name
     */
    getSecret(namespace, name) {
        return _ek.getSecret(namespace, name)
    }

    /**
     * Preprocesses an addon's yaml output and the resources related to the ExternalSecretOperator.
     * Secrets will be created without ESO installed, but ESO's custom resource definitions needs to be
     * available.
     * @param {Map<string,Map<string,any>>} secretMap An arbitrary map which provides the actual keys for the secret
     * @param {string} namespace where to apply the secret
     * @param {string} manifest which output to process, by default, it's the output from kustomize (.out.yaml)
     */
    processExternalSecrets(secretMap, namespace, manifest = '.out.yaml') {
        _ek.processExternalSecrets(secretMap, namespace, manifest);
        return this;
    }

    /**
     * Gets a value from the passed kv easykube commandline argument (easykube add fooaddon --kv foo=bar, bing=baz)
     * @param {string} key in the map
     * @return {string} value or undefined if key does not exist
     */
    kv(key) {
        return _ek.keyValue(key);
    }

    /**
     * Fetch the value of an environment variable
     * @param {string} key
     * @return {string} value of env-var, undefined if non-existent
     */
    env(key) {
        return _ek.env(key)
    }


    /**
     * Represents an HTTP response object.
     * @typedef {Object} HttpResponse
     * @property {(cb: (body: string) => any) => HttpResponse} onSuccess callback to handle successes 2xx
     * @property {(cb: (error: string, httpstatuscode: int) => any) => HttpResponse} onFail hard errors, malformed urls, connection failures etc.
     */

    /**
     * Performs http requests using a very basic client
     * @param {string} url endpoint for request
     * @param {string} method the method, POST, GET, etc.
     * @param {string} header headers for the request
     * @param {Object} body arbitrary json structure or text
     * @returns {HttpResponse}
     */
    http(url,method,header,body) {
        return  _ek.http(url,method,header,body);
    }

    /**
     * Represents a result of exec, this is created and exposed from go
     * @typedef {Object} ExecResult
     * @property {(cb: (output: string) => any) => string} onSuccess
     * @property {(cb: (output: string) => any) => string} onFail
     */

    /**
     * Executes an arbitrary OS command, this calls in to go's os.Exec
     * @param {string} command the command
     * @param {string[]} args arguments for the command
     * @returns {ExecResult} the result
     */
    exec(command, args) {
        return _ek.exec(command, args);
    }

    /**
     * Gets the directory of the current addon
     * @param {string} name optional addon name, if specificed will return the path to another addon in the respository
     * @returns {string} absoultue path to the addon directory
     */
    addonDir(name) {
        return _ek.addonDir(name);
    }
}

const easykube = new Easykube();

// newline at end-of-file is needed
