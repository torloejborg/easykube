VERSION     ?= $(shell git describe --tags --dirty --always)
MODULE_PATH := $(shell grep '^module ' go.mod | cut -d' ' -f2)
LDFLAGS     = -ldflags "-X $(MODULE_PATH)/pkg/vars.Version=$(VERSION)"
TAGS		= ""

linux:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -tags $(TAGS) $(LDFLAGS) -o build/easykube-linux-amd64

osx_amd64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -tags $(TAGS) $(LDFLAGS) -o build/easykube-darwin-amd64

osx_arm64:
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -tags $(TAGS) $(LDFLAGS) -o build/easykube-darwin-arm64

windows:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -tags $(TAGS) $(LDFLAGS) -o build/easykube-windows-amd64.exe

clean:
	rm -rf build dist
	go clean

dist: linux windows osx_amd64 osx_arm64
	mkdir -p dist
	@command -v upx >/dev/null && upx -9 build/easykube-linux-amd64 || echo "UPX not installed, skipping compression"
	@command -v upx >/dev/null && upx -9 build/easykube-windows-amd64.exe || echo "UPX not installed, skipping compression"
	@command -v upx >/dev/null && upx -9 build/easykube-darwin-amd64 || echo "UPX not installed, skipping compression"
	@command -v upx >/dev/null && upx -9 build/easykube-darwin-arm64 || echo "UPX not installed, skipping compression"
	zip -jv dist/easykube-linux-amd64.zip build/easykube-linux-amd64
	zip -jv dist/easykube-windows-amd64.zip build/easykube-windows-amd64.exe
	zip -jv dist/easykube-darwin-amd64.zip build/easykube-darwin-amd64
	zip -jv dist/easykube-darwin-arm64.zip build/easykube-darwin-arm64

.PHONY:mock
mock:
	# core interfaces
	mockgen -typed --source pkg/core/addon.go --destination mock/m_addon.go
	mockgen -typed --source pkg/core/addonreader.go --destination mock/m_addon_reader.go
	mockgen -typed --source pkg/core/clusterutils.go --destination mock/m_cluster_utils.go
	mockgen -typed --source pkg/core/cobra_command_helper.go --destination mock/m_cobra_command_helper.go
	mockgen -typed --source pkg/core/config.go --destination mock/m_config_utils.go
	mockgen -typed --source pkg/core/container.go --destination mock/m_container_runtime.go
	mockgen -typed --source pkg/core/externaltools.go --destination mock/m_external_m.go
	mockgen -typed --source pkg/core/k8s.go --destination mock/m_k8s_utils.go
	mockgen -typed --source pkg/core/osdetails.go --destination mock/m_os_details.go
	mockgen -typed --source pkg/core/printer.go --destination mock/m_printer.go
	mockgen -typed --source pkg/core/skaffold.go  --destination mock/m_skaffold.go
	mockgen -typed --source pkg/core/status.go --destination mock/m_status.go
	mockgen -typed --source pkg/core/task_manager.go --destination mock/m_task_manager.go
	mockgen -typed --source pkg/core/utils.go --destination mock/m_utils.go

gomod2nix:
	nix run github:nix-community/gomod2nix -- generate
