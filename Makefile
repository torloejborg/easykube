VERSION     ?= $(shell git describe --tags --dirty --always)
MODULE_PATH := $(shell grep '^module ' go.mod | cut -d' ' -f2)
LDFLAGS     = -ldflags "-X $(MODULE_PATH)/pkg/vars.Version=$(VERSION)"

linux:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o build/easykube-linux-amd64

osx_amd64:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o build/easykube-darwin-amd64

osx_arm64:
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o build/easykube-darwin-arm64

windows:
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o build/easykube-windows-amd64.exe

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
	mockgen -typed --source pkg/ez/addon_reader.go --destination mock/m_addon_reader.go
	mockgen -typed --source pkg/ez/container_runtime.go --destination mock/m_container_runtime.go
	mockgen -typed --source pkg/ez/cobra_command_helper.go --destination mock/m_cobra_command_helper.go
	mockgen -typed --source pkg/ez/config_utils.go --destination mock/m_config_utils.go
	mockgen -typed --source pkg/ez/cluster_utils.go --destination mock/m_cluster_utils.go
	mockgen -typed --source pkg/ez/external_tools.go --destination mock/m_external_m.go
	mockgen -typed --source pkg/ez/k8s_utils.go --destination mock/m_k8s_utils.go
	mockgen -typed --source pkg/ez/os_details.go --destination mock/m_os_details.go
	mockgen -typed --source pkg/ez/addon_types.go --destination mock/m_addon.go
	mockgen -typed --source pkg/js/jsrunner.go --package mock_ez --destination mock/m_jsrunner.go


.PHONY:docs
docs:
	asciidoc/ensure-spring-extensions.sh
	antora generate antora-playbook.yml
	touch docs/.nojekyll
