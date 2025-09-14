let configuration = {
    "dependsOn": ["d"],
    "extraMounts" : [
        {
            "hostPath" : "addon-c",
            "containerPath" : "/mount-point"
        }
    ]
}
