linux:
	GOOS=linux GOARCH=amd64 go build -o build/easykube-linux-amd64
osx_amd64:
	GOOS=darwin GOARCH=amd64 go build -o build/easykube-darwin-amd64
osx_arm64:
	GOOS=darwin GOARCH=arm64 go build -o build/easykube-darwin-arm64
windows:
	GOOS=windows GOARCH=amd64 go build -o build/easykube-windows-amd64.exe
clean:
	rm -rf build
	rm -rf dist
	go clean
dist: linux windows osx_amd64 osx_arm64
	mkdir -p dist
	zip -jv dist/easykube-linux-amd64.zip build/easykube-linux-amd64
	zip -jv dist/easykube-windows-amd64.zip build/easykube-windows-amd64.exe
	zip -jv dist/easykube-darwin-amd64.zip build/easykube-darwin-amd64
	zip -jv dist/easykube-darwin-arm64.zip build/easykube-darwin-arm64
build:
	go build -o easykube-linux-amd64
