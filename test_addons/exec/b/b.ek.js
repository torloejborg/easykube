let configuration = {
    "dependsOn": ["a"],
    "extraPorts": [{
        "nodePort": 32000,
        "hostPort": 9999,
        "protocol": "TCP"
    }],
    "extraMounts": [{
        "hostPath": "addon-a-data",
        "containerPath": "/storage/addon-a"
    }]
}

console.info("B addon is being processed in JS")

git.sparseCheckout("git@github.com/torloejborg/easykube-addons","main",["foo","bar"],".maybe-fluxcd-stuff")