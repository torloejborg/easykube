let configuration = {
    "dependsOn": ["b","c"],
    "extraPorts": [
        {
            "hostPort" : 4000,
            "nodePort" : 10000
        }
    ],
    "extraMounts" : [
        {
            "hostPath" : "addon-a",
            "containerPath" : "/mount-point"
        }
    ]
}
