let configuration = {}

easykube.kustomize()

easykube.exec("cat",["/proc/cpuinfo"])
    .onSuccess((s)=> {console.info(s)})
