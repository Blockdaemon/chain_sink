OUT_DIR := ./out

build-linux:
	CGO_ENABLED=1 GOARCH=amd64 GOOS=linux go build -o $(OUT_DIR)/chain_sink_linux_amd64 ./cmd/chain_sink

build-darwin:
	CGO_ENABLED=1 GOARCH=arm64 GOOS=darwin go build -o $(OUT_DIR)/chain_sink_darwin_arm64 ./cmd/chain_sink

build-release-linux: build-linux
	zip -j $(OUT_DIR)/chain_sink_linux_amd64.zip $(OUT_DIR)/chain_sink_linux_amd64

build-release-darwin: build-darwin
	zip -j $(OUT_DIR)/chain_sink_darwin_arm64.zip $(OUT_DIR)/chain_sink_darwin_arm64
