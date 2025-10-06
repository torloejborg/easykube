let configuration = {}

easykube.kustomize()

easykube.exec("ls",["."])
    .onSuccess((s)=> {console.info(s)})
