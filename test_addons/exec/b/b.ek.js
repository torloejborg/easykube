let configuration = {
    "dependsOn": ["a"]
}

console.info("B addon is being processed in JS")

git.sparseCheckout("git@github.com/torloejborg/easykube-addons","main",["foo","bar"],".maybe-fluxcd-stuff")