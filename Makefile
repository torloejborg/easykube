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
	mockgen -typed --source pkg/core/interfaces.go --destination mock/m_interfaces.go

gomod2nix:
	nix run github:nix-community/gomod2nix -- generate

.PHONY:docs
docs:
	asciidoc/ensure-npm-dependencies.sh
	node asciidoc/generate-doc.js "Easykube API Reference" asciidoc/modules/ROOT/examples/1-easykube.js > asciidoc/modules/ROOT/pages/js_reference/easykube.adoc
	node asciidoc/generate-doc.js "Postgres API Reference" asciidoc/modules/ROOT/examples/2-postgres.js > asciidoc/modules/ROOT/pages/js_reference/postgres.adoc
	node asciidoc/generate-doc.js "Postgres API Reference" asciidoc/modules/ROOT/examples/3-utils.js > asciidoc/modules/ROOT/pages/js_reference/utils.adoc
	antora generate --stacktrace antora-playbook.yml
	touch docs/.nojekyll
