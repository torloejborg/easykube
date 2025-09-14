let configuration = {
    "dependsOn": ["d"],
    "extraMounts" : [
        {
            "hostPath" : "addon-b",
            "containerPath" : "/mount-point"
        }
    ]
}
